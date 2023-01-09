package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/joinVc"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/azuki-bar/ceviord/pkg/slashCmd"
	"github.com/azuki-bar/ceviord/pkg/speech/grpc"

	"github.com/bwmarrin/discordgo"
	"github.com/go-gorp/gorp"
	"github.com/go-sql-driver/mysql"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var (
	dbTimeout        = 2 * time.Second
	dbChallengeTimes = 3
	Version          = "snapshot"
)

type conf struct {
	param *ceviord.Param
	auth  *ceviord.Auth
	log   *logConf
}

type logConf struct {
	Level *zapcore.Level `envconfig:"optional"`
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
	var logConf logConf
	err = func() error {
		if err := envconfig.Init(&auth); err != nil {
			return err
		}
		if err := envconfig.Init(&logConf); err != nil {
			return err
		}
		return nil
	}()
	if err == nil {
		return &conf{param: &param, auth: &auth, log: &logConf}, nil
	}
	log.Print("read config from env vars occurs error", err)
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
		log.Fatalf("get config failed err=`%s`", err)
	}
	logger := zap.Must(zap.NewDevelopment(zap.IncreaseLevel(conf.log.Level)))
	logger.Info("logger initialize successful!",
		zap.Stringer("logLevel", logger.Level()),
		zap.String("ceviord version", Version),
	)
	ceviord.SetLogger(logger)
	ceviord.SetParam(conf.param)

	dgSess, err := discordgo.New("Bot " + conf.auth.CeviordConn.Discord)
	if err != nil {
		logger.Fatal("discord conn failed", zap.Error(err))
		return
	}

	dgSess.AddHandler(func(_ *discordgo.Session, _ *discordgo.Connect) { logger.Info("discord connection established") })
	dgSess.AddHandler(slashCmd.InteractionHandler)
	dgSess.AddHandler(joinVc.VoiceStateUpdateHandler)
	dgSess.AddHandler(ceviord.MessageCreate)
	// dgSess.Debug = true
	gTalker, closer := grpc.NewTalker(logger, &conf.auth.CeviordConn, &conf.param.Parameters[0])
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
		conf := mysql.NewConfig()
		conf.User = dbConf.User
		conf.Passwd = dbConf.Password
		conf.Net = dbConf.Protocol
		conf.Addr = dbConf.Addr
		conf.DBName = dbConf.Name
		conf.ParseTime = true

		db, err = sql.Open("mysql", conf.FormatDSN())
		if err == nil && db.Ping() == nil {
			// connection established
			break
		}
		time.Sleep(dbTimeout)
	}
	if err != nil {
		logger.Fatal("db connection failed", zap.Error(err), zap.Int("db challenge time", dbChallengeTimes))
		return
	}
	defer db.Close()
	dialect := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	r, err := replace.NewReplacer(logger, db, dialect)
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
	logger.Info("discord session opened")
	sg := slashCmd.NewSlashCmdGenerator(logger)
	err = sg.AddCastOpt(conf.param.Parameters)
	if err != nil {
		logger.Error("slash command generate failed", zap.Error(err), zap.Any("parameters", conf.param.Parameters))
	}
	if _, err := slashCmd.NewCmds(dgSess, "", sg.Generate()); err != nil {
		logger.Fatal("slash command applier failed", zap.Error(err))
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	for k, v := range dgSess.VoiceConnections {
		k := k
		v := v
		go func() {
			v.Close()
			logger.Info("close vc connection", zap.String("guildID", k))
		}()
	}
}
