package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/azuki-bar/ceviord/pkg/speechGrpc"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

func parseToken(flag, envName string) (string, error) {
	env := os.Getenv(envName)
	if flag == "" && env == "" {
		return "", fmt.Errorf("token is not provided")
	} else if env != "" {
		return env, nil
	}
	return flag, nil
}
func main() {
	conffile, err := ioutil.ReadFile("./parameter.yaml")
	if err != nil {
		panic(any(fmt.Errorf("load config file failed `%w`", err)))
	}
	var conf ceviord.Config
	yaml.Unmarshal(conffile, &conf)
	fmt.Println(conf)
	ceviord.SetConf(&conf)

	disTokFlag := flag.String("t", "", "discord token")
	cevioTokFlag := flag.String("c", "", "cevio token")
	flag.Parse()

	disT, err := parseToken(*disTokFlag, "CEVIORD_DISCORD_TOKEN")
	if err != nil {
		if conf.Conn.Discord == "" {
			log.Fatalln("discord token is not provided")
		}
	} else {
		conf.Conn.Discord = disT
	}
	cevioTok, err := parseToken(*cevioTokFlag, "CEVIORD_CEVIO_TOKEN")
	if err != nil {
		if conf.Conn.Cevio == "" {
			log.Fatalln("cevio token is not provided")
		}
	} else {
		conf.Conn.Cevio = cevioTok
	}

	// Create a new Discordgo session
	dgSess, err := discordgo.New("Bot " + conf.Conn.Discord)
	if err != nil {
		log.Println("create discord go session failed `%w`", err)
		return
	}

	// Create a new Application
	ap := &discordgo.Application{}
	ap.Name = "ceviord"
	ap.Description = "read text with cevigo"
	ap, err = dgSess.ApplicationCreate(ap)
	dgSess.AddHandler(ceviord.MessageCreate)
	//ceviord.SetNewTalker(speechApi.NewTalker(&conf.Parameters[0]))
	gTalker, closer := speechGrpc.NewTalker(&conf.Conn, &conf.Parameters[0])
	defer closer()
	ceviord.SetNewTalker(gTalker)

	db, err := sql.Open(conf.Conn.DriverName, conf.Conn.Dsn)
	if err != nil {
		log.Println(fmt.Errorf("db connection failed `%w`", err))
		return
	}
	defer db.Close()
	r, err := replace.NewReplacer(db)
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
