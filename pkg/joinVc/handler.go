package joinVc

import (
	"fmt"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type Handler struct {
	*discordgo.VoiceStateUpdate
	session        *discordgo.Session
	changeState    ChangeRoomState
	user           discord.User
	joinedChannels ceviord.Channels
	logger         *zap.Logger
}

func (h *Handler) handle(speaker func(text string, guildId string, session *discordgo.Session) error, c *ceviord.Channel) error {
	switch h.changeState.(type) {
	case intoRoom, outRoom:
		msg, err := c.DictController.ApplyUserDict(h.changeState.GetText())
		if err != nil {
			return err
		}
		return speaker(msg, h.VoiceStateUpdate.GuildID, h.session)
	default:
		return nil
	}
}

type ChangeRoomState interface {
	GetText() string
}

func NewChangeRoomState(logger *zap.Logger, vsu *discordgo.VoiceStateUpdate, s *discordgo.Session, cs *ceviord.Channels) ChangeRoomState {
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
		logger.Warn("something", zap.Error(err))
		return outOfScope{}
	}
	scn, err := u.ScreenName()
	if err != nil {
		logger.Warn("screen name fetch failed", zap.Error(err), zap.Any("user", u))
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

func NewHandler(logger *zap.Logger, s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) (Handler, error) {
	u, err := discord.NewUser(vsu.UserID, s, vsu.GuildID)
	if err != nil {
		return Handler{}, err
	}
	cs := NewChangeRoomState(logger, vsu, s, &ceviord.Cache.Channels)
	if cs == nil {
		// ignore not covered event
		return Handler{}, nil
	}
	return Handler{
		session:          s,
		VoiceStateUpdate: vsu,
		user:             u,
		changeState:      cs,
		joinedChannels:   ceviord.Cache.Channels,
		logger:           logger,
	}, nil
}

func VoiceStateUpdateHandler(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	h, err := NewHandler(ceviord.Cache.Logger, s, vsu)
	if err != nil {
		h.logger.Error("voice state update handler", zap.Error(err))
		return
	}
	channel, err := ceviord.Cache.Channels.GetChannel(vsu.GuildID)
	if err != nil {
		h.logger.Info("get channel err, this will no connections to voice Channel", zap.Error(err), zap.String("voice status update guildID", vsu.GuildID))
		return
	}
	err = h.handle(ceviord.RawSpeak, channel)
	if err != nil {
		h.logger.Error("handler failed", zap.Error(err))
		return
	}
}
