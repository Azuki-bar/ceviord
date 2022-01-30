package main

import (
	"ceviord/pkg/ceviord"
	"flag"
	"github.com/bwmarrin/discordgo"
	"log"
)

func main() {

	token := flag.String("t", "", "discord token")
	if *token == "" {
		return
	}

	// Create a new Discordgo session
	dg, err := discordgo.New(token)
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
	return
}
