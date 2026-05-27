package commands

import (
	"mkw_tt_event_discord_bot/internal/export"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"
	"mkw_tt_event_discord_bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

var _ Command = (*NewSeasonCommand)(nil)

type NewSeasonCommand struct {
	timeTrialChannelService *services.TimeTrialChannelsService
	standingsService        *services.StandingsService
}

func NewNewSeasonCommand(timeTrialChannelService *services.TimeTrialChannelsService, standingsService *services.StandingsService) *NewSeasonCommand {
	return &NewSeasonCommand{
		timeTrialChannelService: timeTrialChannelService,
		standingsService:        standingsService,
	}
}

func (c *NewSeasonCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "newseason",
		Description: "End old season and start new",
	}
}

// New season means:
//
// 1. Locking the current TT forum channels
//
// 2. New TT forum containing channels:
//   - Discussion
//   - Custom track submissions
//   - Submissions
//   - Table
//
// 3. Reset times and exported config
func (cmd *NewSeasonCommand) Execute(ctx *Context) error {
	err := cmd.timeTrialChannelService.LockCurrentChannels(ctx.Session, ctx.GuildId)
	if err != nil {
		logger.Log.Warn("Could not lock current channels")
	}

	forumId, err := cmd.timeTrialChannelService.CreateNewChannels(ctx.Session, ctx.GuildId)
	if err != nil {
		return err
	}

	err = ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Started <#" + forumId + ">. Use /esn to initiate a track for submission. Use /tracklist to get available tracks with documented BKTs"},
	})
	if err != nil {
		return err
	}

	cmd.standingsService.ClearAll()
	exportChannels := cmd.timeTrialChannelService.ToExportChannels()

	return export.Save(
		export.Export{
			TimeTrialChannels: &exportChannels,
			Times:             []timetrial.Time{},
			BKT:               nil,
			FtreExpert:        nil,
			RRExpert:          nil,
			CurrentTrack:      "",
		})
}
