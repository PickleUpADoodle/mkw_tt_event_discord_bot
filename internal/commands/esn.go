package commands

import (
	"mkw_tt_event_discord_bot/internal"
	"mkw_tt_event_discord_bot/internal/services"
	l "mkw_tt_event_discord_bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

var _ Command = (*EsnCommand)(nil)

const esnCommandTrackParameterName = "track"

type EsnCommand struct {
	bktService              *services.BKTService
	standingsService        *services.StandingsService
	lorenziService          *services.Lorenzi
	timeTrialChannelService *services.TimeTrialChannelsService
}

func NewEsnCommand(
	bktService *services.BKTService,
	standingsService *services.StandingsService,
	lorenziService *services.Lorenzi,
	timeTrialChannelService *services.TimeTrialChannelsService) *EsnCommand {
	return &EsnCommand{
		bktService:              bktService,
		standingsService:        standingsService,
		lorenziService:          lorenziService,
		timeTrialChannelService: timeTrialChannelService,
	}
}

func (c *EsnCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "esn",
		Description: "Stop tracking the current track and start tracking a new track",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        esnCommandTrackParameterName,
				Description: "Title used in the table and to derive BKT. see /tracknames",
				Required:    true,
			},
		},
	}
}

func (cmd *EsnCommand) Execute(ctx *Context) error {
	data := ctx.Interaction.ApplicationCommandData()

	var track string

	for _, opt := range data.Options {
		if opt.Name == esnCommandTrackParameterName {
			track = opt.StringValue()
		}
	}

	if track == "" {
		return NewMissingArgumentsError(esnCommandTrackParameterName)
	}

	cmd.standingsService.SetCurrentTrack(track)

	bkt, err := cmd.bktService.GetBKT(track)

	if err != nil {
		l.Log.Warn("Could not find BKT", "track", track)

		err := ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not find BKT for track: " + track + ".\nNot allocating BKT points. Use /tracklist to view known track names"},
		})
		if err != nil {
			l.Log.Error("Could not inform about missing BKT")
		}
	}

	cmd.standingsService.SetBktTime(&bkt)
	cmd.standingsService.ClearTimes()

	err = internal.RefreshTableMessage(ctx.Session, cmd.lorenziService, cmd.standingsService, cmd.timeTrialChannelService, true)

	if err != nil {
		return err
	}

	err = ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Initialized standings for " + track + ".\nView the table in <#" + cmd.timeTrialChannelService.GetTableChannelId() + ">.\nSubmit times in <#" + cmd.timeTrialChannelService.GetSubmissionsId() + ">"},
	})
	if err != nil {
		return err
	}

	return nil
}
