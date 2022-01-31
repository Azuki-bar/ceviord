package replace

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"strings"
)

type Dict struct {
	gorm.Model
	Before  string `gorm:"not null"`
	After   string `gorm:"not null"`
	AddUser string `gorm:"not null"`
	GuildId string `gorm:"not null"`
}
type Replacer struct {
	db      *gorm.DB
	guildId string
}

func NewReplacer(db *gorm.DB) (*Replacer, error) {
	rs := &Replacer{}
	err := rs.SetDb(db)
	if err != nil {
		return nil, err
	}
	return rs, nil
}
func (rs *Replacer) SetDb(db *gorm.DB) error {
	rs.db = db
	err := db.AutoMigrate(&Dict{})
	if err != nil {
		return fmt.Errorf("db auto migration failed `%w`", err)
	}
	return nil
}
func (rs *Replacer) SetGuildId(guildId string) { rs.guildId = guildId }
func (rs *Replacer) Add(dict *Dict) error {
	findRes := Dict{}
	result := rs.db.Where("before = ?", dict.Before).First(&findRes)
	isExist := errors.Is(result.Error, gorm.ErrRecordNotFound)
	if isExist {
		result = rs.db.Create(dict)
	} else {
		result = rs.db.Save(dict)
	}
	return result.Error
}
func (rs *Replacer) Delete(dictId uint) ([]Dict, error) {
	result := rs.db.Where(&Dict{GuildId: rs.guildId, Model: gorm.Model{ID: dictId}}).First(&Dict{})
	if result.Error != nil {
		return []Dict{}, result.Error
	}
	var deletedRecord []Dict
	result = rs.db.Delete(&deletedRecord, dictId)
	return deletedRecord, result.Error
}

func (rs *Replacer) ApplyUserDict(msg string) (string, error) {
	var records []Dict
	res := rs.db.Where(&Dict{GuildId: rs.guildId}).Find(&records)
	if res.Error == nil {
	} else if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
	} else {
		return "", res.Error
	}
	d := dicts(records)
	return d.replace(msg), nil
}

func ApplySysDict(msg string) string {
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

type dicts []Dict

func (ds *dicts) replace(msg string) string {
	rMsg := []rune(msg)
	for cur := 0; cur < len(rMsg); {
		isReplaced := false
		for _, record := range *ds {
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
	return string(rMsg)
}
