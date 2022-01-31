package ceviord

import (
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

type Ceviord struct {
	isJoin        bool
	VoiceConn     *discordgo.VoiceConnection
	pickedChannel string
	cevioWav      *cevioWav
	parameters    *Parameters
	currentParam  *Parameter
	mutex         sync.Mutex
}

type Parameters struct {
	Parameters []Parameter `yaml:"parameters"`
}

type Parameter struct {
	Name      string         `yaml:"name"`
	Cast      string         `yaml:"cast"`
	Volume    int            `yaml:"volume"`
	Speed     int            `yaml:"speed"`
	Tone      int            `yaml:"tone"`
	Tonescale int            `yaml:"tonescale"`
	Alpha     int            `yaml:"alpha"`
	Emotions  map[string]int `yaml:"emotions"`
}

const prefix = "!"
const strLenMax = 150

var tmpDir = filepath.Join(os.TempDir(), "ceviord")

var ceviord = Ceviord{
	isJoin:        false,
	pickedChannel: "",
	mutex:         sync.Mutex{},
}

func SetNewTalker(wav *cevioWav) {
	ceviord.cevioWav = wav
}

func SetParameters(para *Parameters) {
	ceviord.parameters = para
}

func FindJoinedVC(s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.Channel {
	st, err := s.GuildChannels(m.GuildID)
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
		return nil
	}
	vcs, err := s.State.VoiceState(m.GuildID, m.Author.ID)
	for _, c := range st {
		switch c.Type {
		case discordgo.ChannelTypeGuildVoice:
			if c.ID == vcs.ChannelID {
				return c
			}
		}
	}
	return nil
}

// MessageCreate will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example, but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Author.Bot {
		return
	}
	vcs, err := s.State.VoiceState(m.GuildID, s.State.User.ID)
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
	}

	isJoined := false
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
		//todo; implement
		isJoined = false
	} else if vcs.ChannelID != "" {
		isJoined = true
	}

	if strings.TrimPrefix(m.Content, prefix) == "sasara" && !isJoined {
		ceviord.VoiceConn, err = s.ChannelVoiceJoin(m.GuildID, FindJoinedVC(s, m).ID, false, false)
		if err != nil {
			log.Println(fmt.Errorf("joining: %w", err))
		}
		ceviord.pickedChannel = m.ChannelID
	}
	if strings.TrimPrefix(m.Content, prefix) == "bye" && isJoined {
		defer ceviord.VoiceConn.Close()
		err = ceviord.VoiceConn.Speaking(false)
		if err != nil {
			log.Println(fmt.Errorf("speaking falsing: %w", err))
		}
		err = ceviord.VoiceConn.Disconnect()
		if err != nil {
			log.Println(fmt.Errorf("disconnecting: %w", err))
		}
		return
	}

	fmt.Println(strings.TrimPrefix(m.Content, "!"))
	if strings.HasPrefix(strings.TrimPrefix(m.Content, prefix), "change ") {
		for _, p := range ceviord.parameters.Parameters {
			got := strings.TrimPrefix(m.Content, prefix+"change ")
			if got == p.Name {
				ceviord.currentParam = &p
				ceviord.cevioWav.ApplyEmotions(ceviord.currentParam)
				err := speak(fmt.Sprintf("パラメータを %s に変更しました", p.Name))
				if err != nil {
					log.Println(fmt.Errorf("speaking about paramerter setting: %w", err))
				}
			}
		}
		return
	}

	if !(isJoined && m.ChannelID == ceviord.pickedChannel) {
		return
	}

	ceviord.mutex.Lock()
	err = speak(GetMsg(m))
	if err != nil {
		log.Println(err)
	}
	ceviord.mutex.Unlock()

	//if vcs.ChannelID == "" {
	//	s.ChannelVoiceJoin(m.GuildID, FindJoinedVC(s, m).ID, false, false)
	//}

}

func speak(text string) error {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return fmt.Errorf("generating rand: %w", err)
	}
	fPath := fmt.Sprintf("%8x", buf)
	fPath = filepath.Join(tmpDir, fPath)
	err = os.MkdirAll(filepath.Dir(fPath), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("making dir: %w", err)
	}
	err = ceviord.cevioWav.OutputWaveToFile(text, fPath)
	defer os.Remove(fPath)
	if err != nil {
		return fmt.Errorf("outputting: %w", err)
	}
	dgvoice.PlayAudioFile(ceviord.VoiceConn, fPath, make(chan bool))
	return nil
}

func VoiceStateUpdate(session *discordgo.Session, update discordgo.VoiceStateUpdate) {

}

func RandFileNameGen(m *discordgo.MessageCreate) (string, error) {
	hash := crypto.MD5.New()
	defer hash.Reset()
	t, err := m.Timestamp.Parse()
	if err != nil {
		return "", err
	}
	hash.Write([]byte(t.String() + m.Content))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func GetMsg(m *discordgo.MessageCreate) string {
	var name string
	if m.Member.Nick == "" {
		name = m.Author.Username
	} else {
		name = m.Member.Nick
	}
	msg := []rune(m.Content)
	msg = []rune(name + "。" + ReplaceMsg(string(msg)))
	if len(msg) > strLenMax {
		return string(msg[0:strLenMax])
	} else {
		return string(msg)
	}
}

func ReplaceMsg(msg string) string {
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
