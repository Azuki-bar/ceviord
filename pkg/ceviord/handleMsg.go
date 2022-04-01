package ceviord

import (
	"crypto/rand"
	"fmt"
	"github.com/azuki-bar/ceviord/pkg/dgvoice"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Ceviord struct {
	isJoin         bool
	VoiceConn      *discordgo.VoiceConnection
	pickedChannel  string
	cevioWav       CevioWav
	param          *Param
	Auth           *Auth
	currentParam   *Parameter
	mutex          sync.Mutex
	dictController replace.DbController
}

type Param struct {
	Parameters []Parameter `yaml:"parameters"`
}

type Auth struct {
	CeviordConn Conn `yaml:"conn"`
}

type Conn struct {
	Discord string `yaml:"discord"`
	Cevio   struct {
		Token    string `yaml:"cevioToken"`
		EndPoint string `yaml:"cevioEndPoint"`
	} `yaml:",inline"`
	Dsn        string `yaml:"dsn"`
	DriverName string `yaml:"driverName"`
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
const strLenMax = 300

var tmpDir = filepath.Join(os.TempDir(), "ceviord")

var ceviord = Ceviord{
	isJoin:        false,
	pickedChannel: "",
	mutex:         sync.Mutex{},
}

func SetNewTalker(wav CevioWav)              { ceviord.cevioWav = wav }
func SetDbController(r replace.DbController) { ceviord.dictController = r }
func SetParam(param *Param)                  { ceviord.param = param; ceviord.currentParam = &param.Parameters[0] }

func FindJoinedVC(s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.Channel {
	st, err := s.GuildChannels(m.GuildID)
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
		return nil
	}
	vcs, err := s.State.VoiceState(m.GuildID, m.Author.ID)
	if err != nil {
		log.Println(fmt.Errorf("find joinedVc err occurred `%w`", err))
		return nil
	}
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

func parseUserCmd(msg string) (userMainCmd, error) {
	rawCmd := regexp.MustCompile(`[\s　]+`).Split(msg, -1)
	if len(rawCmd) < 1 {
		return nil, fmt.Errorf("parsing user cmd failed. user msg is `%s`\n", msg)
	}
	var mainCmd userMainCmd
	switch rawCmd[0] {
	case "sasara":
		mainCmd = new(sasara)
	case "bye":
		mainCmd = new(bye)
	case "dict":
		mainCmd = new(dict)
	case "change":
		mainCmd = new(change)
	case "help", "man":
		mainCmd = new(help)
	default:
		return nil, fmt.Errorf("unknown user cmd `%s` \n", rawCmd[0])
	}
	if err := mainCmd.parse(rawCmd); err != nil {
		return nil, err
	}
	return mainCmd, nil
}

// MessageCreate will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func MessageCreate(sess *discordgo.Session, msg *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example, but it's a good practice.
	if msg.Author.ID == sess.State.User.ID {
		return
	}
	if msg.Author.Bot {
		return
	}
	vcs, err := sess.State.VoiceState(msg.GuildID, sess.State.User.ID)
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

	if !strings.HasPrefix(msg.Content, prefix) {
		if !(isJoined && msg.ChannelID == ceviord.pickedChannel) {
			return
		}

		err = rawSpeak(GetMsg(msg, sess))
		if err != nil {
			log.Println(err)
		}
		return
	}
	ceviord.isJoin = isJoined
	ceviord.dictController.SetGuildId(msg.GuildID)
	cmd, err := parseUserCmd(strings.TrimPrefix(msg.Content, prefix))
	if err != nil {
		log.Println(fmt.Errorf("error occured in user cmd parser `%w`", err))
		return
	}
	if err = cmd.handle(sess, msg); err != nil {
		log.Println(fmt.Errorf("error occured in cmd handler %T; `%w`", cmd, err))
	}
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

func SendEmbedMsg(embed *discordgo.MessageEmbed, session *discordgo.Session) error {
	if session == nil {
		return fmt.Errorf("discord go session is nil")
	}
	_, err := session.ChannelMessageSendEmbed(ceviord.pickedChannel, embed)
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
		return ""
	}
	return stringMax(rawMsg, strLenMax)
}
