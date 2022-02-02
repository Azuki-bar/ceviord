package ceviord

import (
	"ceviord/pkg/replace"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func handleDictCmd(content, authorId, guildId, dictCmd string) error {
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
		err := ceviord.dictController.Add(&replace.UserDictInput{Word: stringMax(cmd[1], strLenMax), Yomi: stringMax(strings.Join(cmd[2:], " "), strLenMax),
			ChangedUserId: authorId, GuildId: guildId})
		if err != nil {
			return fmt.Errorf("dict add failed `%w`", err)
		}
		log.Println("dictionary add succeed")
		log.Println(cmd)
	case "del":
		if len(cmd) < 2 {
			return fmt.Errorf("delete id is not specification")
		}
		id, err := strconv.Atoi(cmd[1])
		if err != nil {
			return fmt.Errorf("id specification failed `%w`", err)
		}
		_, err = ceviord.dictController.Delete(uint(id))
		if err != nil {
			return fmt.Errorf("table delete failed `%w`", err)
		}
		log.Println("dictionary delete succeed")
		log.Println(cmd)
	case "list":
		lists, err := ceviord.dictController.Dump()
		if err != nil {
			return fmt.Errorf("dictionnary dump failed `%w`", err)
		}
		if lists == nil {
			log.Printf("no dictionary record")
			return nil
		}
		dicts := replace.Dicts(lists)
		d := dicts.Dump()
		printsStr := make([]string, 1)
		limit := 2000
		cur := 0
		for _, v := range d {
			for len([]rune(printsStr[cur]+v+"\n")) < limit {
				printsStr[cur] = printsStr[cur] + v + "\n"
			}
			printsStr = append(printsStr, "")
			cur++
		}
		for _, v := range printsStr {
			err := SendMsg(v)
			if err != nil {
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
