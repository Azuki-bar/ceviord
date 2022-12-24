package slashCmd

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type help struct {
	logger *zap.Logger
}

func (h *help) handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "コマンドリファレンス",
				Description: "コマンドはこのページを参考に入力してください。",
				URL:         "https://github.com/Azuki-bar/ceviord/blob/main/docs/cmd.md",
			},
			},
		},
	},
	)
	ctx.Done()
	if err != nil {
		h.logger.Warn("interaction respond failed", zap.Error(err))
	}
}
