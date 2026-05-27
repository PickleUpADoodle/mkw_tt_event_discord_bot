package settime

import (
	"mkw_tt_event_discord_bot/internal/commands"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"

	"github.com/bwmarrin/discordgo"
)

var _ commands.Command = (*SetRRExpert)(nil)

type SetRRExpert struct {
	standingsService        *services.StandingsService
	lorenziService          *services.Lorenzi
	timeTrialChannelService *services.TimeTrialChannelsService
}

func NewSetRRExpertCommand(
	standingsService *services.StandingsService,
	lorenziService *services.Lorenzi,
	timeTrialChannelService *services.TimeTrialChannelsService,
) *SetRRExpert {
	return &SetRRExpert{
		standingsService:        standingsService,
		lorenziService:          lorenziService,
		timeTrialChannelService: timeTrialChannelService,
	}
}

func (c *SetRRExpert) Definition() *discordgo.ApplicationCommand {
	return definition("setrrexpert", "Set the RR expert ghost time to beat for the current track")
}

func (cmd *SetRRExpert) Execute(ctx *commands.Context) error {
	return execute(ctx, func(time timetrial.Time) {
		cmd.standingsService.SetRRExpertTime(&time)
	}, "RR expert time", cmd.lorenziService, cmd.standingsService, cmd.timeTrialChannelService)
}
