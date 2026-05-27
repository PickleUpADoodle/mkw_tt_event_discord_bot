package commands

import (
	"mkw_tt_event_discord_bot/internal/services"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

var _ Command = (*TrackListCommand)(nil)

type TrackListCommand struct {
	bktService *services.BKTService
}

func NewTrackListCommand(bktService *services.BKTService) *TrackListCommand {
	return &TrackListCommand{
		bktService: bktService,
	}
}

func (c *TrackListCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "tracklist",
		Description: "Expose all tracks with registered BKTs",
	}
}

func (cmd *TrackListCommand) Execute(ctx *Context) error {
	var sb strings.Builder

	sb.WriteString("Tracklist with BKTs:")
	sb.WriteString("\n")

	trackNames := cmd.bktService.GetAllBKTTrackNames()

	// NOTE: as max message length is 2000, multiple messages are sent
	const DiscordMaxMessageLength = 1500

	var count int = 0
	sentInitialResponse := false
	for _, trackName := range trackNames {
		trackLen := utf8.RuneCountInString(trackName)

		if count+trackLen >= DiscordMaxMessageLength {
			if !sentInitialResponse {
				err := ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: sb.String(),
					},
				})
				if err != nil {
					return err
				}
				sentInitialResponse = true
			} else {
				_, err := ctx.Session.FollowupMessageCreate(
					ctx.Interaction.Interaction,
					false, &discordgo.WebhookParams{
						Content: sb.String()})
				if err != nil {
					return err
				}
			}

			count = 0
			sb = strings.Builder{}
		}

		// Assuming a singular trackname length is never > 2000
		sb.WriteString(trackName)
		sb.WriteString("\n")
		count += trackLen
	}

	if count > 0 {
		_, err := ctx.Session.ChannelMessageSend(ctx.Interaction.ChannelID, sb.String())
		if err != nil {
			return err
		}
	}
	return nil
}
