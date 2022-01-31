package replace

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"regexp"
	"strings"
)

type Record struct {
	gorm.Model
	Before  string `gorm:"not null"`
	After   string `gorm:"not null"`
	AddUser string `gorm:"not null"`
	GuildId int    `gorm:"not null"`
}
type Replacer struct {
	db      *gorm.DB
	GuildId int
}

func (rs *Replacer) SetDb(db *gorm.DB) {
	rs.db = db
}
func (rs *Replacer) CloseDb() {
	db, err := rs.db.DB()
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
	}
	err = db.Close()
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
	}
}
func (rs *Replacer) Add(record *Record) error {
	findRes := Record{}
	result := rs.db.Where("before = ?", record.Before).First(&findRes)
	isExist := errors.Is(result.Error, gorm.ErrRecordNotFound)
	if isExist {
		result = rs.db.Create(record)
	} else {
		result = rs.db.Save(record)
	}
	return result.Error
}
func (rs *Replacer) Delete(recordId uint) ([]Record, error) {
	result := rs.db.Where(&Record{GuildId: rs.GuildId, Model: gorm.Model{ID: recordId}}).First(&Record{})
	if result.Error != nil {
		return []Record{}, result.Error
	}
	var deletedRecord []Record
	result = rs.db.Delete(&deletedRecord, recordId)
	return deletedRecord, result.Error
}

func (rs *Replacer) Replace(msg string) (string, error) {
	rMsg := []rune(msg)
	var records []Record
	res := rs.db.Where(&Record{GuildId: rs.GuildId}).Find(&records)
	if res.Error == nil {
	} else if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
	} else {
		return "", res.Error
	}
	for cur := 0; cur < len(rMsg); {
		isReplaced := false
		for _, record := range records {
			befLen := len([]rune(record.Before))
			if cur+befLen > len(rMsg) {
				continue
			}
			if !strings.Contains(string(rMsg[cur:cur+befLen]), record.Before) {
				continue
			}
			rMsg = append(rMsg[0:cur], []rune(strings.Replace(string(rMsg[cur:]), record.Before, record.After, 1))...)
			cur += len([]rune(record.After))
			isReplaced = true
			break
		}
		if !isReplaced {
			cur++
		}
	}
	return string(rMsg), nil
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
