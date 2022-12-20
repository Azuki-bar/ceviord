package slashCmd

import (
	"fmt"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type leave struct {
	logger *zap.Logger
}

func (l *leave) handle(finish chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := l.rawHandle(s, i)
	var msg string
	msg = "successfully leaved!"
	if err != nil {
		msg = fmt.Sprintln(fmt.Errorf("error in leave handler `%w`", err))
		l.logger.Warn("raw handle failed", zap.Error(err))
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	if err != nil {
		l.logger.Warn("error in leave interaction respond", zap.Error(err))
	}
	finish <- true
}
func (l *leave) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	cev, err := ceviord.Cache.Channels.GetChannel(i.GuildID)
	if err != nil || cev == nil {
		return fmt.Errorf("connection not found")
	}
	isJoin, err := cev.IsActorJoined(s)
	if !isJoin || cev.VoiceConn == nil {
		return fmt.Errorf("voice actor is already disconnected")
	}
	if err != nil {
		return err
	}
	defer func() {
		if cev.VoiceConn != nil {
			cev.VoiceConn.Close()
			err = ceviord.Cache.Channels.DeleteChannel(i.GuildID)
			if err != nil {
				l.logger.Warn("delete channel from cache failed", zap.Error(err), zap.String("guildID", i.GuildID))
			}
		}
	}()
	err = cev.VoiceConn.Speaking(false)
	if err != nil {
		return fmt.Errorf("speaking falsing: %w", err)
	}
	err = cev.VoiceConn.Disconnect()
	if err != nil {
		return fmt.Errorf("disconnecting: %w", err)
	}
	return nil
}
