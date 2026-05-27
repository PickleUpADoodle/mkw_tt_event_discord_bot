package commands

import (
	"github.com/bwmarrin/discordgo"
)

func NewContext(
	session *discordgo.Session,
	message *discordgo.MessageCreate,
	interaction *discordgo.InteractionCreate,
	guildId string) *Context {
	return &Context{
		Session:     session,
		Message:     message,
		Interaction: interaction,
		GuildId:     guildId,
	}
}

type Context struct {
	Session     *discordgo.Session
	Message     *discordgo.MessageCreate
	Interaction *discordgo.InteractionCreate
	GuildId     string
}
