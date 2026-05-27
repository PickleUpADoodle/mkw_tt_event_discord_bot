package internal

import (
	"bytes"

	"mkw_tt_event_discord_bot/internal/export"
	"mkw_tt_event_discord_bot/internal/services"

	"github.com/bwmarrin/discordgo"
)

// NOTE:This functionality came up multiple times accross files... not sure which place fit best

func RefreshTableMessage(
	session *discordgo.Session,
	lorenziService *services.Lorenzi,
	standingsService *services.StandingsService,
	timeTrialChannelService *services.TimeTrialChannelsService,
	createNewTable bool) error {

	standingsStr, standings := standingsService.GenerateStandings()

	table, err := lorenziService.GetTableBase64PNG(standingsService.GetCurrentTrack(), standings)
	if err != nil {
		return err
	}

	if createNewTable {
		err = timeTrialChannelService.NewTableMessage(session, standingsStr,
			&discordgo.File{
				Name:   "table.png",
				Reader: bytes.NewReader(table),
			},
		)
		if err != nil {
			return err
		}
	} else {
		err = timeTrialChannelService.UpdateTableMessage(session, standingsStr,
			&discordgo.File{
				Name:   "table.png",
				Reader: bytes.NewReader(table),
			},
		)
		if err != nil {
			return err
		}
	}

	exportChannels := timeTrialChannelService.ToExportChannels()

	return export.Save(
		export.Export{
			TimeTrialChannels: &exportChannels,
			Times:             standingsService.GetSortedTimes(),
			BKT:               standingsService.GetBktTime(),
			FtreExpert:        standingsService.GetFtreExpertTime(),
			RRExpert:          standingsService.GetRRExpertTime(),
			CurrentTrack:      standingsService.GetCurrentTrack(),
		})
}
