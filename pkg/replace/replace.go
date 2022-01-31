package replace

import (
	"gorm.io/gorm"
	"regexp"
	"strings"
)

type Record struct {
	gorm.Model
	Before  string `gorm:"not null"`
	After   string `gorm:"not null"`
	AddUser string `gorm:"not null"`
	GuildId string `gorm:"not null"`
}
type Records []Record

func (rs *Records) Replace(msg string) string {
	rMsg := []rune(msg)
	for cur := 0; cur < len(rMsg); {
		isReplaced := false
		for _, r := range *rs {
			befLen := len([]rune(r.Before))
			if cur+befLen > len(rMsg) {
				continue
			}
			if !strings.Contains(string(rMsg[cur:cur+befLen]), r.Before) {
				continue
			}
			rMsg = append(rMsg[0:cur], []rune(strings.Replace(string(rMsg[cur:]), r.Before, r.After, 1))...)
			cur += len([]rune(r.After))
			isReplaced = true
			break
		}
		if !isReplaced {
			cur++
		}
	}
	return string(rMsg)
}
func ApplyDict(msg string) string {
	type dict struct {
		before *regexp.Regexp
		after  string
	}
	var dicts []dict
	var newDict dict
	newDict.before = regexp.MustCompile(`https?://.*`)
	newDict.after = "ゆーあーるえる。"
	dicts = append(dicts, newDict)

	newDict.before = regexp.MustCompile("(?s)```(.*)```")
	newDict.after = "コードブロック"
	dicts = append(dicts, newDict)

	newDict.before = regexp.MustCompile("\n")
	newDict.after = " "
	dicts = append(dicts, newDict)

	newDict.before = regexp.MustCompile("~")
	newDict.after = "ー"
	dicts = append(dicts, newDict)

	newDict.before = regexp.MustCompile("〜")
	newDict.after = "ー"
	dicts = append(dicts, newDict)

	for _, d := range dicts {
		msg = d.before.ReplaceAllString(msg, d.after)
	}
	return msg
}
