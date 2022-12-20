package slashCmd

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type ping struct {
	logger *zap.Logger
}

func (p *ping) handle(c chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "your message have been trapped on ceviord server"},
	})
	c <- true
	if err != nil {
		p.logger.Warn("ping handler failed", zap.Error(err))
	}
}
