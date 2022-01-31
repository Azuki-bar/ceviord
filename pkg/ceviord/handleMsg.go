package ceviord

import (
	"ceviord/pkg/replace"
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Ceviord struct {
	isJoin        bool
	VoiceConn     *discordgo.VoiceConnection
	pickedChannel string
	cevioWav      *cevioWav
	mutex         sync.Mutex
	replacer      replace.Replacer
}

const prefix = "!"
const strLenMax = 150

var tmpDir = filepath.Join(os.TempDir(), "ceviord")

var ceviord = Ceviord{
	isJoin:        false,
	pickedChannel: "",
	mutex:         sync.Mutex{},
}

func SetNewTalker(wav *cevioWav)     { ceviord.cevioWav = wav }
func SetReplacer(r replace.Replacer) { ceviord.replacer = r }

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
			log.Println(fmt.Errorf("%w", err))
		}
		ceviord.pickedChannel = m.ChannelID
	}
	if strings.TrimPrefix(m.Content, prefix) == "bye" && isJoined {
		defer ceviord.VoiceConn.Close()
		err = ceviord.VoiceConn.Speaking(false)
		if err != nil {
			log.Println(fmt.Errorf("%w", err))
		}
		err = ceviord.VoiceConn.Disconnect()
		if err != nil {
			log.Println(fmt.Errorf("%w", err))
		}
		return
	}

	// todo; なにかの関数に押し込む
	dictCmd := "dict"
	if strings.HasPrefix(strings.TrimPrefix(m.Content, prefix), dictCmd+" ") {
		var cmd []string
		for _, c := range strings.Split(strings.TrimPrefix(m.Content, prefix), " ")[1:] {
			if c != "" {
				cmd = append(cmd, c)
			}
		}
		if len(cmd) < 3 {

		}
		switch cmd[0] {
		case "add":
			err := ceviord.replacer.Add(&replace.Dict{Before: cmd[1], After: strings.Join(cmd[2:], ""),
				AddUser: m.Author.ID, GuildId: m.GuildID})
			if err != nil {
			}
		case "del":
			id, err := strconv.Atoi(cmd[1])
			if err != nil {
			}
			_, err = ceviord.replacer.Delete(uint(id))
			if err != nil {

			}
		}

	}

	if !(isJoined && m.ChannelID == ceviord.pickedChannel) {
		return
	}

	ceviord.mutex.Lock()
	fPath, err := RandFileNameGen(m)
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
		return
	}
	fPath = filepath.Join(tmpDir, fPath)
	err = os.MkdirAll(filepath.Dir(fPath), os.FileMode(0755))
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
		return
	}
	err = ceviord.cevioWav.OutputWaveToFile(GetMsg(m), fPath)
	defer os.Remove(fPath)
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
		return
	}
	dgvoice.PlayAudioFile(ceviord.VoiceConn, fPath, make(chan bool))
	ceviord.mutex.Unlock()

	//if vcs.ChannelID == "" {
	//	s.ChannelVoiceJoin(m.GuildID, FindJoinedVC(s, m).ID, false, false)
	//}

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
	msg := []rune(name + "。" + replace.ApplySysDict(m.Content))
	if len(msg) > strLenMax {
		return string(msg[0:strLenMax])
	} else {
		return string(msg)
	}
}
