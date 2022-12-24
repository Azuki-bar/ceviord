package slashCmd

import (
	"context"
	"fmt"
	"log"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type join struct {
	logger *zap.Logger
}

func (j *join) handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := j.rawHandle(ctx, s, i)
	var msg string
	msg = "successfully joined!"
	if err != nil {
		msg = fmt.Sprintln(fmt.Errorf("error in join handler `%w`", err))
		j.logger.Warn("join handler error", zap.Error(err))
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	ctx.Done()
	if err != nil {
		j.logger.Warn("interaction respond failed", zap.Error(err))
	}
}
func (*join) rawHandle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Member == nil {
		return fmt.Errorf("member field is nil. so cannot detect user status")
	}
	vc := ceviord.FindJoinedVC(s, i.GuildID, i.Member.User.ID)
	if vc == nil {
		return fmt.Errorf("voice connection not found")
	}
	if ceviord.Cache.Channels.IsExistChannel(i.Member.GuildID) {
		c, err := ceviord.Cache.Channels.GetChannel(i.Member.GuildID)
		if err != nil {
			return fmt.Errorf("some error occurred in user joined channel searcher")
		}
		isJoin, err := c.IsActorJoined(s)
		if err != nil || isJoin {
			return fmt.Errorf("sasara is already joined")
		}
	}
	voiceConn, err := s.ChannelVoiceJoin(i.GuildID, vc.ID, false, true)
	if err != nil {
		log.Println(fmt.Errorf("joining: %w", err))
		return err
	}
	ceviord.Cache.Channels.AddChannel(
		ceviord.Channel{PickedChannel: i.ChannelID, VoiceConn: voiceConn},
		i.GuildID,
	)
	return nil
}
