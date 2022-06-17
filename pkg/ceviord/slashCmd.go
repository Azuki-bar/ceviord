package ceviord

import (
	"fmt"
	"log"

	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/bwmarrin/discordgo"
)

var Cmds = []*discordgo.ApplicationCommand{
	{
		Name:        "join",
		Description: "voice actor join",
	},
	{
		Name:        "bye",
		Description: "voice actor disconnect",
	},
	{
		Name:        "help",
		Description: "get command reference",
	},
	{
		Name:        "ping",
		Description: "check connection status",
	},
}

func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h, err := parseCommands(i.ApplicationCommandData().Name)
	if err != nil {
		logger.Log(logging.INFO, fmt.Errorf("parse command failed err is `%w`", err))
		return
	}
	h.handle(s, i)
}

type CommandHandler interface {
	handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}
type join struct{}
type leave struct{}
type help struct{}
type ping struct{}

func parseCommands(name string) (CommandHandler, error) {
	var h CommandHandler
	switch name {
	case "join":
		h = new(join)
	case "bye":
		h = new(leave)
	case "help":
		h = new(help)
	case "ping":
		h = new(ping)
	default:
		return nil, fmt.Errorf("command `%s` is not found", name)
	}
	return h, nil
}

func (j *join) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := j.rawHandle(s, i)
	var msg string
	msg = "successfully joined!"
	if err != nil {
		msg = fmt.Sprintln(fmt.Errorf("error in join handler `%w`", err))
		logger.Log(logging.WARN, fmt.Errorf("error in join handler"))
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("error in `join` interactoin respond err is `%w`", err))
	}
}
func (_ *join) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Member == nil {
		return fmt.Errorf("member field is nil. so cannot detect user status")
	}
	vc := FindJoinedVC(s, i.GuildID, i.Member.User.ID)
	if vc == nil {
		return fmt.Errorf("voice connection not found")
	}
	if ceviord.Channels.isExistChannel(i.Member.GuildID) {
		c, err := ceviord.Channels.getChannel(i.Member.GuildID)
		if err != nil {
			return fmt.Errorf("some error occurred in user joined channel searcher")
		}
		isJoin, err := c.isActorJoined(s)
		if err != nil || isJoin {
			return fmt.Errorf("sasara is already joined")
		}
	}
	voiceConn, err := s.ChannelVoiceJoin(i.GuildID, vc.ID, false, true)
	if err != nil {
		log.Println(fmt.Errorf("joining: %w", err))
		return err
	}
	ceviord.Channels.addChannel(
		Channel{pickedChannel: i.ChannelID, VoiceConn: voiceConn},
		i.GuildID,
	)
	return nil
}

func (l *leave) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := l.rawHandle(s, i)
	var msg string
	msg = "successfully leaved!"
	if err != nil {
		msg = fmt.Sprintln(fmt.Errorf("error in leave handler `%w`", err))
		logger.Log(logging.WARN, fmt.Errorf("error in leave handler `%w`", err))
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("error in `leave` interaction respond, err is `%w`", err))
	}
}
func (_ *leave) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	cev, err := ceviord.Channels.getChannel(i.GuildID)
	if err != nil || cev == nil {
		return fmt.Errorf("connection not found")
	}
	isJoin, err := cev.isActorJoined(s)
	if !isJoin || cev.VoiceConn == nil {
		return fmt.Errorf("voice actor is already disconnected")
	}
	defer func() {
		if cev.VoiceConn != nil {
			cev.VoiceConn.Close()
			ceviord.Channels.deleteChannel(i.GuildID)
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

func (_ *help) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "コマンドリファレンス",
				Description: "コマンドはこのページを参考に入力してください。",
				URL:         "https://github.com/Azuki-bar/ceviord/blob/main/doc/cmd.md",
			},
			},
		},
	},
	)
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("help handler failed err is `%w`", err))
	}
}
func (_ *ping) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "your message have been trapped on ceviord server"},
	})
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("ping handler failed err is `%w`", err))
	}
}
