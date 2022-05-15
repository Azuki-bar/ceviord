package main

import (
	"database/sql"
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/azuki-bar/ceviord/pkg/speechGrpc"
	"github.com/go-gorp/gorp"
	"github.com/vrischmann/envconfig"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
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

	ap := &discordgo.Application{}
	ap.Name = "ceviord"
	ap.Description = "read text with cevigo"
	ap, err = dgSess.ApplicationCreate(ap)
	dgSess.AddHandler(ceviord.MessageCreate)
	gTalker, closer := speechGrpc.NewTalker(&conf.auth.CeviordConn, &conf.param.Parameters[0])
	defer closer()
	ceviord.SetNewTalker(gTalker)

	var db *sql.DB
	for i := 1; i <= dbChallengeTimes; i++ {
		db, err = sql.Open(conf.auth.CeviordConn.DriverName, conf.auth.CeviordConn.Dsn)
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
	var dialect gorp.Dialect
	if conf.auth.CeviordConn.DriverName == "mysql" {
		dialect = gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}
	} else if conf.auth.CeviordConn.DriverName == "sqlite3" {
		dialect = gorp.SqliteDialect{}
	}
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

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return
}
