package slashCmd

import (
	"context"
	"fmt"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type change struct {
	changeTo string
	logger   *zap.Logger
}

func (c *change) handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, o := range i.ApplicationCommandData().Options {
		if o.Name == "cast" {
			c.changeTo = o.StringValue()
		}
	}
	err := c.rawHandle(s, i)
	msg := fmt.Sprintf("successfully change cast to %s", c.changeTo)
	if err != nil {
		msg = err.Error()
		c.logger.Error("message handle failed in change", zap.Error(err))
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	ctx.Done()
	if err != nil {
		c.logger.Error("change handler failed", zap.Error(err))
	}
}
func (c *change) rawHandle(s *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	cev, err := ceviord.Cache.Channels.GetChannel(interaction.GuildID)
	if err != nil {
		return fmt.Errorf("voice connection not found")
	}
	isJoin, err := cev.IsActorJoined(s)
	if err != nil || !isJoin {
		return fmt.Errorf("voice connection not found")
	}
	for i, p := range ceviord.Cache.Param.Parameters {
		if c.changeTo == p.Name {
			cev.CurrentParam = &ceviord.Cache.Param.Parameters[i]
			if err := ceviord.RawSpeak(fmt.Sprintf("パラメータを %s に変更しました。", p.Name), interaction.GuildID, s); err != nil {
				return fmt.Errorf("speaking about parameter setting: `%w`", err)
			}
		}
	}
	return nil
}
