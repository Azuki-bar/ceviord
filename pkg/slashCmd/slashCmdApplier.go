package slashCmd

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type SlashCmds struct {
	appliedCmds []*discordgo.ApplicationCommand
}

func NewCmds(s *discordgo.Session, guildId string, cmds []*discordgo.ApplicationCommand) (*SlashCmds, error) {
	sc := SlashCmds{}
	for _, cmd := range cmds {
		ac, err := s.ApplicationCommandCreate(s.State.User.ID, guildId, cmd)
		if err != nil {
			log.Printf("cmd: %+v, err: %v", cmd, err)
			return nil, err
		}
		log.Printf("slash command appling... %#+v", ac)
		sc.appliedCmds = append(sc.appliedCmds, ac)
	}
	return &sc, nil
}

func (sc *SlashCmds) DeleteCmds(s *discordgo.Session, guildId string) error {
	for _, cmd := range sc.appliedCmds {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildId, cmd.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
