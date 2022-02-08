package replace

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"regexp"
	"strings"
	"time"
)

type Props struct {
	ID        uint      `db:"id, primarykey, autoincrement"`
	CreatedAt time.Time `db:"created_at, notnull"`
	UpdatedAt time.Time `db:"updated_at, notnull"`
}
type UserDictInput struct {
	Word          string `db:"word, notnull"`
	Yomi          string `db:"yomi, notnull"`
	ChangedUserId string `db:"changed_user_id, notnull"`
	GuildId       string `db:"guild_id, notnull"`
}
type Dict struct {
	Props
	UserDictInput
}
type Replacer struct {
	gorpDb  *gorp.DbMap
	guildId string
}
type DbController interface {
	Add(dict *UserDictInput) error
	Delete(dictId uint) (Dict, error)
	ApplyUserDict(msg string) (string, error)
	SetGuildId(guildId string)
	Dump() ([]Dict, error)
}

func initDb(db *sql.DB) (*gorp.DbMap, error) {
	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbMap.AddTableWithName(Dict{}, "Dicts").SetKeys(true, "ID")
	err := dbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Println(fmt.Errorf("create table failed `%w`", err))
		return nil, err
	}
	return dbMap, nil
}
func NewReplacer(db *sql.DB) (*Replacer, error) {
	rs := &Replacer{}
	dbMap, err := initDb(db)
	if err != nil {
		return nil, err
	}
	rs.gorpDb = dbMap
	return rs, nil
}
func (rs *Replacer) SetGuildId(guildId string) { rs.guildId = guildId }

func (rs *Replacer) Add(dict *UserDictInput) error {
	var findRes []Dict
	_, err := rs.gorpDb.Select(&findRes, "select * from dicts where word = ? and guild_id = ? order by updated_at desc;",
		dict.Word, dict.GuildId)
	if err != nil {
		return fmt.Errorf("upsert failed `%w`", err)
	}
	isExist := len(findRes) != 0
	if !isExist {
		insertDict := Dict{
			Props:         Props{ID: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			UserDictInput: *dict,
		}
		err := rs.gorpDb.Insert(&insertDict)
		if err != nil {
			return fmt.Errorf("insert failed `%w`", err)
		}
		return nil
	} else {
		if len(findRes) > 1 {
			i, err := rs.gorpDb.Delete(findRes[1:])
			if i == 0 {
				log.Printf("want to delete dupricate record but not deleted")
			}
			if err != nil {
				return fmt.Errorf("deplicate record delete err `%w`", err)
			}
		}
		updateDict := findRes[0]
		updateDict.UpdatedAt = time.Now()
		updateDict.UserDictInput = *dict
		i, err := rs.gorpDb.Update([]Dict{updateDict})
		if i != 0 {
			return fmt.Errorf("no update execute")
		}
		if err != nil {
			return fmt.Errorf("update execute failed `%w`", err)
		}
		return nil
	}
}

func (rs *Replacer) Delete(dictId uint) (Dict, error) {
	dict := Dict{}
	err := rs.gorpDb.SelectOne(&dict, "select * from dicts where guild_id = ? and id = ?", rs.guildId, dictId)
	if err != nil {
		return Dict{}, fmt.Errorf("record not found `%w`", err)
	}
	_, err = rs.gorpDb.Delete(&dict)
	return dict, err
}

func (rs *Replacer) ApplyUserDict(msg string) (string, error) {
	var records []Dict
	_, err := rs.gorpDb.Select(&records, "select (word, yomi) from dicts where guild_id = ? order by length(word) desc, updated_at desc;", rs.guildId)
	if err != nil {
		return "", fmt.Errorf("retrieve user dict failed `%w`", err)
	}
	d := Dicts(records)
	return d.replace(msg), nil
}
func (rs *Replacer) Dump() ([]Dict, error) {
	var dictList []Dict
	_, err := rs.gorpDb.Select(&dictList, "select * from dicts where guild_id = ? order by updated_at desc", rs.guildId)
	if err != nil {
		return nil, fmt.Errorf("dump dict error `%w`", err)
	}
	return dictList, nil
}
func ApplySysDict(msg string) string {
	type dict struct {
		before *regexp.Regexp
		after  string
	}
	dicts := []dict{
		{before: regexp.MustCompile(`https?://.*`), after: "URL "},
		{before: regexp.MustCompile("(?s)```(.*)```"), after: "コードブロック"},
		{before: regexp.MustCompile("\n"), after: " "},
		{before: regexp.MustCompile("~"), after: "ー"},
		{before: regexp.MustCompile("〜"), after: "ー"},
	}

	for _, d := range dicts {
		msg = d.before.ReplaceAllString(msg, d.after)
	}
	return replaceCustomEmoji(msg)
}
func replaceCustomEmoji(msg string) string {
	return regexp.MustCompile(`<a?:.*:.*>`).ReplaceAllString(msg, "")
}

func (ds *Dicts) Dump() []string {
	var res []string
	res = append(res, "ID\t単語\tよみ", "---------------------------")
	for _, v := range *ds {
		res = append(res, fmt.Sprintf("%d\t%s\t%s", v.ID, v.Word, v.Yomi))
	}
	return res
}

type Dicts []Dict

func (ds *Dicts) replace(msg string) string {
	rMsg := []rune(msg)
	for cur := 0; cur < len(rMsg); {
		isReplaced := false
		for _, record := range *ds {
			befLen := len([]rune(record.Word))
			if cur+befLen > len(rMsg) {
				continue
			}
			if !strings.EqualFold(string(rMsg[cur:cur+befLen]), record.Word) {
				continue
			}
			rMsg = append(rMsg[0:cur],
				[]rune(strings.Replace(
					toLowerSubStr(string(rMsg[cur:]), 0, befLen),
					strings.ToLower(record.Word),
					record.Yomi,
					1),
				)...)
			cur += len([]rune(record.Yomi))
			isReplaced = true
			break
		}
		if !isReplaced {
			cur++
		}
	}
	return string(rMsg)
}

func toLowerSubStr(s string, start, end int) string {
	rs := []rune(s)
	res := rs[0:start]
	replaced := []rune(strings.ToLower(string(rs[start:end])))
	res = append(res, replaced...)
	res = append(res, rs[end:]...)
	return string(res)
}
