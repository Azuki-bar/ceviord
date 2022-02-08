package ceviord

import (
	"ceviord/pkg/replace"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

type Ceviord struct {
	isJoin         bool
	VoiceConn      *discordgo.VoiceConnection
	pickedChannel  string
	cevioWav       CevioWav
	conf           *Config
	currentParam   *Parameter
	mutex          sync.Mutex
	dictController replace.DbController
}

type Config struct {
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

type CevioWav interface {
	OutputWaveToFile(talkWord, path string) (err error)
	ApplyEmotions(param *Parameter) (err error)
}

const prefix = "!"
const strLenMax = 150

var tmpDir = filepath.Join(os.TempDir(), "ceviord")

var ceviord = Ceviord{
	isJoin:        false,
	pickedChannel: "",
	mutex:         sync.Mutex{},
}

func SetNewTalker(wav CevioWav)              { ceviord.cevioWav = wav }
func SetDbController(r replace.DbController) { ceviord.dictController = r }
func SetParameters(para *Config)             { ceviord.conf = para }

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

	fmt.Println(strings.TrimPrefix(m.Content, prefix))
	if strings.HasPrefix(strings.TrimPrefix(m.Content, prefix), "change ") {
		for _, p := range ceviord.conf.Parameters {
			got := strings.TrimPrefix(m.Content, prefix+"change ")
			if got == p.Name {
				ceviord.currentParam = &p
				ceviord.cevioWav.ApplyEmotions(ceviord.currentParam)
				err := rawSpeak(fmt.Sprintf("パラメータを %s に変更しました", p.Name))
				if err != nil {
					log.Println(fmt.Errorf("speaking about paramerter setting: %w", err))
				}
			}
		}
		return
	}

	dictCmd := "dict"
	if strings.HasPrefix(strings.TrimPrefix(m.Content, prefix), dictCmd+" ") {
		err := handleDictCmd(m.Content, m.Author.ID, m.GuildID, dictCmd, s)
		if err != nil {
			log.Println(fmt.Errorf("dictionaly handler failed `%w`", err))
			return
		}
		return
	}

	if !(isJoined && m.ChannelID == ceviord.pickedChannel) {
		return
	}

	err = rawSpeak(GetMsg(m, s))
	if err != nil {
		log.Println(err)
	}

	//if vcs.ChannelID == "" {
	//	s.ChannelVoiceJoin(m.GuildID, FindJoinedVC(s, m).ID, false, false)
	//}

}

func rawSpeak(text string) error {
	ceviord.mutex.Lock()
	defer ceviord.mutex.Unlock()
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

func SendMsg(msg string, session *discordgo.Session) error {
	// https://discord.com/developers/docs/resources/channel#create-message-jsonform-params
	if len([]rune(msg)) > 2000 {
		return fmt.Errorf("discord message send limitation error")
	} else if len([]rune(msg)) == 0 {
		return fmt.Errorf("message len is 0")
	}
	_, err := session.ChannelMessageSend(ceviord.pickedChannel, msg)
	return err
}

func GetMsg(m *discordgo.MessageCreate, s *discordgo.Session) string {
	var name string
	if m.Member.Nick == "" {
		name = m.Author.Username
	} else {
		name = m.Member.Nick
	}
	cont, err := m.ContentWithMoreMentionsReplaced(s)
	if err != nil {
		log.Println(fmt.Errorf("replace mention failed `%w`", err))
		return ""
	}
	msg := []rune(name + "。" + replace.ApplySysDict(cont))

	ceviord.dictController.SetGuildId(m.GuildID)
	rawMsg, err := ceviord.dictController.ApplyUserDict(string(msg))
	if err != nil {
		log.Println("apply user dict failed `%w`", err)
	}
	msg = []rune(rawMsg)
	if len(msg) > strLenMax {
		return string(msg[0:strLenMax])
	} else {
		return string(msg)
	}
}
