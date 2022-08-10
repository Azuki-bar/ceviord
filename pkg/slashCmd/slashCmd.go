package slashCmd

import (
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"time"

	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/bwmarrin/discordgo"
)

const (
	joinCmdName   = "join"
	byeCmdName    = "bye"
	helpCmdName   = "help"
	dictCmdName   = "dict"
	changeCmdName = "cast"
	pingCmdName   = "ping"
)

type Generator struct {
	cmds []*discordgo.ApplicationCommand
}

func NewSlashCmdGenerator() *Generator {
	s := Generator{cmds: slashCmdList}
	return &s
}
func (s *Generator) AddCastOpt(ps []ceviord.Parameter) error {
	var c *discordgo.ApplicationCommand
	for _, rawC := range s.cmds {
		if rawC.Name == changeCmdName {
			c = rawC
			break
		}
	}
	if c == nil {
		return fmt.Errorf("change cast command not found")
	}
	castOptPos := -1
	for i, o := range c.Options {
		if o.Name == "cast" {
			castOptPos = i
			break
		}
	}
	if castOptPos < 0 {
		return fmt.Errorf("cast option not found")
	}
	co := c.Options[castOptPos]
	for _, p := range ps {
		co.Choices = append(co.Choices,
			&discordgo.ApplicationCommandOptionChoice{
				Name:  p.Name,
				Value: p.Name,
			})
	}
	return nil
}

func (s *Generator) Generate() []*discordgo.ApplicationCommand {
	return s.cmds
}

func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h, err := parseCommands(i.ApplicationCommandData().Name)
	if err != nil {
		ceviord.Logger.Log(logging.INFO, fmt.Errorf("parse command failed err is `%w`", err))
		return
	}
	finish := make(chan bool)
	// TODO; タイムアウト時に handle内でメッセージを送信しないように変更。
	go h.handle(finish, s, i)
	select {
	case <-finish:
		return
	case <-time.After(2500 * time.Millisecond):
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "handler connection timeout"},
		})
	}
}

type CommandHandler interface {
	handle(finished chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate)
}

func parseCommands(name string) (CommandHandler, error) {
	var h CommandHandler
	switch name {
	case joinCmdName:
		h = new(join)
	case byeCmdName:
		h = new(leave)
	case helpCmdName:
		h = new(help)
	case "ping":
		h = new(ping)
	case "dict":
		h = new(dict)
	case "cast":
		h = new(change)
	default:
		return nil, fmt.Errorf("command `%s` is not found", name)
	}
	return h, nil
}
