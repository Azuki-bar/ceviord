package slashCmd

import (
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/bwmarrin/discordgo"
)

type ping struct{}

func (*ping) handle(c chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "your message have been trapped on ceviord server"},
	})
	c <- true
	if err != nil {
		ceviord.Logger.Log(logging.WARN, fmt.Errorf("ping handler failed err is `%w`", err))
	}
}
