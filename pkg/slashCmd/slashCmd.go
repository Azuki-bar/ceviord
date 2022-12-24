package slashCmd

import (
	"context"
	"fmt"
	"time"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"go.uber.org/zap"

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
		ceviord.Cache.Logger.Warn("parse command failed", zap.Error(err))
		return
	}
	timeoutDuration := 2500 * time.Millisecond
	ctx, close := context.WithTimeout(context.Background(), timeoutDuration)
	defer close()
	// TODO; タイムアウト時に handle内でメッセージを送信しないように変更。
	go h.handle(ctx, s, i)
	select {
	case <-ctx.Done():
		return
	case <-time.After(timeoutDuration):
		replySimpleMsg(ceviord.Cache.Logger, "コネクションがタイムアウトしました。", s, i.Interaction)
		ceviord.Cache.Logger.Error("connection timeout", zap.Duration("time out limit", timeoutDuration))
	}
}

type CommandHandler interface {
	handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate)
}

func parseCommands(name string) (CommandHandler, error) {
	var h CommandHandler
	switch name {
	case joinCmdName:
		h = &join{logger: ceviord.Cache.Logger}
	case byeCmdName:
		h = &leave{}
	case helpCmdName:
		h = new(help)
	case "ping":
		h = &ping{logger: ceviord.Cache.Logger}
	case "dict":
		h = &dict{logger: ceviord.Cache.Logger}
	case "cast":
		h = &change{logger: ceviord.Cache.Logger}
	default:
		return nil, fmt.Errorf("command `%s` is not found", name)
	}
	return h, nil
}
