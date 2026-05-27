package commands

import "github.com/bwmarrin/discordgo"

type Command interface {
	Definition() *discordgo.ApplicationCommand
	Execute(ctx *Context) error
}
