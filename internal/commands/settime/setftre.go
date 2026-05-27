package settime

import (
	"mkw_tt_event_discord_bot/internal/commands"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"

	"github.com/bwmarrin/discordgo"
)

var _ commands.Command = (*SetFtreCommand)(nil)

type SetFtreCommand struct {
	standingsService        *services.StandingsService
	lorenziService          *services.Lorenzi
	timeTrialChannelService *services.TimeTrialChannelsService
}

func NewSetFtreCommand(
	standingsService *services.StandingsService,
	lorenziService *services.Lorenzi,
	timeTrialChannelService *services.TimeTrialChannelsService,
) *SetFtreCommand {
	return &SetFtreCommand{
		standingsService:        standingsService,
		lorenziService:          lorenziService,
		timeTrialChannelService: timeTrialChannelService,
	}
}

func (c *SetFtreCommand) Definition() *discordgo.ApplicationCommand {
	return definition("setftre", "Set the Ftre ghost time to beat for the current track")
}

func (cmd *SetFtreCommand) Execute(ctx *commands.Context) error {
	return execute(ctx, func(time timetrial.Time) {
		cmd.standingsService.SetFtreExpertTime(&time)
	}, "Ftre expert time", cmd.lorenziService, cmd.standingsService, cmd.timeTrialChannelService)
}
