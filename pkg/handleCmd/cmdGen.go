package handleCmd

import (
	"fmt"
	"log"

	"github.com/azuki-bar/ceviord/pkg/logging"
	"github.com/bwmarrin/discordgo"
)

var (
	Cmds = []*discordgo.ApplicationCommand{
		{
			Name:        "join",
			Description: "sasara join",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Japanese: "キャストを現在入っているボイスチャンネルに招待します。",
			},
		},
		{
			Name:        "bye",
			Description: "good bye sasara",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Japanese: "キャストにボイスチャンネルから退出してもらいます。",
			},
		},
		{
			Name:        "help",
			Description: "post link to command reference page",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Japanese: "コマンドリファレンスへのリンクを投稿します。",
			},
		},
		/* {
			Name:        "dict",
			Description: "",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Japanese: "読み方を登録する単語辞書を操作します。",
			},
			Type: discordgo.ApplicationCommandType(discordgo.ApplicationCommandOptionSubCommandGroup),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "dictionary add",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "yomi",
							Description: "add yomi",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
						},
						{
							Name:        "kana",
							Description: "add kana",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				},
				{
					Name:        "delete",
					Description: "dictionary delete",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:     "id",
							Type:     discordgo.ApplicationCommandOptionInteger,
							Required: true,
						},
					},
				},
			},
		}, */
	}
)

func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h, err := parseCommands(i.ApplicationCommandData().Name)
	if err != nil {
		log.Println(err)
		return
	}
	h.handle(s, i)
}

type CommandHandler interface {
	handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func parseCommands(commandName string) (CommandHandler, error) {
	log.Printf("parse command `%s`\n", commandName)
	var h CommandHandler
	switch commandName {
	case "join":
		h = new(join)
	/* case "help":
	h = new(help) */
	case "leave":
		h = new(leave)
	default:
		return nil, fmt.Errorf("command not found")
	}
	return h, nil
}

type join struct{}

func (j *join) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := j.rawHandle(s, i)
	msg := j.successMsg(i.GuildLocale)
	_ = msg
	if err != nil {
		msg = j.errMsg(i.GuildLocale)
		logger.Log(logging.WARN, "error in join handler `%w`", err)
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	if err != nil {
		logger.Log(logging.WARN, "error in `join` interaction respond")
	}
}

func (_ *join) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Member == nil {
		return fmt.Errorf("member field is nil. so cannot detect user status.")
	}
	vc := findJoinedVC(s, i.GuildID, i.Member.User.ID)
	if vc == nil {
		return fmt.Errorf("voice connection not found")
	}
	if ceviord.Channels.isExistChannel(i.Member.GuildID) {
		c, err := ceviord.Channels.getChannel(i.Member.GuildID)
		if err != nil {
			return fmt.Errorf("some error occurred in user joined channgel searcher")
		}
		isJoin, err := c.isActorJoined(s)
		if err != nil || isJoin {
			return fmt.Errorf("sasara is already joined")
		}
	}

	log.Printf("guildId `%s` vc.ID `%s`\n", i.Member.GuildID, vc.ID)
	voiceConn, err := s.ChannelVoiceJoin(i.Member.GuildID, vc.ID, false, true)
	if err != nil {
		log.Println(fmt.Errorf("joining: %w", err))
		return err
	}
	ceviord.Channels.addChannel(
		Channel{pickedChannel: i.ChannelID, VoiceConn: voiceConn},
		i.Member.GuildID,
	)
	return nil
}

func (_ *join) successMsg(l *discordgo.Locale) string {
	switch *l {
	case discordgo.Japanese:
		return "ボイスチャンネルに参加しました！"
	default:
		return "successfully joined to Voice Channel!"
	}
}

func (_ *join) errMsg(l *discordgo.Locale) string {
	switch *l {
	case discordgo.Japanese:
		return "`join`コマンドハンドラで何かエラーが発生しました。"
	default:
		return "something error occured in `join` command handler."
	}
}

type leave struct{}

func (l *leave) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := l.rawHandle(s, i)
	msg := "leave successfully"
	if err != nil {
		msg = "something occured in leave handler"
		logger.Log(logging.WARN, err)
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	if err != nil {
		logger.Log(logging.WARN, "error in `leave` handler `%w`", err)
	}
}
func (_ *leave) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	cev, err := ceviord.Channels.getChannel(i.Message.GuildID)
	if err != nil || cev == nil {
		return fmt.Errorf("connection not found")
	}
	isJoin, err := cev.isActorJoined(s)
	if !isJoin || cev.VoiceConn == nil {
		return fmt.Errorf("handleCmd is already disconnected\n")
	}
	defer func() {
		if cev.VoiceConn != nil {
			cev.VoiceConn.Close()
			ceviord.Channels.deleteChannel(i.Message.GuildID)
		}
	}()
	err = cev.VoiceConn.Speaking(false)
	if err != nil {
		log.Println(fmt.Errorf("speaking falsing: %w", err))
	}
	err = cev.VoiceConn.Disconnect()
	if err != nil {
		log.Println(fmt.Errorf("disconnecting: %w", err))
	}
	return nil
}

type helpOld struct{}
