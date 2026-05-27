package settime

import (
	"mkw_tt_event_discord_bot/internal/commands"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"

	"github.com/bwmarrin/discordgo"
)

var _ commands.Command = (*SetBktCommand)(nil)

type SetBktCommand struct {
	standingsService        *services.StandingsService
	lorenziService          *services.Lorenzi
	timeTrialChannelService *services.TimeTrialChannelsService
}

func NewSetBktCommand(
	standingsService *services.StandingsService,
	lorenziService *services.Lorenzi,
	timeTrialChannelService *services.TimeTrialChannelsService,
) *SetBktCommand {
	return &SetBktCommand{
		standingsService:        standingsService,
		lorenziService:          lorenziService,
		timeTrialChannelService: timeTrialChannelService,
	}
}

func (cmd *SetBktCommand) Definition() *discordgo.ApplicationCommand {
	return definition("setbkt", "Set the BKT expert ghost time to beat for the current track")
}

func (cmd *SetBktCommand) Execute(ctx *commands.Context) error {
	return execute(ctx, func(time timetrial.Time) {
		cmd.standingsService.SetBktTime(&time)
	}, "BKT time", cmd.lorenziService, cmd.standingsService, cmd.timeTrialChannelService)
}
