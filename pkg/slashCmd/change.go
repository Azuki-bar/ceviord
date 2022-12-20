package slashCmd

import (
	"fmt"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/bwmarrin/discordgo"
)

type change struct {
	changeTo string
}

func (c *change) handle(finish chan<- bool, s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, o := range i.ApplicationCommandData().Options {
		switch o.Name {
		case "cast":
			c.changeTo = o.StringValue()
		}
	}
	err := c.rawHandle(s, i)
	msg := fmt.Sprintf("successfully change cast to %s", c.changeTo)
	if err != nil {
		msg = err.Error()
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
	finish <- true
	if err != nil {
		ceviord.Logger.Log(logging.WARN, fmt.Errorf("change handler failed. err is `%w`", err))
	}
}
func (c *change) rawHandle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	cev, err := ceviord.Cache.Channels.GetChannel(i.GuildID)
	if err != nil {
		return fmt.Errorf("voice connection not found")
	}
	isJoin, err := cev.IsActorJoined(s)
	if err != nil || !isJoin {
		return fmt.Errorf("voice connection not found")
	}
	for _, p := range ceviord.Cache.Param.Parameters {
		if c.changeTo == p.Name {
			cev.CurrentParam = &p
			if err := ceviord.RawSpeak(fmt.Sprintf("パラメータを %s に変更しました。", p.Name), i.GuildID, s); err != nil {
				return fmt.Errorf("speaking about parameter setting: `%w`", err)
			}
		}
	}
	return nil
}
