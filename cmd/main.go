package main

import (
	"ceviord/pkg/ceviord"
	"ceviord/pkg/replace"
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
	token := flag.String("t", "", "discord token")
	flag.Parse()
	if *token == "" {
		return
	}

	// Create a new Discordgo session
	dg, err := discordgo.New("Bot " + *token)
	if err != nil {
		log.Println(err)
		return
	}

	conffile, err := ioutil.ReadFile("./parameter.yaml")
	if err != nil {
		panic(err)
	}
	var conf ceviord.Config
	yaml.Unmarshal(conffile, &conf)
	fmt.Println(conf)
	ceviord.SetParameters(&conf)
	// Create a new Application
	ap := &discordgo.Application{}
	ap.Name = "ceviord"
	ap.Description = "read text with cevigo"
	ap, err = dg.ApplicationCreate(ap)
	dg.AddHandler(ceviord.MessageCreate)
	ceviord.SetNewTalker(ceviord.NewTalker())

	//db, err := gorm.Open(sqlite.Open(filepath.Join("./", "dictionaries.sqlite3")))
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
	if err != nil {
		log.Println(fmt.Errorf("error opening Discord session: `%w`", err))
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
	return
}
