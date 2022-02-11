package ceviord

import (
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
)

func handleDictCmd(content, authorId, guildId, dictCmd string, session *discordgo.Session) error {
	if !strings.HasPrefix(content, prefix+dictCmd) {
		return fmt.Errorf("dict cmd not called")
	}
	var cmd []string
	for _, c := range strings.Split(strings.TrimPrefix(content, prefix), " ")[1:] {
		if c != "" {
			cmd = append(cmd, c)
		}
	}
	if len(cmd) < 1 {
		return fmt.Errorf("dictionaly cmd is not specific")
	}
	if ceviord.dictController == nil {
		return fmt.Errorf("db controller is not defined")
	}
	switch cmd[0] {
	case "add":
		if len(cmd) < 3 {
			return fmt.Errorf("dictionaly yomi record not shown")
		}
		word := stringMax(cmd[1], strLenMax)
		yomi := stringMax(strings.Join(cmd[2:], " "), strLenMax)
		if err := ceviord.dictController.Add(&replace.UserDictInput{Word: word, Yomi: yomi, ChangedUserId: authorId, GuildId: guildId}); err != nil {
			return fmt.Errorf("dict add failed `%w`", err)
		}
		log.Println("dictionary add succeed")
		if err := SendEmbedMsg(
			&discordgo.MessageEmbed{
				Title:       "単語追加",
				Description: "辞書に以下のレコードを追加しました。",
				Fields:      []*discordgo.MessageEmbedField{{Name: word, Value: yomi}},
			}, session); err != nil {
			log.Println(fmt.Errorf("send add msg failed `%w`", err))
		}
	case "del", "delete":
		if len(cmd) < 2 {
			return fmt.Errorf("delete id is not specification")
		}
		id, err := strconv.Atoi(cmd[1])
		if err != nil {
			return fmt.Errorf("id specification failed `%w`", err)
		}
		d, err := ceviord.dictController.Delete(uint(id))
		if err != nil {
			return fmt.Errorf("table delete failed `%w`", err)
		}
		log.Println("dictionary delete succeed")
		if err = SendEmbedMsg(
			&discordgo.MessageEmbed{
				Title:       "単語削除",
				Description: "辞書から以下のレコードを削除しました。",
				Fields:      []*discordgo.MessageEmbedField{{Name: d.Word, Value: d.Yomi}},
			}, session); err != nil {
			log.Println(fmt.Errorf("send delete msg failed `%w`", err))
		}
	case "list", "show":
		lists, err := ceviord.dictController.Dump()
		if err != nil {
			return fmt.Errorf("dictionnary dump failed `%w`", err)
		}
		if lists == nil {
			log.Printf("no dictionary record")
			return nil
		}
		dicts := replace.Dicts(lists)
		dump := dicts.Dump()
		printsStr := make([]string, 1)
		limit := 2000
		cur := 0
		for _, d := range dump {
			if len([]rune(printsStr[cur]+d+"\n")) >= limit {
				printsStr = append(printsStr, d+"\n")
				cur++
			} else {
				printsStr[cur] = printsStr[cur] + d + "\n"
			}
		}
		for _, v := range printsStr {
			if v == "" {
				continue
			}
			if err := SendMsg(v, session); err != nil {
				return fmt.Errorf("dump dict list failed `%w`", err)
			}
		}

	default:
		return fmt.Errorf("dictionaly cmd not found")
	}
	return nil
}

func stringMax(msg string, max int) string {
	lenMsg := len([]rune(msg))
	if lenMsg > max {
		return string([]rune(msg)[0:max])
	}
	return msg
}
