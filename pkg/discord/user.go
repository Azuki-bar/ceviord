package discord

import "github.com/bwmarrin/discordgo"

type User struct {
	*discordgo.User
	sess    *discordgo.Session
	guildId string
}

func NewUser(userId string, s *discordgo.Session, guildId string) (User, error) {
	u, err := s.User(userId)
	if err != nil {
		return User{}, err
	}
	return User{
		User:    u,
		sess:    s,
		guildId: guildId,
	}, nil
}

// ScreenName returns NickName if defined, and returns Username in else.
func (u User) ScreenName() (string, error) {
	m, err := u.sess.GuildMember(u.guildId, u.ID)
	if err != nil {
		return "", err
	}
	if m.Nick != "" {
		return m.Nick, nil
	}
	return u.Username, nil
}
