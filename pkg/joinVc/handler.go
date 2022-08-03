package joinVc

import (
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/discord"
	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/bwmarrin/discordgo"
)

type handler struct {
	*discordgo.VoiceStateUpdate
	session        *discordgo.Session
	changeState    ChangeRoomState
	user           discord.User
	joinedChannels ceviord.Channels
}

func (h *handler) handle() error {
	switch h.changeState.(type) {
	case intoRoom, outRoom:
		return ceviord.RawSpeak(h.changeState.GetText(), h.VoiceStateUpdate.GuildID, h.session)
	default:
		return nil
	}
}

type Handler interface {
	handle() error
}

type ChangeRoomState interface {
	GetText() string
}

func NewChangeRoomState(vsu *discordgo.VoiceStateUpdate, s *discordgo.Session, cs *ceviord.Channels) ChangeRoomState {
	if !cs.IsExistChannel(vsu.GuildID) {
		return outOfScope{}
	}
	c, err := cs.GetChannel(vsu.GuildID)
	if err != nil {
		return outOfScope{}
	}
	u, err := discord.NewUser(vsu.UserID, s, vsu.GuildID)
	if u.Bot {
		return outOfScope{}
	}
	if err != nil {
		ceviord.Logger.Log(logging.WARN, err)
		return outOfScope{}
	}
	scn, err := u.ScreenName()
	if err != nil {
		ceviord.Logger.Log(logging.WARN, err)
		return outOfScope{}
	}
	var state ChangeRoomState = outOfScope{}
	if (vsu.BeforeUpdate == nil || vsu.BeforeUpdate.ChannelID != c.VoiceConn.ChannelID) && vsu.VoiceState != nil {
		state = intoRoom{screenName: scn}
	} else if vsu.BeforeUpdate != nil && vsu.VoiceState.ChannelID != c.VoiceConn.ChannelID {
		state = outRoom{screenName: scn}
	}
	switch state.(type) {
	case intoRoom:
		if vsu.ChannelID == c.VoiceConn.ChannelID {
			return state
		}
	case outRoom:
		if vsu.BeforeUpdate.ChannelID == c.VoiceConn.ChannelID {
			return state
		}
	default:
		return outOfScope{}
	}
	return outOfScope{}
}

type intoRoom struct{ screenName string }

func (r intoRoom) GetText() string {
	return fmt.Sprintf("%sさんが入室しました。", r.screenName)
}

type outRoom struct{ screenName string }

func (r outRoom) GetText() string {
	return fmt.Sprintf("%sさんが退室しました。", r.screenName)
}

type outOfScope struct{ ChangeRoomState }

func NewHandler(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) (Handler, error) {
	u, err := discord.NewUser(vsu.UserID, s, vsu.GuildID)
	if err != nil {
		return nil, err
	}
	cs := NewChangeRoomState(vsu, s, &ceviord.Cache.Channels)
	if cs == nil {
		// ignore not covered event
		return nil, nil
	}
	return &handler{
		session:          s,
		VoiceStateUpdate: vsu,
		user:             u,
		changeState:      cs,
		joinedChannels:   ceviord.Cache.Channels,
	}, nil
}

func VoiceStateUpdateHandler(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	h, err := NewHandler(s, vsu)
	if err != nil {
		ceviord.Logger.Log(logging.WARN, err)
		return
	}
	err = h.handle()
	if err != nil {
		ceviord.Logger.Log(logging.WARN, err)
		return
	}
	return
}