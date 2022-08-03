package main

import (
	"database/sql"
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/joinVc"
	"github.com/azuki-bar/ceviord/pkg/slashCmd"
	"github.com/azuki-bar/ceviord/pkg/speech/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/bwmarrin/discordgo"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/k0kubun/pp"
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
	log.Println(fmt.Errorf("parse env config `%w`", err))
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
	ceviord.SetParam(conf.param)

	dgSess, err := discordgo.New("Bot " + conf.auth.CeviordConn.Discord)
	if err != nil {
		log.Println("create discord go session failed `%w`", err)
		return
	}

	dgSess.AddHandler(func(s *discordgo.Session, _ *discordgo.Connect) { log.Println("connect to discord") })
	dgSess.AddHandler(ceviord.MessageCreate)
	dgSess.AddHandler(slashCmd.InteractionHandler)
	dgSess.AddHandler(joinVc.VoiceStateUpdateHandler)
	// dgSess.Debug = true
	gTalker, closer := grpc.NewTalker(&conf.auth.CeviordConn, &conf.param.Parameters[0])
	defer closer()
	ceviord.SetNewTalker(gTalker)

	var db *sql.DB
	for i := 1; i <= dbChallengeTimes; i++ {
		dbConf := conf.auth.CeviordConn.DB
		dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true", dbConf.User, dbConf.Password, dbConf.Protocol, dbConf.Addr, dbConf.Name)
		db, err = sql.Open("mysql", dsn)
		if err == nil && db.Ping() == nil {
			break
		}
		time.Sleep(dbTimeoutSecond)
	}
	if err != nil {
		log.Println(fmt.Errorf("db connection failed `%w`", err))
		return
	}
	defer db.Close()
	dialect := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	r, err := replace.NewReplacer(db, dialect)
	if err != nil {
		log.Println(fmt.Errorf("db set failed `%w`", err))
		return
	}
	ceviord.SetDbController(r)

	// Open the websocket and begin listening.
	err = dgSess.Open()
	defer dgSess.Close()
	if err != nil {
		log.Fatalln(fmt.Errorf("error opening Discord session: `%w`", err))
	}
	sg := slashCmd.NewSlashCmdGenerator()
	err = sg.AddCastOpt(conf.param.Parameters)
	if err != nil {
		log.Println("slash command generate failed")
	}
	pp.Print(sg.Generate())
	slashCmds, err := slashCmd.NewCmds(dgSess, "", sg.Generate())
	defer func() {
		if slashCmds != nil {
			slashCmds.DeleteCmds(dgSess, "")
		}
	}()
	if err != nil {
		log.Println(fmt.Errorf("slash command applier failed `%w`", err))
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return
}
