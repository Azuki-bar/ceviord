package slashCmd

import (
	"context"
	"fmt"
	"log"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/bwmarrin/discordgo"
	"github.com/k0kubun/pp"
	"go.uber.org/zap"
)

type dict struct {
	logger *zap.Logger
}

func (d *dict) handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	cev, err := ceviord.Cache.Channels.GetChannel(i.GuildID)
	cev.DictController.SetGuildID(i.GuildID)
	if err != nil {
		// voice channel connection not found
		replySimpleMsg(d.logger, fmt.Sprintf("dict handler failed. err is `%s`", err.Error()), s, i.Interaction)
		return
	}
	subCmd, err := dictSubCmdParse(d.logger, i.ApplicationCommandData().Options[0])
	if err != nil {
		replySimpleMsg(d.logger, fmt.Sprintf("dict sub cmd parser failed. err is `%s`", err.Error()), s, i.Interaction)
		return
	}
	response, err := subCmd.execute(i.GuildID, i.Member.User.ID)
	if err != nil {
		pp.Print(err)
		replySimpleMsg(d.logger, fmt.Sprintf("dict sub cmd handler failed. err is `%s`", err.Error()), s, i.Interaction)
		return
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: response,
	})
	if err != nil {
		log.Println(err)
	}
	ctx.Done()
}

func replySimpleMsg(logger *zap.Logger, msg string, s *discordgo.Session, i *discordgo.Interaction) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	if err != nil {
		logger.Warn("reply simple msg", zap.Error(err), zap.String("msg", msg))
	}
}

func dictSubCmdParse(logger *zap.Logger, opt *discordgo.ApplicationCommandInteractionDataOption) (dictSubCmd, error) {
	if opt.Type != discordgo.ApplicationCommandOptionSubCommand {
		return nil, fmt.Errorf("option type failed")
	}
	switch opt.Name {
	case "add":
		return newDictAdd(logger, opt.Options)
	case "del":
		return newDictDel(logger, opt.Options)
	case "show":
		return newDictShow(opt.Options)
	case "dump":
		return &dictDump{}, nil
	default:
		return nil, fmt.Errorf("dict sub command parse failed. %s", opt.Name)
	}
}

type (
	dictSubCmd interface {
		execute(guildID, authorID string) (*discordgo.InteractionResponseData, error)
	}
	dictAdd struct {
		yomi   string
		word   string
		logger *zap.Logger
	}
	dictDel struct {
		id     uint
		logger *zap.Logger
	}
	dictShow struct {
		isLatest bool
		limit    uint
	}
	dictDump struct {
	}
)

func newDictAdd(logger *zap.Logger, opts []*discordgo.ApplicationCommandInteractionDataOption) (*dictAdd, error) {

	da := dictAdd{logger: logger}
	for _, o := range opts {
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
func (da *dictAdd) execute(guildID, authorID string) (*discordgo.InteractionResponseData, error) {
	if len(da.word) == 0 || len(da.yomi) == 0 {
		return nil, fmt.Errorf("dict add field are not satisfied")
	}
	if ceviord.Cache.Channels == nil {
		return nil, fmt.Errorf("channel connection not found")
	}
	cev, err := ceviord.Cache.Channels.GetChannel(guildID)
	cev.DictController.SetGuildID(guildID)
	if err != nil {
		return nil, err
	}
	err = cev.DictController.Add(
		&replace.UserDictInput{
			Word:          da.word,
			Yomi:          da.yomi,
			ChangedUserID: authorID,
			GuildID:       guildID,
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

func newDictDel(logger *zap.Logger, opts []*discordgo.ApplicationCommandInteractionDataOption) (*dictDel, error) {
	dd := dictDel{logger: logger}
	for _, o := range opts {
		switch o.Name {
		case "id":
			dd.id = uint(o.IntValue())
		default:
			return nil, fmt.Errorf("undefined option appear in dict del handler")
		}
	}
	return &dd, nil
}

func (dd *dictDel) execute(guildID, _ string) (*discordgo.InteractionResponseData, error) {
	if dd.id == 0 {
		return nil, fmt.Errorf("dict del id is not provided")
	}
	cev, err := ceviord.Cache.Channels.GetChannel(guildID)
	if err != nil {
		return nil, err
	}
	cev.DictController.SetGuildID(guildID)
	del, err := cev.DictController.Delete(dd.id)
	if err != nil {
		return nil, fmt.Errorf("dict delete failed `%w`", err)
	}
	dd.logger.Info("dict delte succeed", zap.Any("delete entry", del))

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
func (ds *dictShow) execute(guildID, authorID string) (*discordgo.InteractionResponseData, error) {
	dicts, err := fetchRecords(guildID, ds.limit)
	if err != nil {
		return nil, err
	}
	returnedStr := make([]string, 1)
	cur := 0
	returnedStr[cur] = ds.getOptStr()

	for _, s := range dicts.GetStringSlice() {
		if len([]rune(returnedStr[cur]+s+"\n")) >= ceviord.DiscordPostLenLimit {
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

func fetchRecords(guildID string, limit uint) (*replace.Dicts, error) {
	var lists []replace.Dict
	cev, err := ceviord.Cache.Channels.GetChannel(guildID)
	if err != nil || cev == nil {
		return nil, err
	}
	lists, err = cev.DictController.Dump(limit)
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

func (dd *dictDump) execute(guildID, authorID string) (*discordgo.InteractionResponseData, error) {
	dicts, err := fetchRecords(guildID, 1<<32-1)
	if err != nil {
		return nil, err
	}
	returnedStr := make([]string, 1)
	cur := 0
	returnedStr[cur] = dd.getOptStr()

	for _, s := range dicts.GetStringSlice() {
		if len([]rune(returnedStr[cur]+s+"\n")) >= ceviord.DiscordPostLenLimit {
			returnedStr = append(returnedStr, s+"\n")
			cur++
		} else {
			returnedStr[cur] += s + "\n"
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
	return "全てのレコードを表示します\n"
}
