package ceviord

import (
	"crypto"
	"encoding/hex"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Ceviord struct {
	isJoin        bool
	VoiceConn     *discordgo.VoiceConnection
	pickedChannel string
	cevioWav      *cevioWav
}

const prefix = "!"
const strLenMax = 150

var tmpDir = filepath.Join(os.TempDir(), "ceviord")

var ceviord = Ceviord{
	isJoin:        false,
	pickedChannel: "",
	cevioWav:      NewTalker(),
}

func FindJoinedVC(s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.Channel {
	st, err := s.GuildChannels(m.GuildID)
	if err != nil {
		log.Println(err)
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
		log.Fatalln(err)
	}

	isJoined := false
	if vcs.ChannelID != "" {
		isJoined = true
	}
	if strings.TrimPrefix(m.Content, prefix) == "sasara" && !isJoined {
		ceviord.VoiceConn, err = s.ChannelVoiceJoin(m.GuildID, FindJoinedVC(s, m).ID, false, false)
		if err != nil {
			log.Println(err)
		}
		ceviord.pickedChannel = m.ChannelID
	}
	if strings.TrimPrefix(m.Content, prefix) == "bye" && isJoined {
		ceviord.VoiceConn.Close()
		return
	}

	if !(isJoined && m.ChannelID == ceviord.pickedChannel) {
		return
	}

	fPath, err := RandFilePathGen(m)
	if err != nil {
		log.Println(err)
		return
	}
	fPath = filepath.Join(tmpDir, fPath)
	err = os.MkdirAll(filepath.Dir(fPath), os.FileMode(0755))
	if err != nil {
		log.Println(err)
		return
	}
	err = ceviord.cevioWav.OutputWaveToFile(GetMsg(m), fPath)
	if err != nil {
		log.Println(err)
		return
	}
	dgvoice.PlayAudioFile(ceviord.VoiceConn, fPath, make(chan bool))

	//if vcs.ChannelID == "" {
	//	s.ChannelVoiceJoin(m.GuildID, FindJoinedVC(s, m).ID, false, false)
	//}

}

func VoiceStateUpdate(session *discordgo.Session, update discordgo.VoiceStateUpdate) {

}

func RandFilePathGen(m *discordgo.MessageCreate) (string, error) {
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
	return (m.Member.Nick + m.Content)[0:strLenMax]
}
