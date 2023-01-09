package slashCmd

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type SlashCmds struct {
	appliedCmds []*discordgo.ApplicationCommand
}

func ApplyCmds(logger *zap.Logger, s *discordgo.Session, guildID string, cmds []*discordgo.ApplicationCommand) (*SlashCmds, error) {
	sc := SlashCmds{}
	sc.appliedCmds = make([]*discordgo.ApplicationCommand, len(cmds))
	for i, cmd := range cmds {
		i := i
		cmd := cmd
		ac, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		logger.Debug("apply cmd start")
		if err != nil {
			logger.Error("apply command failed", zap.Error(err), zap.Any("command", cmd))
			return nil, err
		}
		logger.Debug("slash command apply successful!", zap.Any("command", ac))
		sc.appliedCmds[i] = ac
	}
	logger.Debug("all apply cmd finish")
	return &sc, nil
}

func (sc *SlashCmds) DeleteCmds(s *discordgo.Session, guildID string) error {
	for _, cmd := range sc.appliedCmds {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
