package commands

import (
	"errors"

	"mkw_tt_event_discord_bot/internal"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"
	l "mkw_tt_event_discord_bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

const resetCommandMessageIdParameterName = "messageid"

var _ Command = (*ResetCommand)(nil)

type ResetCommand struct {
	timeTrialChannelsService *services.TimeTrialChannelsService
	standingsService         *services.StandingsService
	lorenziService           *services.Lorenzi
}

func NewResetCommand(
	standingsService *services.StandingsService,
	timeTrialChannelsService *services.TimeTrialChannelsService,
	lorenziService *services.Lorenzi) *ResetCommand {
	return &ResetCommand{
		standingsService:         standingsService,
		timeTrialChannelsService: timeTrialChannelsService,
		lorenziService:           lorenziService,
	}
}

func (*ResetCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "reset",
		Description: "Recalculate standings for current track from <messageid> onwards until now.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        resetCommandMessageIdParameterName,
				Description: "Message ID from which to reculculate the standings (excluding given ID)",
				Required:    true,
			},
		},
	}
}

func (cmd *ResetCommand) Execute(ctx *Context) error {
	data := ctx.Interaction.ApplicationCommandData()

	var messageId string

	for _, opt := range data.Options {
		if opt.Name == resetCommandMessageIdParameterName {
			messageId = opt.StringValue()
		}
	}

	if messageId == "" {
		return NewMissingArgumentsError(resetCommandMessageIdParameterName)
	}

	submissionsChannelId := cmd.timeTrialChannelsService.GetSubmissionsId()
	if submissionsChannelId == "" {
		return errors.New("Cannot reset as no submissions channel id")
	}

	msgs, err := ctx.Session.ChannelMessages(submissionsChannelId, 100, "", messageId, "")

	if err != nil {
		// NOTE: want to ensure that at least some messages can be retrieved before nuking standings
		err := ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not retrieve messages from messageId: " + messageId + ".\nNot resetting standings"},
		})
		if err != nil {
			l.Log.Error("Could not inform about not being able to retrieve messages")
		}
		return nil
	}

	err = ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Resetting onwards from <#" + messageId + ">.\n"},
	})
	if err != nil {
		return errors.New("Could not inform about ressetting")
	}

	cmd.standingsService.ClearTimes()

	for {
		var lastHandledMessageID string

		for _, msg := range msgs {
			lastHandledMessageID = msg.ID
			if msg.Attachments == nil || len(msg.Attachments) == 0 {
				continue
			}

			for _, a := range msg.Attachments {
				time, error := timetrial.TimeFromFileName(a.Filename)

				if error != nil {
					continue
				}

				time.UserID = msg.Author.ID
				time.UserDisplayName = msg.Author.DisplayName()

				cmd.standingsService.HandleNewTime(time)
			}
		}

		msgs, err = ctx.Session.ChannelMessages(submissionsChannelId, 100, "", lastHandledMessageID, "")

		if err != nil {
			break
		}

		if len(msgs) == 0 {
			break
		}
	}

	err = internal.RefreshTableMessage(ctx.Session, cmd.lorenziService, cmd.standingsService, cmd.timeTrialChannelsService, false)

	if err != nil {
		l.Log.Error("Could not refresh table message", err)
		return err
	}

	return nil
}
