package ceviord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type userMainCmd interface {
	parse(subCmds []string) error
	handle(sess *discordgo.Session, msg *discordgo.MessageCreate) error
}

type change struct {
	changeTo string
}

func (c *change) handle(_ *discordgo.Session, _ *discordgo.MessageCreate) error {
	for _, p := range ceviord.conf.Parameters {
		if c.changeTo == p.Name {
			ceviord.currentParam = &p
			ceviord.cevioWav.ApplyEmotions(ceviord.currentParam)
			if err := rawSpeak(fmt.Sprintf("パラメータを %s に変更しました。", p.Name)); err != nil {
				return fmt.Errorf("speaking about parameter setting: `%w`", err)
			}
		}
	}
	return nil
}

func (c *change) parse(cmds []string) error {
	if len(cmds) != 1 {
		return fmt.Errorf("apply commands are not correct")
	}
	c.changeTo = cmds[0]
	return nil
}

type sasara struct{}

func (_ *sasara) handle(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	if ceviord.isJoin {
		return fmt.Errorf("sasara is already joined\n")
	}
	vc := FindJoinedVC(sess, msg)
	if vc == nil {
		//todo fix err msg
		return fmt.Errorf("voice conn ")
	}
	var err error
	ceviord.VoiceConn, err = sess.ChannelVoiceJoin(msg.GuildID, vc.ID, false, false)
	if err != nil {
		log.Println(fmt.Errorf("joining: %w", err))
		return err
	}
	ceviord.pickedChannel = msg.ChannelID
	return nil
}
func (s *sasara) parse(_ []string) error { return nil }

type bye struct{}

func (b *bye) parse(_ []string) error { return nil }
func (b *bye) handle(_ *discordgo.Session, _ *discordgo.MessageCreate) error {
	//TODO implement me
	if !ceviord.isJoin || ceviord.VoiceConn == nil {
		return fmt.Errorf("ceviord is already disconnected\n")
	}
	defer func() {
		if ceviord.VoiceConn != nil {
			ceviord.VoiceConn.Close()
		}
	}()
	var err error
	err = ceviord.VoiceConn.Speaking(false)
	if err != nil {
		log.Println(fmt.Errorf("speaking falsing: %w", err))
	}
	err = ceviord.VoiceConn.Disconnect()
	if err != nil {
		log.Println(fmt.Errorf("disconnecting: %w", err))
	}
	return nil
}

type dict struct {
	sub  string
	opts []string
}

func (d *dict) handle(sess *discordgo.Session, msg *discordgo.MessageCreate) error {
	//TODO implement me
	panic("implement me")
}

func (d *dict) parse(cmds []string) error {
	if len(cmds) < 3 {
		return fmt.Errorf("sub cmd are not satisfied. \n")
	}
	d.sub = cmds[0]
	d.opts = cmds[1:]
	return nil
}
