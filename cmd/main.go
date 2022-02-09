package main

import (
	"ceviord/pkg/ceviord"
	"ceviord/pkg/replace"
	"ceviord/pkg/speechApi"
	"ceviord/pkg/speechGrpc"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

func main() {
	conffile, err := ioutil.ReadFile("./parameter.yaml")
	if err != nil {
		panic(err)
	}
	var conf ceviord.Config
	yaml.Unmarshal(conffile, &conf)
	fmt.Println(conf)
	ceviord.SetParameters(&conf)

	tok := flag.String("t", "", "discord token")
	flag.Parse()
	if *tok == "" && conf.Conn.Discord == "" {
		log.Fatalln("discord token is not provided")
	}
	if *tok != "" {
		conf.Conn.Discord = *tok
	}

	// Create a new Discordgo session
	dg, err := discordgo.New("Bot " + conf.Conn.Discord)
	if err != nil {
		log.Println("create discord go session failed `%w`", err)
		return
	}

	// Create a new Application
	ap := &discordgo.Application{}
	ap.Name = "ceviord"
	ap.Description = "read text with cevigo"
	ap, err = dg.ApplicationCreate(ap)
	dg.AddHandler(ceviord.MessageCreate)
	ceviord.SetNewTalker(speechApi.NewTalker(&conf.Parameters[0]))
	gTalker, closer := speechGrpc.NewTalker(&conf.Conn, &conf.Parameters[0])
	defer closer()
	ceviord.SetNewTalker(gTalker)

	db, err := sql.Open("sqlite3", filepath.Join("./", "dictionaries.sqlite3"))
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
	err = dg.Open()
	defer dg.Close()
	if err != nil {
		log.Println(fmt.Errorf("error opening Discord session: `%w`", err))
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return
}
