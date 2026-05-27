package main

import (
	"errors"
	"strings"

	"mkw_tt_event_discord_bot/internal/commands"
	"mkw_tt_event_discord_bot/internal/messagehandlers"
	"mkw_tt_event_discord_bot/internal/services"
	l "mkw_tt_event_discord_bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session         *discordgo.Session
	commands        map[string]commands.Command
	messageHandlers []messagehandlers.MessageHandler
	guildId         string
}

func NewBot(
	apiKey string,
	guildId string,
	resetCommands bool,
	timeTrialChannelService *services.TimeTrialChannelsService,
	cmds []commands.Command,
	handlers []messagehandlers.MessageHandler) (*Bot, error) {

	s, err := discordgo.New("Bot " + apiKey)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		session:  s,
		commands: make(map[string]commands.Command),
		guildId:  guildId,
	}

	for _, cmd := range cmds {
		b.commands[cmd.Definition().Name] = cmd
	}

	b.messageHandlers = handlers

	b.session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuilds

	err = b.session.Open()
	if err != nil {
		return nil, err
	}

	err = timeTrialChannelService.Setup(b.session, b.guildId)
	if err != nil {
		return nil, err
	}

	app, err := b.session.Application("@me")
	if err != nil {
		return nil, err
	}

	b.session.AddHandler(b.handleMessageCreate)
	b.session.AddHandler(b.handleInteractionCreate)

	appCmds, err := b.session.ApplicationCommands(app.ID, guildId)

	if err != nil {
		return nil, err
	}

	if len(appCmds) == 0 {
		l.Log.Info("No commands found")
	}

	if resetCommands {
		l.Log.Info("Deleting commands")
		for _, existingCmd := range appCmds {
			b.session.ApplicationCommandDelete(app.ID, guildId, existingCmd.ID)
			b.session.ApplicationCommandDelete(app.ID, "", existingCmd.ID)
		}
	}

	l.Log.Info("Adding missing commands")
	for _, cmd := range b.commands {
		d := cmd.Definition()
		if !doesCommandExist(d.Name, appCmds) {
			// NOTE: Not speciying Guild ID creates a global command
			_, err := b.session.ApplicationCommandCreate(app.ID, guildId, d)
			if err != nil {
				return nil, err
			}
			l.Log.Info("Created command: ", "cmd", d.Name)
		} else {
			l.Log.Info("Command already exists", "cmd", d.Name)
		}
	}

	// NOTE: for testing to remove TT channels
	//b.removeTTChannels()

	return b, nil
}

func (b *Bot) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name
	cmd, ok := b.commands[cmdName]

	if !ok {
		l.Log.Error("Could not find command", cmdName)
		return
	}

	l.Log.Info("Command invoked", "cmd", cmdName, "channelID", i.ChannelID)

	ctx := commands.NewContext(s, nil, i, b.guildId)

	var missingArgsError *commands.MissingArgumentsError

	err := cmd.Execute(ctx)

	if err == nil {
		return
	}

	l.Log.Error("Error during cmd: ", err)

	var errorMessage string

	switch {
	case errors.As(err, &missingArgsError):
		errorMessage = "Missing required parameter(s): " + strings.Join(missingArgsError.Arguments, ",")
		break
	default:
		errorMessage = "Internal server error"
		break
	}

	err = ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: errorMessage,
		},
	})
	if err != nil {
		l.Log.Error("", err)
	}
}

func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, handler := range b.messageHandlers {
		err := handler.Handle(s, m)

		if err != nil {
			l.Log.Error("Error during message handler: ", err)
		}
	}
}

func (b *Bot) Close() error {
	return b.session.Close()
}

func doesCommandExist(command string, existingCommands []*discordgo.ApplicationCommand) bool {
	if len(existingCommands) == 0 {
		return false
	}

	for _, cmd := range existingCommands {
		if cmd.Name == command {
			return true
		}
	}

	return false
}

func (b *Bot) removeTTChannels() {
	channels, err := b.session.GuildChannels(b.guildId)

	if err != nil {
		return
	}

	prefix := "month-"
	suffix := "-time-trials"

	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildForum &&
			strings.HasPrefix(ch.Name, prefix) &&
			strings.HasSuffix(ch.Name, suffix) {
			b.session.ChannelDelete(ch.ID)
		}
	}
}
