package ceviord

import (
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
)

type userMainCmd interface {
	parse(cmd []string) error
	handle(sess *discordgo.Session, msg *discordgo.MessageCreate) error
}

type change struct {
	changeTo string
}

func (c *change) handle(_ *discordgo.Session, _ *discordgo.MessageCreate) error {
	for _, p := range ceviord.conf.Parameters {
		if c.changeTo == p.Name {
			ceviord.currentParam = &p
			if err := ceviord.cevioWav.ApplyEmotions(ceviord.currentParam); err != nil {
				return fmt.Errorf("apply emotion failed; emotion is %+v", ceviord.currentParam)
			}
			if err := rawSpeak(fmt.Sprintf("パラメータを %s に変更しました。", p.Name)); err != nil {
				return fmt.Errorf("speaking about parameter setting: `%w`", err)
			}
		}
	}
	return nil
}

func (c *change) parse(cmds []string) error {
	if len(cmds) < 2 {
		return fmt.Errorf("apply commands are not correct")
	}
	c.changeTo = cmds[1]
	return nil
}

type sasara struct{}

func (*sasara) handle(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	if ceviord.isJoin {
		return fmt.Errorf("sasara is already joined\n")
	}
	vc := FindJoinedVC(sess, msg)
	if vc == nil {
		//todo fix err msg
		return fmt.Errorf("voice conn ")
	}
	var err error
	ceviord.VoiceConn, err = sess.ChannelVoiceJoin(msg.GuildID, vc.ID, false, false)
	if err != nil {
		log.Println(fmt.Errorf("joining: %w", err))
		return err
	}
	ceviord.pickedChannel = msg.ChannelID
	return nil
}
func (*sasara) parse(_ []string) error { return nil }

type bye struct{}

func (*bye) parse(_ []string) error { return nil }
func (*bye) handle(_ *discordgo.Session, _ *discordgo.MessageCreate) error {
	if !ceviord.isJoin || ceviord.VoiceConn == nil {
		return fmt.Errorf("ceviord is already disconnected\n")
	}
	defer func() {
		if ceviord.VoiceConn != nil {
			ceviord.VoiceConn.Close()
		}
	}()
	var err error
	err = ceviord.VoiceConn.Speaking(false)
	if err != nil {
		log.Println(fmt.Errorf("speaking falsing: %w", err))
	}
	err = ceviord.VoiceConn.Disconnect()
	if err != nil {
		log.Println(fmt.Errorf("disconnecting: %w", err))
	}
	return nil
}

type dict struct {
	sub userMainCmd
}

func (d *dict) handle(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	if d.sub == nil {
		return fmt.Errorf("dict sub cmd not provided")
	}
	return d.sub.handle(sess, msg)
}

func (d *dict) parse(cmds []string) error {
	if len(cmds) <= 1 {
		return fmt.Errorf("sub cmd are not satisfied. \n")
	}
	var dictCmd userMainCmd
	switch cmds[1] {
	case "add":
		dictCmd = new(dictAdd)
	case "del", "delete", "rm":
		dictCmd = new(dictDel)
	case "list", "ls", "show":
		dictCmd = new(dictList)
	default:
		return fmt.Errorf("unknown sub command `%s`", cmds[0])
	}
	if err := dictCmd.parse(cmds[1:]); err != nil {
		return fmt.Errorf("dict subcmd `%T` parse failed", dictCmd)
	}
	d.sub = dictCmd
	return nil
}

type dictAdd struct {
	word string
	yomi string
}

func (d *dictAdd) handle(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	if len(d.word) == 0 || len(d.yomi) == 0 {
		return fmt.Errorf("dict add field are not satisfied")
	}
	err := ceviord.dictController.Add(
		&replace.UserDictInput{
			Word:          d.word,
			Yomi:          d.yomi,
			ChangedUserId: msg.Author.ID,
			GuildId:       msg.GuildID,
		},
	)
	if err != nil {
		return fmt.Errorf("dict add failed `%w`", err)
	}
	log.Println("dictionary add succeed")

	err = SendEmbedMsg(
		&discordgo.MessageEmbed{
			Title:       "単語追加",
			Description: "辞書に以下のレコードを追加しました。",
			Fields:      []*discordgo.MessageEmbedField{{Name: d.word, Value: d.yomi}},
		}, sess)
	if err != nil {
		return fmt.Errorf("send add msg failed `%w`", err)
	}
	return nil
}

func (d *dictAdd) parse(cmd []string) error {
	if len(cmd) <= 2 {
		return fmt.Errorf("dict add option are not satisfied\n")
	}
	d.word = stringMax(cmd[1], strLenMax)
	d.yomi = stringMax(cmd[2], strLenMax)
	return nil
}

type dictDel struct {
	ids []uint
}

func (d *dictDel) handle(sess *discordgo.Session, _ *discordgo.MessageCreate) error {
	if d.ids == nil || len(d.ids) == 0 {
		return fmt.Errorf("dict del id is not provided")
	}
	for _, id := range d.ids {
		del, err := ceviord.dictController.Delete(id)
		if err != nil {
			return fmt.Errorf("dict delete failed `%w`", err)
		}
		log.Printf("dict delete succeed. dict is %+v\n", del)

		err = SendEmbedMsg(
			&discordgo.MessageEmbed{
				Title:       "単語削除",
				Description: "辞書から以下のレコードを削除しました。",
				Fields:      []*discordgo.MessageEmbedField{{Name: del.Word, Value: del.Yomi}},
			}, sess)
		if err != nil {
			return fmt.Errorf("send delete msg failed `%w`", err)
		}
	}
	return nil
}

func (d *dictDel) parse(cmd []string) error {
	if len(cmd) < 2 {
		return fmt.Errorf("dict del option are not satisfied\n")
	}
	if d.ids == nil {
		d.ids = make([]uint, 0)
	}
	for _, sId := range cmd[1:] {
		id, err := strconv.Atoi(sId)
		if err != nil {
			d.ids = make([]uint, 0)
			return fmt.Errorf("parse id failed `%w`\n", err)
		}
		if id < 0 {
			d.ids = make([]uint, 0)
			return fmt.Errorf("id range error")
		}
		d.ids = append(d.ids, uint(id))
	}
	return nil
}

type dictList struct {
	isLatest bool
	from     uint
	to       uint
	limit    uint
}

func (d *dictList) parse(cmd []string) error {
	switch len(cmd) {
	case 0:
		return fmt.Errorf("dict list sub cmd not provided")
	case 1:
		d.isLatest = true
		d.limit = 10
		return nil
	}
	switch len(cmd[1:]) {
	case 1:
		d.isLatest = true
		if cmd[1] == "all" {
			d.limit = 1<<32 - 1
		} else {
			l, err := strconv.Atoi(cmd[1])
			if err != nil {
				return fmt.Errorf("dict list parse string failed `%w`", err)
			}
			if l < 0 {
				return fmt.Errorf("dict list negative number provided ")
			}
			d.limit = uint(l)
		}
	default:
		d.isLatest = false
		ids := make([]uint, 0)
		for i, sId := range cmd[1:] {
			if i >= 2 {
				break
			}
			id, err := strconv.Atoi(sId)
			if id <= 0 || err != nil {
				return fmt.Errorf("parse id failed `%w`", err)
			}
			ids = append(ids, uint(id))
		}
		d.from = ids[0]
		d.to = ids[1]
	}
	return nil
}

const discordPostLenLimit = 2000

func (d *dictList) handle(sess *discordgo.Session, _ *discordgo.MessageCreate) error {
	var lists []replace.Dict
	var err error
	if d.isLatest {
		lists, err = ceviord.dictController.Dump(d.limit)
	} else {
		lists, err = ceviord.dictController.DumpAtoB(d.from, d.to)
	}
	if err != nil {
		return fmt.Errorf("dictionary dump failed `%w`", err)
	}
	if lists == nil {
		return fmt.Errorf("fetch db records failed")
	}
	dicts := replace.Dicts(lists)
	printsStr := make([]string, 1)
	cur := 0
	printsStr[cur] = d.getOptStr()
	for _, s := range dicts.GetStringSlice() {
		if len([]rune(printsStr[cur]+s+"\n")) >= discordPostLenLimit {
			printsStr = append(printsStr, s+"\n")
		} else {
			printsStr[cur] += s + "\n"
		}
	}
	for _, v := range printsStr {
		if v == "" {
			continue
		}
		if err := SendMsg(v, sess); err != nil {
			return fmt.Errorf("send dict list to Discord failed `%w`", err)
		}
	}
	return nil
}
func (d *dictList) getOptStr() string {
	if d.isLatest {
		if d.limit == 1<<32-1 {
			return "全ての単語辞書を表示します。\n"
		} else {
			return fmt.Sprintf("直近の%dレコードを表示します\n", d.limit)
		}
	} else {
		return fmt.Sprintf("IDが%dから%dのレコードを表示します。\n", d.from, d.to)
	}
}

func stringMax(msg string, max int) string {
	lenMsg := len([]rune(msg))
	if lenMsg > max {
		return string([]rune(msg)[0:max])
	}
	return msg
}
