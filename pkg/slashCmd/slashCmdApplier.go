package slashCmd

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type SlashCmds struct {
	appliedCmds []*discordgo.ApplicationCommand
}

const connectionNum = 4

func ApplyCmds(logger *zap.Logger, s *discordgo.Session, guildID string, cmds []*discordgo.ApplicationCommand) (*SlashCmds, error) {
	sc := SlashCmds{}
	sc.appliedCmds = make([]*discordgo.ApplicationCommand, len(cmds))
	sem := make(chan struct{}, connectionNum)
	wg := &sync.WaitGroup{}
	for i, cmd := range cmds {
		wg.Add(1)
		cmd := cmd
		i := i
		go func() error {
			sem <- struct{}{}
			defer func() { <-sem }()
			logger.Debug("sem start")
			ac, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
			if err != nil {
				log.Printf("cmd: %+v, err: %v", cmd, err)
				return err
			}
			logger.Info("slash command apply successful!", zap.Any("command", ac))
			sc.appliedCmds[i] = ac
			wg.Done()
			return nil
		}()
	}
	wg.Wait()
	logger.Debug("finish applied")
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
