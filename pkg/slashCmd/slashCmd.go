package slashCmd

import (
	"context"
	"fmt"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/samber/lo"
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
	cmds   []*discordgo.ApplicationCommand
	logger *zap.Logger
}

func NewSlashCmdGenerator(logger *zap.Logger) *Generator {
	return &Generator{
		cmds:   slashCmdList,
		logger: logger,
	}
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
	co.Choices = lo.Map(ps, func(item ceviord.Parameter, index int) *discordgo.ApplicationCommandOptionChoice {
		return &discordgo.ApplicationCommandOptionChoice{
			Name:  item.Name,
			Value: item.Name,
		}
	})
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
	ctx := context.Background()
	go h.handle(ctx, s, i)
	<-ctx.Done()
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
		h = &leave{logger: ceviord.Cache.Logger}
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
