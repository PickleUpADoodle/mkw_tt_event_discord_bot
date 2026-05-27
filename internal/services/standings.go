package services

import (
	"sort"
	"strings"
	"sync"

	"mkw_tt_event_discord_bot/internal/timetrial"
)

type UserID string

type StandingsService struct {
	timesMu sync.Mutex
	Times   map[UserID]timetrial.Time

	bktMu   sync.Mutex
	bktTime *timetrial.Time

	ftreExpertTimeMu sync.Mutex
	ftreExpertTime   *timetrial.Time

	rrExpertTimeMu sync.Mutex
	rrExpertTime   *timetrial.Time

	currentTrackMu sync.Mutex
	currentTrack   string
}

func (s *StandingsService) GetBktTime() *timetrial.Time {
	s.bktMu.Lock()
	defer s.bktMu.Unlock()
	if s.bktTime == nil {
		return nil
	}
	copy := *s.bktTime
	return &copy
}

func (s *StandingsService) SetBktTime(time *timetrial.Time) {
	s.bktMu.Lock()
	defer s.bktMu.Unlock()
	if time == nil {
		s.bktTime = nil
		return
	}

	copy := *time
	s.bktTime = &copy
}

func (s *StandingsService) GetFtreExpertTime() *timetrial.Time {
	s.ftreExpertTimeMu.Lock()
	defer s.ftreExpertTimeMu.Unlock()
	if s.ftreExpertTime == nil {
		return nil
	}

	copy := *s.ftreExpertTime
	return &copy
}

func (s *StandingsService) SetFtreExpertTime(time *timetrial.Time) {
	s.ftreExpertTimeMu.Lock()
	defer s.ftreExpertTimeMu.Unlock()
	if time == nil {
		s.ftreExpertTime = nil
		return
	}

	copy := *time
	s.ftreExpertTime = &copy
}

func (s *StandingsService) GetRRExpertTime() *timetrial.Time {
	s.rrExpertTimeMu.Lock()
	defer s.rrExpertTimeMu.Unlock()
	if s.rrExpertTime == nil {
		return nil
	}

	copy := *s.rrExpertTime
	return &copy
}

func (s *StandingsService) SetRRExpertTime(time *timetrial.Time) {
	s.rrExpertTimeMu.Lock()
	defer s.rrExpertTimeMu.Unlock()
	if time == nil {
		s.rrExpertTime = nil
		return
	}

	copy := *time
	s.rrExpertTime = &copy
}

func (s *StandingsService) GetCurrentTrack() string {
	s.currentTrackMu.Lock()
	defer s.currentTrackMu.Unlock()
	return s.currentTrack
}

func (s *StandingsService) SetCurrentTrack(currentTrack string) {
	s.currentTrackMu.Lock()
	defer s.currentTrackMu.Unlock()
	s.currentTrack = currentTrack
}

func NewStandings(
	times []timetrial.Time,
	currentTrack string,
	RRExpert *timetrial.Time,
	Bkt *timetrial.Time,
	ftreExpert *timetrial.Time) *StandingsService {
	s := &StandingsService{
		Times:          make(map[UserID]timetrial.Time),
		ftreExpertTime: ftreExpert,
		rrExpertTime:   RRExpert,
		bktTime:        Bkt,
		currentTrack:   currentTrack,
	}

	if times != nil {
		for _, t := range times {
			s.Times[UserID(t.UserID)] = t
		}
	}

	return s
}

func (s *StandingsService) ClearAll() {
	s.timesMu.Lock()
	s.Times = make(map[UserID]timetrial.Time)
	s.timesMu.Unlock()

	s.bktMu.Lock()
	s.bktTime = nil
	s.bktMu.Unlock()

	s.ftreExpertTimeMu.Lock()
	s.ftreExpertTime = nil
	s.ftreExpertTimeMu.Unlock()

	s.rrExpertTimeMu.Lock()
	s.rrExpertTime = nil
	s.rrExpertTimeMu.Unlock()
}

func (s *StandingsService) ClearTimes() {
	s.timesMu.Lock()
	defer s.timesMu.Unlock()

	s.Times = make(map[UserID]timetrial.Time)
}

func (s *StandingsService) RemoveTime(userID UserID) *timetrial.Time {
	s.timesMu.Lock()
	defer s.timesMu.Unlock()

	if v, ok := s.Times[userID]; ok {
		delete(s.Times, userID)
		return &v
	}
	return nil
}

func (s *StandingsService) HandleNewTime(newTime timetrial.Time) (isNewTimeFaster bool) {
	s.timesMu.Lock()
	defer s.timesMu.Unlock()

	if currentTime, ok := s.Times[UserID(newTime.UserID)]; ok {
		if currentTime.IsFasterThan(newTime) {
			return false
		}
	}

	s.Times[UserID(newTime.UserID)] = newTime
	return true
}

func (s *StandingsService) GenerateStandings() (string, []PlayerStanding) {
	header := s.generateStandingsHeader()
	standingsStr, standings := s.generateStandingsBody(header)
	return standingsStr, standings
}

func (s *StandingsService) generateStandingsHeader() *strings.Builder {
	var sb strings.Builder

	bktTime := s.GetBktTime()
	if bktTime == nil {
		sb.WriteString("BKT: Unknown (use /setbkt)\n\n")
	} else {
		sb.WriteString("BKT: ")
		sb.WriteString(bktTime.String())
		sb.WriteString(" (use /setbkt)\n")
	}

	ftreExpertTime := s.GetFtreExpertTime()
	if ftreExpertTime == nil {
		sb.WriteString("Ftre expert time: Unknown (use /setftre)\n")
	} else {
		sb.WriteString("Ftre expert time: ")
		sb.WriteString(ftreExpertTime.String())
		sb.WriteString("\n")
	}

	rrExpertTime := s.GetRRExpertTime()
	if rrExpertTime == nil {
		sb.WriteString("RR expert time: Unknown (use /setrrexpert)\n\n")
	} else {
		sb.WriteString("RR expert time: ")
		sb.WriteString(rrExpertTime.String())
		sb.WriteString("\n\n")
	}

	return &sb
}

// b. Scoring is calculated as follows:
//
//	i. Points are awarded linearly.
//	   • If 5 people submit times:
//	     1st = 5 points
//	     2nd = 4 points
//	     etc.
//	ii. Beat the Retro Rewind staff ghost > +1 point
//	iii. Beat the "Ftre expert ghost" > +3 points
//	iv. Beat or tie the BKT (set before the track start date) > Score is doubled
func (s *StandingsService) generateStandingsBody(sb *strings.Builder) (string, []PlayerStanding) {
	sb.WriteString("CurrentStandings:\n")
	if len(s.Times) == 0 {
		sb.WriteString("No submissions")
		return sb.String(), []PlayerStanding{}
	}

	sortedTimes := s.GetSortedTimes()

	standings := make([]PlayerStanding, 0, len(s.Times))

	for i, t := range sortedTimes {
		ps := PlayerStanding{}

		basePts := len(s.Times) - i

		beatRRPts := 0
		rrExpertTime := s.GetRRExpertTime()
		if rrExpertTime != nil && t.IsFasterThan(*rrExpertTime) {
			beatRRPts = 1
		}

		beatFtrePts := 0
		ftreExpertTime := s.GetFtreExpertTime()
		if ftreExpertTime != nil && t.IsFasterThan(*ftreExpertTime) {
			beatFtrePts = 3
		}

		beatBKTPts := 0
		bktTime := s.GetBktTime()
		if bktTime != nil && t.IsFasterThan(*bktTime) {
			beatBKTPts = 2 * (basePts + beatRRPts + beatFtrePts)
		}

		p := [4]int{basePts, beatRRPts, beatFtrePts, beatBKTPts}

		ps.Points = p
		ps.Name = t.UserDisplayName
		standings = append(standings, ps)

		sb.WriteString(t.String())
		sb.WriteString(" ")
		sb.WriteString(ps.Name)
		sb.WriteString("\n")
	}

	return sb.String(), standings
}

func (s *StandingsService) GetSortedTimes() []timetrial.Time {
	s.timesMu.Lock()
	defer s.timesMu.Unlock()

	sortedTimes := make([]timetrial.Time, 0, len(s.Times))
	for _, v := range s.Times {
		sortedTimes = append(sortedTimes, v)
	}

	sort.Slice(sortedTimes, func(i, j int) bool {
		return sortedTimes[i].IsFasterThan(sortedTimes[j])
	})

	return sortedTimes
}
