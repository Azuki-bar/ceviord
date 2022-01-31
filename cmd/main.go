package main

import (
	"ceviord/pkg/ceviord"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

	// Create a new Application
	ap := &discordgo.Application{}
	ap.Name = "ceviord"
	ap.Description = "read text with cevigo"
	ap, err = dg.ApplicationCreate(ap)
	dg.AddHandler(ceviord.MessageCreate)
	ceviord.SetNewTalker(ceviord.NewTalker())

	db, err := gorm.Open(sqlite.Open(filepath.Join("./", "db.sqlite3")))
	if err != nil {
		ceviord.SetDb(db)
	}
	defer ceviord.CloseDb()
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
