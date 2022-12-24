package discord

import "github.com/bwmarrin/discordgo"

type User struct {
	*discordgo.User
	sess    *discordgo.Session
	guildID string
}

func NewUser(userID string, s *discordgo.Session, guildID string) (User, error) {
	u, err := s.User(userID)
	if err != nil {
		return User{}, err
	}
	return User{
		User:    u,
		sess:    s,
		guildID: guildID,
	}, nil
}

// ScreenName returns NickName if defined, and returns Username in else.
func (u User) ScreenName() (string, error) {
	m, err := u.sess.GuildMember(u.guildID, u.ID)
	if err != nil {
		return "", err
	}
	if m.Nick != "" {
		return m.Nick, nil
	}
	return u.Username, nil
}
