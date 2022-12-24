package slashCmd

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type ping struct {
	logger *zap.Logger
}

func (p *ping) handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "your message have been trapped on ceviord server"},
	})
	ctx.Done()
	if err != nil {
		p.logger.Warn("ping handler failed", zap.Error(err))
	}
}
