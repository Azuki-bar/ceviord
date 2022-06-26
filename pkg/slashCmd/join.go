package slashCmd

import (
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/bwmarrin/discordgo"
	"log"
)

type join struct{}

func (j *join) handle(c chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := j.rawHandle(s, i)
	var msg string
	msg = "successfully joined!"
	if err != nil {
		msg = fmt.Sprintln(fmt.Errorf("error in join handler `%w`", err))
		ceviord.Logger.Log(logging.WARN, fmt.Errorf("error in join handler"))
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	c <- true
	if err != nil {
		ceviord.Logger.Log(logging.WARN, fmt.Errorf("error in `join` interaction respond err is `%w`", err))
	}
}
func (*join) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
