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
	if len(cmd) < 2 {
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
		err := ceviord.dictController.Add(&replace.UserDictInput{Word: cmd[2], Yomi: strings.Join(cmd[3:], ""),
			ChangedUserId: authorId, GuildId: guildId})
		if err != nil {
			return fmt.Errorf("dict add failed `%w`", err)
		}
		log.Println("dictionary add succeed")
		log.Println(cmd)
	case "del":
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
	default:
		return fmt.Errorf("dictionaly cmd not found")
	}
	return nil
}
