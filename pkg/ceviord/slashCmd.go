package ceviord

import (
	"fmt"
	"log"

	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/bwmarrin/discordgo"
	"github.com/k0kubun/pp"
)

var Cmds = []*discordgo.ApplicationCommand{
	{
		Name:        "join",
		Description: "join voice actor",
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
	{
		Name:        "dict",
		Description: "manage dict records",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "add record",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "word",
						Description: "word",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "yomi",
						Description: "how to read that word",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "del",
				Description: "delete record",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "id",
						Description: "dictionary record id",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    true,
					},
				},
			},
			{
				Name:        "show",
				Description: "show records",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			/* {
				Name:        "search",
				Description: "search record with effect",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "search string",
						Description: "search",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			}, */
		},
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
type dict struct{}
type change struct{}

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
	case "dict":
		h = new(dict)
	// case "change":
	// 	h = new(change)
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

func (_ *dict) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cev, err := ceviord.Channels.getChannel(i.GuildID)
	cev.dictController.SetGuildId(i.GuildID)
	if err != nil {
		// voice channel connection not found
		replySimpleMsg(fmt.Sprintf("dict handler failed. err is `%s`", err.Error()), s, i.Interaction)
		return
	}
	subCmd, err := dictSubCmdParse(i.ApplicationCommandData().Options[0])
	if err != nil {
		replySimpleMsg(fmt.Sprintf("dict sub cmd parser failed. err is `%s`", err.Error()), s, i.Interaction)
		return
	}
	d, err := subCmd.execute(i.GuildID, i.Member.User.ID)
	if err != nil {
		pp.Print(err)
		replySimpleMsg(fmt.Sprintf("dict sub cmd handler failed. err is `%s`", err.Error()), s, i.Interaction)
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: d,
	})
}
func replySimpleMsg(msg string, s *discordgo.Session, i *discordgo.Interaction) {
	s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
}
func dictSubCmdParse(opt *discordgo.ApplicationCommandInteractionDataOption) (dictSubCmd, error) {
	if opt.Type != discordgo.ApplicationCommandOptionSubCommand {
		return nil, fmt.Errorf("option type failed")
	}
	switch opt.Name {
	case "add":
		return newDictAdd(opt.Options)
	case "del":
		return newDictDel(opt.Options)
	default:
		return nil, fmt.Errorf("dict sub command parse failed")
	}
}

type dictSubCmd interface {
	execute(guildId, authorId string) (*discordgo.InteractionResponseData, error)
}
type dictAdd struct {
	yomi string
	word string
}
type dictDel struct {
	id uint
}
type dictShow struct{}

func newDictAdd(opt []*discordgo.ApplicationCommandInteractionDataOption) (*dictAdd, error) {
	var da dictAdd
	for _, o := range opt {
		switch o.Name {
		case "yomi":
			da.yomi = o.StringValue()
		case "word":
			da.word = o.StringValue()
		default:
			return nil, fmt.Errorf("undefined option appear in dict add handler")
		}
	}
	return &da, nil
}
func (da *dictAdd) execute(guildId, authorId string) (*discordgo.InteractionResponseData, error) {
	if len(da.word) == 0 || len(da.yomi) == 0 {
		return nil, fmt.Errorf("dict add field are not satisfied")
	}
	if ceviord.Channels == nil {
		return nil, fmt.Errorf("channel connection not found")
	}
	cev, err := ceviord.Channels.getChannel(guildId)
	cev.dictController.SetGuildId(guildId)
	if err != nil {
		return nil, err
	}
	err = cev.dictController.Add(
		&replace.UserDictInput{
			Word:          da.word,
			Yomi:          da.yomi,
			ChangedUserId: authorId,
			GuildId:       guildId,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("dict add failed `%w`", err)
	}

	return &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{{
		Title:       "単語追加",
		Description: "辞書に以下のレコードを追加しました。",
		Fields:      []*discordgo.MessageEmbedField{{Name: da.word, Value: da.yomi}}},
	}}, nil
}

func newDictDel(opt []*discordgo.ApplicationCommandInteractionDataOption) (*dictDel, error) {
	var dd dictDel
	for _, o := range opt {
		switch o.Name {
		case "id":
			dd.id = uint(o.IntValue())
		default:
			return nil, fmt.Errorf("undefined option appear in dict del handler")
		}
	}
	return &dd, nil
}

func (dd *dictDel) execute(guildId, _ string) (*discordgo.InteractionResponseData, error) {
	if dd.id == 0 {
		return nil, fmt.Errorf("dict del id is not provided")
	}
	cev, err := ceviord.Channels.getChannel(guildId)
	if err != nil {
		return nil, err
	}
	cev.dictController.SetGuildId(guildId)
	del, err := cev.dictController.Delete(dd.id)
	if err != nil {
		return nil, fmt.Errorf("dict delete failed `%w`", err)
	}
	log.Printf("dict delete succeed. dict is %+v\n", del)

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       "単語削除",
			Description: "辞書から以下のレコードを削除しました。",
			Fields:      []*discordgo.MessageEmbedField{{Name: del.Word, Value: del.Yomi}},
		}}}, nil
}
