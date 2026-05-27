package commands

import (
	"mkw_tt_event_discord_bot/internal"
	"mkw_tt_event_discord_bot/internal/services"

	"github.com/bwmarrin/discordgo"
)

var _ Command = (*PurgeCommand)(nil)

const purgeCommandDiscordIdParameterName = "userid"

type PurgeCommand struct {
	standingsService         *services.StandingsService
	lorenziService           *services.Lorenzi
	timeTrialChannelsService *services.TimeTrialChannelsService
}

func NewPurgeCommand(
	standingsService *services.StandingsService,
	timeTrialChannelsService *services.TimeTrialChannelsService,
	lorenziService *services.Lorenzi) *PurgeCommand {
	return &PurgeCommand{
		standingsService:         standingsService,
		timeTrialChannelsService: timeTrialChannelsService,
		lorenziService:           lorenziService,
	}
}

func (*PurgeCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "purge",
		Description: "Removes a time from standings",
		// TODO: Implement finding most recent time by player
		//Description: "Removes a time from standings. Then tries to find the previous most recent time sent by player", For now either ask person to resubmit or change standings.json export and restart bot
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        purgeCommandDiscordIdParameterName,
				Description: "Discord user id of the player to remove time from",
				Required:    true,
			},
		},
	}
}

func (cmd *PurgeCommand) Execute(ctx *Context) error {
	data := ctx.Interaction.ApplicationCommandData()

	var playerId string

	for _, opt := range data.Options {
		switch opt.Name {
		case purgeCommandDiscordIdParameterName:
			playerId = opt.StringValue()
			break
		}
	}

	if playerId == "" {
		return NewMissingArgumentsError(purgeCommandDiscordIdParameterName)
	}

	cmd.standingsService.RemoveTime(services.UserID(playerId))

	err := internal.RefreshTableMessage(ctx.Session, cmd.lorenziService, cmd.standingsService, cmd.timeTrialChannelsService, false)

	if err != nil {
		return err
	}

	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Removed standing for <@" + playerId + ">.\n\nSubmit a new time in <#" + cmd.timeTrialChannelsService.GetSubmissionsId() + ">!"},
	})

	// submissionsChannelId := cmd.timeTrialChannelsService.Channels.SubmissionsId
	//
	// // TODO: should keep track of date when esn was used
	// msgs, err := ctx.Session.ChannelMessages(submissionsChannelId, 100, "", "", "")
	//
	// for _, msg := range msgs {
	// 	if msg.Author.ID != playerId ||
	// 		msg.Attachments == nil || len(msg.Attachments) == 0 {
	// 		continue
	// 	}
	//
	// 	for _, a := range msg.Attachments {
	// 		time, err := timetrial.TimeFromFileName(a.Filename)
	//
	// 		if err != nil {
	// 			continue
	// 		}
	// 		if !time.Equals(*removedTime) {
	//
	// 		}
	// 	}
	// }
	//
	// if err != nil {
	//
	// }
}
