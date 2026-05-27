package services

import (
	"context"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
	"sync"

	"mkw_tt_event_discord_bot/internal/timetrial"
	"mkw_tt_event_discord_bot/pkg/httpclient"
	l "mkw_tt_event_discord_bot/pkg/logger"
)

type TrackName string

type BKTService struct {
	mu     sync.Mutex
	client *httpclient.Client
	times  map[TrackName]timetrial.Time
}

func NewBKTService(client *httpclient.Client) *BKTService {
	bktService := &BKTService{
		client: client,
		times:  map[TrackName]timetrial.Time{},
	}

	bktService.RefreshBKTData()

	return bktService
}

// Refresh the BKT data stored in memory
func (b *BKTService) RefreshBKTData() error {
	// Retrieve BKT data using the Google spreadsheet (bit dubious)
	csvRes, httpCode, err := b.client.DoRequest(
		context.Background(),
		"GET",
		"https://docs.google.com/spreadsheets/d/1XkHTTuUR3_10-C7geVhJ9TtCb4Bz_gE19NysbGnUOZs/export?format=csv",
		nil,
		map[string]string{
			"Content-Type": "text/csv",
		},
	)

	if err != nil {
		return errors.Join(err, errors.New("httpCode: "+strconv.Itoa(httpCode)))
	}

	if len(csvRes) == 0 {
		return errors.New("csv was empty")
	}

	reader := csv.NewReader(strings.NewReader(string(csvRes)))

	// skip headers
	_, err = reader.Read()
	if err != nil {
		return err
	}
	_, err = reader.Read()
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for {
		record, err := reader.Read()
		if err != nil {
			break // includes io.EOF
		}

		if len(record) != 17 {
			return errors.New("Retrieved record length" +
				strconv.Itoa(len(record)) +
				" was not equal to expected")
		}

		trackName := record[0]
		timeRaw := record[2]

		time, err := timetrial.TimeFromString(timeRaw)

		if err != nil {
			l.Log.Warn("Could not parse time to object. Skipping",
				"time", timeRaw,
				"track", trackName)
			continue
		}

		b.times[TrackName(strings.ToLower(trackName))] = time
	}

	return nil
}

func (b *BKTService) GetBKT(trackName string) (timetrial.Time, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if time, ok := b.times[TrackName(strings.ToLower(trackName))]; ok {
		return time, nil
	} else {
		return timetrial.Time{}, errors.New("Cannot get BKT as could not find time")
	}
}

// Get all BKTs present in cache
func (b *BKTService) GetAllBKTTrackNames() []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	tracks := make([]string, 0, len(b.times))

	for trackName, _ := range b.times {
		tracks = append(tracks, string(trackName))
	}
	return tracks
}
