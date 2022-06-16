package handleCmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/azuki-bar/ceviord/pkg/dgvoice"
	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/azuki-bar/ceviord/pkg/replace"

	"github.com/bwmarrin/discordgo"
)

type Channel struct {
	pickedChannel  string
	VoiceConn      *discordgo.VoiceConnection
	currentParam   *Parameter
	guildId        string
	dictController replace.DbController
}

var logger = logging.NewLog(logging.INFO)

func (c Channel) isActorJoined(sess *discordgo.Session) (bool, error) {
	vcs, err := sess.State.VoiceState(c.guildId, sess.State.User.ID)
	if err != nil {
		logger.Log(logging.INFO, err)
		return false, err
	}
	return vcs.ChannelID != "", nil
}

type Channels map[string]*Channel

func (cs Channels) addChannel(c Channel, guildId string) {
	if _, ok := cs[guildId]; !ok {
		c.currentParam = &ceviord.param.Parameters[0]
		c.guildId = guildId
		c.dictController = ceviord.dictController
		cs[guildId] = &c
	}
}
func (cs Channels) getChannel(guildId string) (*Channel, error) {
	if c, ok := cs[guildId]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("channel not found")
}
func (cs Channels) isExistChannel(guildId string) bool {
	_, ok := cs[guildId]
	return ok
}

func (cs Channels) deleteChannel(guildId string) error {
	if cs.isExistChannel(guildId) {
		delete(cs, guildId)
		return nil
	}
	return fmt.Errorf("guild id not found")
}

type Ceviord struct {
	Channels       Channels
	cevioWav       CevioWav
	param          *Param
	Auth           *Auth
	mutex          sync.Mutex
	dictController replace.DbController
}

type Param struct {
	Parameters []Parameter `yaml:"parameters"`
}

type Auth struct {
	CeviordConn Conn `yaml:"conn"`
}

type DB struct {
	Name     string `yaml:"Name"`
	Addr     string `yaml:"Addr"`
	Port     int    `yaml:"Port"`
	Password string `yaml:"password"`
	User     string `yaml:"User"`
	Protocol string `yaml:"protocol"`
}

type Conn struct {
	Discord string `yaml:"discord"`
	Cevio   struct {
		Token    string `yaml:"cevioToken"`
		EndPoint string `yaml:"cevioEndPoint"`
	} `yaml:",inline"`
	DB DB `yaml:"db"`
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
const readTextMaxLen = 300

var tmpDir = filepath.Join(os.TempDir(), "ceviord")

var ceviord = Ceviord{
	Channels: Channels{},
	mutex:    sync.Mutex{},
}

func SetNewTalker(wav CevioWav)              { ceviord.cevioWav = wav }
func SetDbController(r replace.DbController) { ceviord.dictController = r }
func SetParam(param *Param)                  { ceviord.param = param }

func findJoinedVC(s *discordgo.Session, guildId, authorId string) *discordgo.Channel {
	st, err := s.GuildChannels(guildId)
	if err != nil {
		logger.Log(logging.INFO, err)
		return nil
	}
	vcs, err := s.State.VoiceState(guildId, authorId)
	if err != nil {
		logger.Log(logging.WARN, fmt.Errorf("find joinedVc err occurred `%w`", err))
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
		mainCmd = new(sasaraOld)
	case "bye":
		mainCmd = new(byeOld)
	case "dict":
		mainCmd = new(dictOld)
	case "change":
		mainCmd = new(changeOld)
	case "help", "man":
		mainCmd = new(helpOld)
	case "ping":
		mainCmd = new(pingOld)
	default:
		return nil, fmt.Errorf("unknown user cmd `%s` \n", rawCmd[0])
	}
	// parse sub command
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
	cev, err := ceviord.Channels.getChannel(msg.GuildID)
	if err != nil || cev == nil {
		//todo; チャンネルに入っていないときの挙動を定義
	}
	isJoined := false
	if cev != nil {
		isJoined, err = cev.isActorJoined(sess)
		if err != nil {
			logger.Log(logging.INFO, "Err occurred in actor joined detector")
			return
		}
	}
	if !strings.HasPrefix(msg.Content, prefix) && isJoined {
		if !(isJoined && msg.ChannelID == cev.pickedChannel) {
			return
		}
		err = rawSpeak(GetMsg(msg, sess), msg.GuildID, sess)
		if err != nil {
			logger.Log(logging.INFO, err)
		}
		return
	}
	if cev != nil { // already establish connection
		cev.dictController.SetGuildId(msg.GuildID)
	}
	cmd, err := parseUserCmd(strings.TrimPrefix(msg.Content, prefix))
	if err != nil {
		logger.Log(logging.DEBUG, fmt.Errorf("error occured in user cmd parser `%w`", err))
		return
	}
	if err = cmd.handle(sess, msg); err != nil {
		logger.Log(logging.WARN, fmt.Errorf("error occured in cmd handler %T; `%w`", cmd, err))
	}
}

func rawSpeak(text string, guildId string, sess *discordgo.Session) error {
	cev, err := ceviord.Channels.getChannel(guildId)
	isJoined, err := cev.isActorJoined(sess)
	if err != nil || !isJoined {
		return err
	}
	ceviord.cevioWav.ApplyEmotions(cev.currentParam)
	ceviord.mutex.Lock()
	defer ceviord.mutex.Unlock()
	buf := make([]byte, 16)
	_, err = rand.Read(buf)
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
	dgvoice.PlayAudioFile(cev.VoiceConn, fPath, make(chan bool))
	return nil
}

func SendMsg(msg string, session *discordgo.Session, guildId string) error {
	cev, err := ceviord.Channels.getChannel(guildId)
	isJoined, err := cev.isActorJoined(session)
	if err != nil || !isJoined {
		return err
	}
	// https://discord.com/developers/docs/resources/channel#create-message-jsonform-params
	if len([]rune(msg)) > 2000 {
		return fmt.Errorf("discord message send limitation error")
	} else if len([]rune(msg)) == 0 {
		return fmt.Errorf("message len is 0")
	}
	_, err = session.ChannelMessageSend(cev.pickedChannel, msg)
	return err
}

func SendEmbedMsg(embed *discordgo.MessageEmbed, session *discordgo.Session, guildId string) error {
	cev, err := ceviord.Channels.getChannel(guildId)
	isJoined, err := cev.isActorJoined(session)
	if err != nil || !isJoined {
		return err
	}
	if session == nil {
		return fmt.Errorf("discord go session is nil")
	}
	_, err = session.ChannelMessageSendEmbed(cev.pickedChannel, embed)
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
		logger.Log(logging.WARN, fmt.Errorf("replace mention failed `%w`", err))
		return ""
	}
	msg := []rune(name + "。" + replace.ApplySysDict(cont))

	cev, err := ceviord.Channels.getChannel(m.GuildID)
	cev.dictController.SetGuildId(m.GuildID)
	rawMsg, err := cev.dictController.ApplyUserDict(string(msg))
	if err != nil {
		logger.Log(logging.WARN, "apply user dict failed `%w`", err)
		return ""
	}
	return stringMax(rawMsg, readTextMaxLen)
}