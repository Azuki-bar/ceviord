package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azuki-bar/ceviord/pkg/joinVc"
	"github.com/azuki-bar/ceviord/pkg/slashCmd"
	"github.com/azuki-bar/ceviord/pkg/speech/grpc"
	"go.uber.org/zap"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/bwmarrin/discordgo"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/vrischmann/envconfig"
	"gopkg.in/yaml.v2"
)

var (
	dbTimeoutSecond  = 2 * time.Second
	dbChallengeTimes = 3
)

type conf struct {
	param *ceviord.Param
	auth  *ceviord.Auth
}

func getConf() (*conf, error) {
	paramFile, err := os.ReadFile("./parameter.yaml")
	if err != nil {
		return nil, err
	}
	var param ceviord.Param
	if err = yaml.Unmarshal(paramFile, &param); err != nil {
		return nil, err
	}

	var auth ceviord.Auth
	err = envconfig.Init(&auth)
	if err == nil {
		return &conf{param: &param, auth: &auth}, nil
	}
	authFile, err := os.ReadFile("./auth.yaml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(authFile, &auth); err != nil {
		return nil, err
	}
	return &conf{param: &param, auth: &auth}, nil
}

func main() {
	conf, err := getConf()
	if err != nil {
		log.Fatalln("get config failed `%w`", err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("init zap failed", err)
	}
	ceviord.SetLogger(logger)
	ceviord.SetParam(conf.param)

	dgSess, err := discordgo.New("Bot " + conf.auth.CeviordConn.Discord)
	if err != nil {
		logger.Fatal("discord conn failed", zap.Error(err))
		return
	}

	dgSess.AddHandler(func(s *discordgo.Session, _ *discordgo.Connect) { logger.Info("discord connection established") })
	dgSess.AddHandler(ceviord.MessageCreate)
	dgSess.AddHandler(slashCmd.InteractionHandler)
	dgSess.AddHandler(joinVc.VoiceStateUpdateHandler)
	// dgSess.Debug = true
	gTalker, closer := grpc.NewTalker(&conf.auth.CeviordConn, &conf.param.Parameters[0])
	defer func() {
		err = closer()
		if err != nil {
			logger.Panic("grpc connection close failed", zap.Error(err))
		}
	}()
	ceviord.SetNewTalker(gTalker)

	var db *sql.DB
	for i := 1; i <= dbChallengeTimes; i++ {
		dbConf := conf.auth.CeviordConn.DB
		dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true", dbConf.User, dbConf.Password, dbConf.Protocol, dbConf.Addr, dbConf.Name)
		db, err = sql.Open("mysql", dsn)
		if err == nil && db.Ping() == nil {
			// connection established
			break
		}
		time.Sleep(dbTimeoutSecond)
	}
	if err != nil {
		logger.Fatal("db connection failed", zap.Error(err), zap.Int("db challenge time", dbChallengeTimes))
		return
	}
	defer db.Close()
	dialect := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	r, err := replace.NewReplacer(db, dialect)
	if err != nil {
		logger.Error("db set to replace failed", zap.Error(err))
		return
	}
	ceviord.SetDbController(r)

	// Open the websocket and begin listening.
	err = dgSess.Open()
	if err != nil {
		logger.Fatal("error opening Discord session", zap.Error(err))
	}
	defer dgSess.Close()
	sg := slashCmd.NewSlashCmdGenerator()
	err = sg.AddCastOpt(conf.param.Parameters)
	if err != nil {
		logger.Error("slash command generate failed", zap.Error(err), zap.Any("parameters", conf.param.Parameters))
	}
	slashCmds, err := slashCmd.NewCmds(dgSess, "", sg.Generate())
	defer func() {
		if slashCmds != nil {
			err = slashCmds.DeleteCmds(dgSess, "")
			if err != nil {
				logger.DPanic("slash command delete failed", zap.Error(err))
			}
		}
	}()
	if err != nil {
		logger.Fatal("slash command applier failed", zap.Error(err))
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
