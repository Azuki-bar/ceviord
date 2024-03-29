package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/joinVc"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/azuki-bar/ceviord/pkg/slashCmd"
	"github.com/azuki-bar/ceviord/pkg/speech/grpc"
	"github.com/samber/lo"

	"github.com/bwmarrin/discordgo"
	"github.com/go-gorp/gorp"
	"github.com/go-sql-driver/mysql"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var (
	// replaced when build this plogram
	Version = "snapshot"
)

type conf struct {
	param *ceviord.Param
	auth  *ceviord.Auth
	log   *logConf
}

type logConf struct {
	Level *myLogLevel `envconfig:"optional,default=info"`
}
type myLogLevel zapcore.Level

func (l *myLogLevel) Unmarshal(s string) error {
	zapLevel := zapcore.Level(*l)
	if err := zapLevel.UnmarshalText([]byte(s)); err != nil {
		return err
	}
	*l = myLogLevel(zapLevel)
	return nil
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
		if err := envconfig.InitWithOptions(&logConf, envconfig.Options{Prefix: `CEVIORD_LOG`}); err != nil {
			return err
		}
		return nil
	}()
	if err == nil {
		return &conf{param: &param, auth: &auth, log: &logConf}, nil
	}
	log.Printf("read config from env vars occurs error=`%s`", err)
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
	logger := zap.Must(zap.NewDevelopment(zap.IncreaseLevel(zapcore.Level(*conf.log.Level))))
	defer func() { _ = logger.Sync() }()
	logger.Info("logger initialize successful!",
		zap.Stringer("logLevel", logger.Level()),
		zap.String("version", Version),
	)
	// discordgo.Logger provides package wide logging consistency for discordgo
	// the format, a...  portion this command follows that of fmt.Printf
	//   msgL   : LogLevel of the message
	//   caller : 1 + the number of callers away from the message source
	//   format : Printf style message format
	//   a ...  : comma separated list of values to pass
	discordgo.Logger = func(msgL, caller int, format string, a ...interface{}) {
		switch msgL {
		case discordgo.LogDebug:
			logger.Debug("discordgo log", zap.String("msg from discordgo", fmt.Sprintf(format, a)))
		case discordgo.LogInformational:
			logger.Info("discordgo log", zap.String("msg from discordgo", fmt.Sprintf(format, a)))
		case discordgo.LogWarning:
			logger.Warn("discordgo log", zap.String("msg from discordgo", fmt.Sprintf(format, a)))
		case discordgo.LogError:
			logger.Error("discordgo log", zap.String("msg from discordgo", fmt.Sprintf(format, a)))
		default:
		}
	}
	ceviord.SetLogger(logger)
	ceviord.SetParam(conf.param)
	dgSess, err := discordgo.New("Bot " + conf.auth.CeviordConn.Discord)
	if err != nil {
		logger.Fatal("discord conn failed", zap.Error(err))
		return
	}
	closeFuncs := make([]func(), 0)
	addHandler := func(handlerFunc any) {
		f := dgSess.AddHandler(handlerFunc)
		closeFuncs = append(closeFuncs, f)
	}
	addHandler(func(_ *discordgo.Session, _ *discordgo.Connect) { logger.Info("discord connection established") })
	addHandler(slashCmd.InteractionHandler)
	addHandler(joinVc.VoiceStateUpdateHandler)
	addHandler(ceviord.MessageCreate)
	// dgSess.Debug = logger.Level() <= zap.DebugLevel
	gTalker, closer := grpc.NewTalker(logger, &conf.auth.CeviordConn, &conf.param.Parameters[0])
	defer func() {
		if err = closer(); err != nil {
			logger.Fatal("grpc connection close failed", zap.Error(err))
		}
	}()
	ceviord.SetNewTalker(gTalker)

	dbConf := func(c ceviord.DB) *mysql.Config {
		newConf := mysql.NewConfig()
		newConf.User = c.User
		newConf.Passwd = c.Password
		newConf.Net = c.Protocol
		newConf.Addr = c.Addr
		newConf.DBName = c.Name
		newConf.ParseTime = true
		return newConf
	}(conf.auth.CeviordConn.DB)
	db, err := sql.Open("mysql", dbConf.FormatDSN())
	if err != nil {
		logger.Fatal("db connection failed", zap.Error(err))
		return
	}
	if err == nil && db.Ping() == nil {
		logger.Info("db connection is estabilished")
	}
	defer db.Close()
	dialect := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	r, err := replace.NewReplacer(logger, db, dialect)
	if err != nil {
		logger.Error("db set to replace failed", zap.Error(err))
		return
	}
	ceviord.SetDbController(r)

	if err := dgSess.Open(); err != nil {
		logger.Fatal("error opening Discord session", zap.Error(err))
	}
	defer func() { dgSess.Close(); logger.Info("discord session closed") }()
	logger.Info("discord session opened")
	sg := slashCmd.NewSlashCmdGenerator(logger)
	err = sg.AddCastOpt(conf.param.Parameters)
	if err != nil {
		logger.Error("slash command generate failed", zap.Error(err), zap.Any("parameters", conf.param.Parameters))
	}
	if _, err := slashCmd.ApplyCmds(logger, dgSess, "", sg.Generate()); err != nil {
		logger.Fatal("slash command applier failed", zap.Error(err))
	}

	logger.Info("slash command apply all finished")
	gameStatus := fmt.Sprintf("version: %s", Version)
	if err := dgSess.UpdateGameStatus(0, gameStatus); err != nil {
		logger.Error("updateGameStatus returns err", zap.Error(err))
	}
	logger.Info("Update Game Status finish", zap.String("msg", gameStatus))

	logger.Info("wait for stop signal, Ctrl-C")
	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-sc
	logger.Info("handle signal", zap.Stringer("signal", sig))
	lo.ForEach(closeFuncs, func(item func(), _ int) { go item(); logger.Debug("handler deleting") })
	sem := make(chan struct{}, 4)
	wg := sync.WaitGroup{}
	for k, v := range dgSess.VoiceConnections {
		k := k
		v := v
		wg.Add(1)
		go func() {
			sem <- struct{}{}
			defer func() { <-sem; wg.Done() }()
			v.Close()
			if err := v.Disconnect(); err != nil {
				logger.Error("disconn err", zap.Error(err))
			}
			logger.Info("close vc connection", zap.String("guildID", k))
		}()
	}
	wg.Wait()
}
