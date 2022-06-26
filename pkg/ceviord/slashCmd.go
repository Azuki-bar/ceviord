package ceviord

import (
	"fmt"
	"log"
	"time"

	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/bwmarrin/discordgo"
	"github.com/k0kubun/pp"
)

const (
	joinCmdName   = "join"
	byeCmdName    = "bye"
	helpCmdName   = "help"
	dictCmdName   = "dict"
	changeCmdName = "cast"
	pingCmdName   = "ping"
)

type SlashCmdGenerator struct {
	cmds []*discordgo.ApplicationCommand
}

func NewSlashCmdGenerator() *SlashCmdGenerator {
	s := SlashCmdGenerator{cmds: cmds}
	return &s
}
func (s *SlashCmdGenerator) AddCastOpt(ps []Parameter) error {
	var c *discordgo.ApplicationCommand
	for _, rawC := range s.cmds {
		if rawC.Name == changeCmdName {
			c = rawC
			break
		}
	}
	if c == nil {
		return fmt.Errorf("change cast command not found")
	}
	castOptPos := -1
	for i, o := range c.Options {
		if o.Name == "cast" {
			castOptPos = i
			break
		}
	}
	if castOptPos < 0 {
		return fmt.Errorf("cast option not found")
	}
	co := c.Options[castOptPos]
	for _, p := range ps {
		co.Choices = append(co.Choices,
			&discordgo.ApplicationCommandOptionChoice{
				Name:  p.Name,
				Value: p.Name,
			})
	}
	return nil
}

func (s *SlashCmdGenerator) Generate() []*discordgo.ApplicationCommand {
	return s.cmds
}

var cmds = []*discordgo.ApplicationCommand{
	{
		Name:        joinCmdName,
		Description: "join voice actor",
	},
	{
		Name:        byeCmdName,
		Description: "voice actor disconnect",
	},
	{
		Name:        helpCmdName,
		Description: "get command reference",
	},
	{
		Name:        pingCmdName,
		Description: "check connection status",
	},
	{
		Name:        dictCmdName,
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
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "length",
						Description: "specify number of records",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    false,
					},
				},
			},
			{
				Name:        "dump",
				Description: "dump all records",
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
	{
		Name:        changeCmdName,
		Description: "change voice actor",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "cast",
				Description: "cast name",
				Type:        discordgo.ApplicationCommandOptionString,
				Choices:     []*discordgo.ApplicationCommandOptionChoice{},
				Required:    true,
			},
		},
	},
}

func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h, err := parseCommands(i.ApplicationCommandData().Name)
	if err != nil {
		logger.Log(logging.INFO, fmt.Errorf("parse command failed err is `%w`", err))
		return
	}
	finish := make(chan bool, 0)
	// TODO; タイムアウト時に handle内でメッセージを送信しないように変更。
	go h.handle(finish, s, i)
	select {
	case <-finish:
		return
	case <-time.After(2500 * time.Millisecond):
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "handler connection timeout"},
		})
	}
}

type (
	join   struct{}
	leave  struct{}
	help   struct{}
	ping   struct{}
	dict   struct{}
	change struct {
		changeTo string
	}
)
type CommandHandler interface {
	handle(finished chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate)
}

func parseCommands(name string) (CommandHandler, error) {
	var h CommandHandler
	switch name {
	case joinCmdName:
		h = new(join)
	case byeCmdName:
		h = new(leave)
	case helpCmdName:
		h = new(help)
	case "ping":
		h = new(ping)
	case "dict":
		h = new(dict)
	case "cast":
		h = new(change)
	default:
		return nil, fmt.Errorf("command `%s` is not found", name)
	}
	return h, nil
}

func (j *join) handle(c chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	c <- true
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("error in `join` interaction respond err is `%w`", err))
	}
}
func (*join) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

func (l *leave) handle(finish chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
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
func (*leave) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

func (*help) handle(c chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	c <- true
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("help handler failed err is `%w`", err))
	}
}
func (*ping) handle(c chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "your message have been trapped on ceviord server"},
	})
	c <- true
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("ping handler failed err is `%w`", err))
	}
}

func (c *change) handle(finish chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, o := range i.ApplicationCommandData().Options {
		switch o.Name {
		case "cast":
			c.changeTo = o.StringValue()
		}
	}
	err := c.rawHandle(s, i)
	msg := fmt.Sprintf("successfully change cast to %s", c.changeTo)
	if err != nil {
		msg = err.Error()
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	finish <- true
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("change handler failed. err is `%w`", err))
	}
}
func (c *change) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	cev, err := ceviord.Channels.getChannel(i.GuildID)
	if err != nil {
		return fmt.Errorf("voice connection not found")
	}
	isJoin, err := cev.isActorJoined(s)
	if err != nil || !isJoin {
		return fmt.Errorf("voice connection not found")
	}
	for _, p := range ceviord.param.Parameters {
		if c.changeTo == p.Name {
			cev.currentParam = &p
			if err := rawSpeak(fmt.Sprintf("パラメータを %s に変更しました。", p.Name), i.GuildID, s); err != nil {
				return fmt.Errorf("speaking about parameter setting: `%w`", err)
			}
		}
	}
	return nil
}

func (*dict) handle(c chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	c <- true
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
	case "show":
		return newDictShow(opt.Options)
	case "dump":
		return NewDictDump(opt.Options)
	default:
		return nil, fmt.Errorf("dict sub command parse failed. %s", opt.Name)
	}
}

type (
	dictSubCmd interface {
		execute(guildId, authorId string) (*discordgo.InteractionResponseData, error)
	}
	dictAdd struct {
		yomi string
		word string
	}
	dictDel struct {
		id uint
	}
	dictShow struct {
		isLatest bool
		limit    uint
	}
	dictDump struct {
	}
)

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

const defaultDictShowLimit = 10

func newDictShow(opt []*discordgo.ApplicationCommandInteractionDataOption) (*dictShow, error) {
	ds := dictShow{limit: 0, isLatest: false}
	for _, o := range opt {
		switch o.Name {
		case "length":
			ds.limit = uint(o.IntValue())
		default:
			return nil, fmt.Errorf("undefined option appear in dict show handler")
		}
	}
	if ds.limit == 0 {
		ds.limit = defaultDictShowLimit
	}
	return &ds, nil
}
func (ds *dictShow) execute(guildId, authorId string) (*discordgo.InteractionResponseData, error) {
	dicts, err := fetchRecords(guildId, ds.limit)
	if err != nil {
		return nil, err
	}
	returnedStr := make([]string, 1)
	cur := 0
	returnedStr[cur] = ds.getOptStr()

	for _, s := range dicts.GetStringSlice() {
		if len([]rune(returnedStr[cur]+s+"\n")) >= discordPostLenLimit {
			returnedStr = append(returnedStr, s+"\n")
			cur++
		} else {
			returnedStr[cur] += (s + "\n")
		}
	}
	emds := make([]*discordgo.MessageEmbed, 0)
	for i, v := range returnedStr {
		e := discordgo.MessageEmbed{
			Title:       fmt.Sprintf("page %d/%d", i+1, len(returnedStr)),
			Description: v,
		}
		emds = append(emds, &e)
	}
	pp.Println(emds)
	return &discordgo.InteractionResponseData{
		Title:  "dict record",
		Embeds: emds}, nil
}

func fetchRecords(guildId string, limit uint) (*replace.Dicts, error) {
	var lists []replace.Dict
	cev, err := ceviord.Channels.getChannel(guildId)
	if err != nil || cev == nil {
		return nil, err
	}
	lists, err = cev.dictController.Dump(limit)
	if err != nil {
		return nil, fmt.Errorf("dictionary get failed `%w`", err)
	}
	if lists == nil {
		return nil, fmt.Errorf("fetch dict records failed")
	}
	ds := replace.Dicts(lists)
	return &ds, nil
}

func (ds *dictShow) getOptStr() string {
	return fmt.Sprintf("直近の%dレコードを表示します\n", ds.limit)
}

func NewDictDump(opt []*discordgo.ApplicationCommandInteractionDataOption) (*dictDump, error) {
	return &dictDump{}, nil
}
func (dd *dictDump) execute(guildId, authorId string) (*discordgo.InteractionResponseData, error) {
	dicts, err := fetchRecords(guildId, 1<<32-1)
	if err != nil {
		return nil, err
	}
	returnedStr := make([]string, 1)
	cur := 0
	returnedStr[cur] = dd.getOptStr()

	for _, s := range dicts.GetStringSlice() {
		if len([]rune(returnedStr[cur]+s+"\n")) >= discordPostLenLimit {
			returnedStr = append(returnedStr, s+"\n")
			cur++
		} else {
			returnedStr[cur] += (s + "\n")
		}
	}
	emds := make([]*discordgo.MessageEmbed, 0)
	for i, v := range returnedStr {
		e := discordgo.MessageEmbed{
			Title:       fmt.Sprintf("page %d/%d", i+1, len(returnedStr)),
			Description: v,
		}
		emds = append(emds, &e)
	}
	pp.Println(emds)
	return &discordgo.InteractionResponseData{
		Title:  "dict record",
		Embeds: emds}, nil
}

func (dd *dictDump) getOptStr() string {
	return fmt.Sprintf("全てのレコードを表示します\n")
}
