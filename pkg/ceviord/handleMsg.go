package ceviord

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/azuki-bar/ceviord/pkg/dgvoice"
	"github.com/azuki-bar/ceviord/pkg/replace"
	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
)

type Channel struct {
	PickedChannel  string
	VoiceConn      *discordgo.VoiceConnection
	CurrentParam   *Parameter
	guildID        string
	DictController replace.DbController
	logger         *zap.Logger
}

func (c Channel) IsActorJoined(sess *discordgo.Session) (bool, error) {
	vcs, err := sess.State.VoiceState(c.guildID, sess.State.User.ID)
	if err != nil {
		c.logger.Info("actor join error", zap.Error(err))
		return false, err
	}
	return vcs.ChannelID != "", nil
}

type Channels map[string]*Channel

func (cs Channels) AddChannel(c Channel, guildID string) {
	if _, ok := cs[guildID]; !ok {
		c.CurrentParam = &Cache.Param.Parameters[0]
		c.guildID = guildID
		c.DictController = Cache.dictController
		cs[guildID] = &c
	}
}
func (cs Channels) GetChannel(guildID string) (*Channel, error) {
	if c, ok := cs[guildID]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("voice actor connected channel is not found")
}
func (cs Channels) IsExistChannel(guildID string) bool {
	_, ok := cs[guildID]
	return ok
}

func (cs Channels) DeleteChannel(guildID string) error {
	if cs.IsExistChannel(guildID) {
		delete(cs, guildID)
		return nil
	}
	return fmt.Errorf("guild id not found")
}

type Ceviord struct {
	Channels       Channels
	cevioWav       CevioWav
	Param          *Param
	Auth           *Auth
	mutex          sync.Mutex
	dictController replace.DbController
	Logger         *zap.Logger
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

var tmpDir = filepath.Join(os.TempDir(), "ceviord")

var Cache = Ceviord{
	Channels: Channels{},
	mutex:    sync.Mutex{},
}

func SetNewTalker(wav CevioWav)              { Cache.cevioWav = wav }
func SetDbController(r replace.DbController) { Cache.dictController = r }
func SetParam(param *Param)                  { Cache.Param = param }
func SetLogger(logger *zap.Logger)           { Cache.Logger = logger }

func FindJoinedVC(s *discordgo.Session, guildID, authorID string) *discordgo.Channel {
	st, err := s.GuildChannels(guildID)
	if err != nil {
		Cache.Logger.Error("get guild channels failed", zap.Error(err), zap.String("guildID", guildID), zap.String("authorID", authorID))
		return nil
	}
	vcs, err := s.State.VoiceState(guildID, authorID)
	if err != nil {
		Cache.Logger.Warn("get voice state failed", zap.Error(err), zap.String("guildID", guildID), zap.String("authorID", authorID))
		return nil
	}
	for _, c := range st {
		if c.Type == discordgo.ChannelTypeGuildVoice {
			if c.ID == vcs.ChannelID {
				return c
			}
		}
	}
	return nil
}

// MessageCreate will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func MessageCreate(sess *discordgo.Session, msg *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example, but it's a good practice.
	if msg.Author.ID == sess.State.User.ID {
		Cache.Logger.Debug("handle sasara's msg")
		return
	}
	if msg.Author.Bot {
		Cache.Logger.Debug("user is bot")
		return
	}
	cev, err := Cache.Channels.GetChannel(msg.GuildID)
	if err != nil || cev == nil {
		// TODO: チャンネルに入っていないときの挙動を定義
		log.Println(err)
		return
	}
	isJoined := false
	if cev != nil {
		isJoined, err = cev.IsActorJoined(sess)
		if err != nil {
			Cache.Logger.Info("error occured in actor joined detector", zap.Error(err))
			return
		}
	}
	if !strings.HasPrefix(msg.Content, prefix) && isJoined {
		if !(isJoined && msg.ChannelID == cev.PickedChannel) {
			return
		}
		replacedMsg := GetMsg(msg, sess)
		if len(replacedMsg) != 0 {
			err = RawSpeak(replacedMsg, msg.GuildID, sess)
			if err != nil {
				Cache.Logger.Info("speaking failed", zap.Error(err))
			}
			return
		}
	}
}

func RawSpeak(text string, guildID string, sess *discordgo.Session) error {
	if len(text) == 0 {
		return fmt.Errorf("text length is 0")
	}
	cev, err := Cache.Channels.GetChannel(guildID)
	if cev == nil {
		return fmt.Errorf("get channel failed")
	}
	if err != nil {
		return err
	}
	isJoined, err := cev.IsActorJoined(sess)
	if err != nil || !isJoined {
		return err
	}
	err = Cache.cevioWav.ApplyEmotions(cev.CurrentParam)
	if err != nil {
		return err
	}
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
	err = Cache.cevioWav.OutputWaveToFile(text, fPath)
	defer os.Remove(fPath)
	if err != nil {
		return fmt.Errorf("outputting: %w", err)
	}
	Cache.mutex.Lock()
	defer Cache.mutex.Unlock()
	dgvoice.PlayAudioFile(Cache.Logger, cev.VoiceConn, fPath, make(chan bool))
	return nil
}

func SendMsg(msg string, session *discordgo.Session, guildID string) error {
	cev, err := Cache.Channels.GetChannel(guildID)
	if err != nil {
		return err
	}
	isJoined, err := cev.IsActorJoined(session)
	if err != nil || !isJoined {
		return err
	}
	// https://discord.com/developers/docs/resources/channel#create-message-jsonform-params
	if len([]rune(msg)) > DiscordPostLenLimit {
		return fmt.Errorf("discord message send limitation error")
	} else if len([]rune(msg)) == 0 {
		return fmt.Errorf("message len is 0")
	}
	_, err = session.ChannelMessageSend(cev.PickedChannel, msg)
	return err
}

func SendEmbedMsg(embed *discordgo.MessageEmbed, session *discordgo.Session, guildID string) error {
	cev, err := Cache.Channels.GetChannel(guildID)
	if cev == nil {
		return err
	}
	isJoined, err := cev.IsActorJoined(session)
	if err != nil || !isJoined {
		return err
	}
	if session == nil {
		return fmt.Errorf("discord go session is nil")
	}
	_, err = session.ChannelMessageSendEmbed(cev.PickedChannel, embed)
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
		Cache.Logger.Warn("replace mention failed", zap.Error(err), zap.Any("message", m))
		return ""
	}
	// issue #84
	if regexp.MustCompile(`^!.+$`).ReplaceAllString(cont, "") == "" {
		return ""
	}
	msg := []rune(name + "。" + replace.ApplySysDict(cont))

	cev, err := Cache.Channels.GetChannel(m.GuildID)
	if err != nil {
		return ""
	}
	cev.DictController.SetGuildID(m.GuildID)
	rawMsg, err := cev.DictController.ApplyUserDict(string(msg))
	if err != nil {
		Cache.Logger.Warn("apply user dict failed", zap.Error(err), zap.String("msg", rawMsg))
		return ""
	}
	return stringMax(rawMsg, strLenMax)
}

func stringMax(msg string, max int) string {
	lenMsg := len([]rune(msg))
	if lenMsg > max {
		return string([]rune(msg)[0:max])
	}
	return msg
}
