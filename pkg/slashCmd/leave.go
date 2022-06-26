package slashCmd

import (
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/bwmarrin/discordgo"
)

type leave struct{}

func (l *leave) handle(finish chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := l.rawHandle(s, i)
	var msg string
	msg = "successfully leaved!"
	if err != nil {
		msg = fmt.Sprintln(fmt.Errorf("error in leave handler `%w`", err))
		ceviord.Logger.Log(logging.WARN, fmt.Errorf("error in leave handler `%w`", err))
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	if err != nil {
		ceviord.Logger.Log(logging.WARN, fmt.Errorf("error in `leave` interaction respond, err is `%w`", err))
	}
	finish <- true
}
func (*leave) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	cev, err := ceviord.Cache.Channels.GetChannel(i.GuildID)
	if err != nil || cev == nil {
		return fmt.Errorf("connection not found")
	}
	isJoin, err := cev.IsActorJoined(s)
	if !isJoin || cev.VoiceConn == nil {
		return fmt.Errorf("voice actor is already disconnected")
	}
	defer func() {
		if cev.VoiceConn != nil {
			cev.VoiceConn.Close()
			ceviord.Cache.Channels.DeleteChannel(i.GuildID)
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
