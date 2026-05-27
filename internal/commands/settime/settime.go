package settime

import (
	"fmt"

	"mkw_tt_event_discord_bot/internal"
	"mkw_tt_event_discord_bot/internal/commands"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"
	l "mkw_tt_event_discord_bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

const setTimeCommandGhostFileParameterName = "ghostfile"
const setTimeCommandTimeStringParameterName = "timestring"

func definition(commandName string, commandDescription string) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: commandDescription,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        setTimeCommandGhostFileParameterName,
				Description: "*.rkg containing the time",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        setTimeCommandTimeStringParameterName,
				Description: "String that contains the time (format 'm:ss.mmm')",
				Required:    false,
			},
		},
	}

}

func execute(
	ctx *commands.Context,
	changeTimeFn func(newTime timetrial.Time),
	timeName string,
	lorenziService *services.Lorenzi,
	standingsService *services.StandingsService,
	timeTrialChannelService *services.TimeTrialChannelsService,
) error {
	data := ctx.Interaction.ApplicationCommandData()

	var ghostAttachmentId string
	var timeString string

	for _, opt := range data.Options {
		if opt.Name == setTimeCommandGhostFileParameterName {
			ghostAttachmentId = opt.Value.(string)
		}
		if opt.Name == setTimeCommandTimeStringParameterName {
			timeString = opt.StringValue()
		}
	}

	if ghostAttachmentId == "" && timeString == "" {
		return commands.NewMissingArgumentsError(setTimeCommandGhostFileParameterName, setTimeCommandTimeStringParameterName)
	}

	if ghostAttachmentId != "" && timeString != "" {
		timeString = ""
	}

	var time timetrial.Time
	var err error

	if ghostAttachmentId != "" {
		if timeString != "" {
			l.Log.Warn("Received both ghost and string. Trying to retrieve file from ghost")
		}

		a := data.Resolved.Attachments[ghostAttachmentId]

		if a == nil {
			return fmt.Errorf("Could not retrieve attachment with id: %s", ghostAttachmentId)
		}

		time, err = timetrial.TimeFromFileName(a.Filename)

		if err != nil {
			return fmt.Errorf("Could not retrieve attachment with id: %w", err)
		}

	} else {
		time, err = timetrial.TimeFromString(timeString)
		if err != nil {
			return fmt.Errorf("Could not retrieve time from: %w", err)
		}
	}

	changeTimeFn(time)

	err = internal.RefreshTableMessage(ctx.Session, lorenziService, standingsService, timeTrialChannelService, false)
	if err != nil {
		return err
	}

	err = ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully set " + timeName + " time to: " + time.String(),
		}})
	if err != nil {
		return err
	}

	return nil
}
