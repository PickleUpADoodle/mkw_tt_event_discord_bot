package messagehandlers

import "github.com/bwmarrin/discordgo"

type MessageHandler interface {
	Handle(s *discordgo.Session, m *discordgo.MessageCreate) error
}
