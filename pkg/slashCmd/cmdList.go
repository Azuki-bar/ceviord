package slashCmd

import (
	"github.com/bwmarrin/discordgo"
)

var slashCmdList = []*discordgo.ApplicationCommand{
	{
		Name:        joinCmdName,
		Description: "join voice actor",
	},
	{
		Name:        byeCmdName,
		Description: "voice actor disconnect",
	},
	{
		Name:        helpCmdName,
		Description: "get command reference",
	},
	{
		Name:        pingCmdName,
		Description: "check connection status",
	},
	{
		Name:        dictCmdName,
		Description: "manage dict records",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "add record",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "word",
						Description: "word",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "yomi",
						Description: "how to read that word",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "del",
				Description: "delete record",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "id",
						Description: "dictionary record id",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    true,
					},
				},
			},
			{
				Name:        "show",
				Description: "show records",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "length",
						Description: "specify number of records",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    false,
					},
				},
			},
			{
				Name:        "dump",
				Description: "dump all records",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			/* {
				Name:        "search",
				Description: "search record with effect",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "search string",
						Description: "search",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			}, */
		},
	},
	{
		Name:        changeCmdName,
		Description: "change voice actor",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "cast",
				Description: "cast name",
				Type:        discordgo.ApplicationCommandOptionString,
				Choices:     []*discordgo.ApplicationCommandOptionChoice{},
				Required:    true,
			},
		},
	},
}
