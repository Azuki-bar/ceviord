package handleCmd

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Slashcmds struct {
	appliedCmds []*discordgo.ApplicationCommand
}

func NewCmds(s *discordgo.Session, guildId string, cmds []*discordgo.ApplicationCommand) (*Slashcmds, error) {
	sc := Slashcmds{}
	for _, cmd := range cmds {
		log.Printf("guild id is `%s`\n", guildId)
		appliedCmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildId, cmd)
		if err != nil {
			log.Printf("cmd: %+v, err: %v", cmd, err)
			return nil, err
		}
		sc.appliedCmds = append(sc.appliedCmds, appliedCmd)
	}
	return &sc, nil
}

func (sc *Slashcmds) DeleteCmds(s *discordgo.Session, guildId string) error {
	for _, cmd := range sc.appliedCmds {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildId, cmd.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
