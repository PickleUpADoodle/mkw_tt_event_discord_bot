package export

import (
	"encoding/json"
	"os"

	"mkw_tt_event_discord_bot/internal/timetrial"
	l "mkw_tt_event_discord_bot/pkg/logger"
)

const filePath = "standings.json"

type Export struct {
	TimeTrialChannels *TimeTrialChannelsExport `json:"time_trial_channels"`
	Times             []timetrial.Time         `json:"times"`
	BKT               *timetrial.Time          `json:"bkt"`
	FtreExpert        *timetrial.Time          `json:"ftre_expert"`
	RRExpert          *timetrial.Time          `json:"rr_expert"`
	CurrentTrack      string                   `json:"current_track"`
}

type TimeTrialChannelsExport struct {
	Table                    TableChannelExport `json:"table"`
	SubmissionsId            string             `json:"submissions_id"`
	CustomTrackSubmissionsId string             `json:"ct_submissions_id"`
	DiscussionId             string             `json:"discussion_id"`
}

type TableChannelExport struct {
	ChannelId          string `json:"channel_idd"`
	StandingsMessageId string `json:"standings_message_id"`
}

func Save(export Export) error {
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func Load() *Export {
	standings, err := os.ReadFile(filePath)
	if err != nil {
		l.Log.Warn("Could not get standings", "error", err)
		// NOTE file may not exists, this is fine
		return nil
	}

	var data Export
	err = json.Unmarshal(standings, &data)
	if err != nil {
		panic(err)
	}

	return &data
}
