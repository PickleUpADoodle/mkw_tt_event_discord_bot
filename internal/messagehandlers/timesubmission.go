package messagehandlers

import (
	"mkw_tt_event_discord_bot/internal"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"
	l "mkw_tt_event_discord_bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

var _ MessageHandler = (*TimeSubmissionMessagehandler)(nil)

type TimeSubmissionMessagehandler struct {
	standingsService *services.StandingsService
	channelService   *services.TimeTrialChannelsService
	lorenziService   *services.Lorenzi
}

func NewTimeSubmissionMessageHandler(
	standingsService *services.StandingsService,
	channelService *services.TimeTrialChannelsService,
	lorenziService *services.Lorenzi,
) *TimeSubmissionMessagehandler {
	return &TimeSubmissionMessagehandler{
		standingsService: standingsService,
		channelService:   channelService,
		lorenziService:   lorenziService,
	}
}

func (mh *TimeSubmissionMessagehandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if m.Author.Bot {
		return nil
	}

	l.Log.Info("New msg detected")

	cid := mh.channelService.GetSubmissionsId()
	if cid != m.ChannelID {
		l.Log.Info("Not in correct channel. Required: " + cid)
		return nil
	}

	if mh.standingsService.GetCurrentTrack() == "" {
		l.Log.Info("No track is active")
		return nil
	}

	if m.Attachments == nil || len(m.Attachments) == 0 {
		l.Log.Info("New msg no atachments", len(m.Embeds))
		return nil
	}

	l.Log.Info("New msg has attachments", len(m.Embeds))

	for _, a := range m.Attachments {
		newTime, err := timetrial.TimeFromFileName(a.Filename)

		if err != nil {
			l.Log.Warn("Could not parse time from fileName", "err", err)
			continue
		}

		if m.Author == nil {
			l.Log.Info("author nil")
		} else {
			newTime.UserID = m.Author.ID
			newTime.UserDisplayName = m.Author.DisplayName()
		}

		l.Log.Info("", "time", newTime)

		isNewTimeFaster := mh.standingsService.HandleNewTime(newTime)
		if !isNewTimeFaster {
			return nil
		}
		l.Log.Info("time is faster")

		err = internal.RefreshTableMessage(s, mh.lorenziService, mh.standingsService, mh.channelService, false)
		if err != nil {
			l.Log.Error("Could not refresh table message ", err)
			return err
		}
	}

	l.Log.Info("Finished")
	return nil
}
