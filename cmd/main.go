package main

import (
	"ceviord/pkg/ceviord"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
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

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
	return
}
