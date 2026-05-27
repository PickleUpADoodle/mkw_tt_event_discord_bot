package timetrial

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Time struct {
	UserID string
	// NOTE: not that happy storing this here...
	UserDisplayName string
	Minutes         int
	Seconds         int
	Milliseconds    int
}

// Returns whether self is faster than other
func (self Time) IsFasterThan(other Time) bool {
	return other.Minutes > self.Minutes ||
		(other.Minutes == self.Minutes && other.Seconds > self.Seconds) ||
		(other.Minutes == self.Minutes && other.Seconds == self.Seconds && other.Milliseconds > self.Milliseconds)
}

func (self Time) Equals(other Time) bool {
	return self.Minutes == other.Minutes &&
		self.Seconds == other.Seconds &&
		self.Milliseconds == other.Milliseconds
}

// Derive potential time from attachment. Expecting file with name: <digit>m<digit><digit>s<digit><digit><digit>.rkg
func TimeFromFileName(fileName string) (Time, error) {
	const rkgExtension = ".rkg"

	if !strings.HasSuffix(strings.ToLower(fileName), rkgExtension) {
		return Time{}, fmt.Errorf("FileName: '%s' had different extension than expected '%s'", fileName, rkgExtension)
	}

	fileLen := utf8.RuneCountInString(fileName)

	if fileLen < 4 {
		return Time{}, errors.New("time length was not 4")
	}

	timeRaw := fileName[:len(fileName)-4] // remove ".rkg"
	time, err := TimeFromString(timeRaw)
	if err != nil {
		return Time{}, fmt.Errorf("Could not parse time: '%s' to object", timeRaw)
	}

	return time, nil
}

// Convert time from string (e.g "1:23.333")
// only the digit indexes matter, supports any delimiters
func TimeFromString(time string) (Time, error) {
	if len(time) != 8 {
		return Time{}, errors.New("Time length " + strconv.Itoa(len(time)) + " != 9")
	}

	minRaw := time[:1]
	secRaw := time[2:4]
	msRaw := time[5:]

	min, err := strconv.Atoi(minRaw)
	if err != nil {
		return Time{}, err
	}

	sec, err := strconv.Atoi(secRaw)
	if err != nil {
		return Time{}, err
	}

	ms, err := strconv.Atoi(msRaw)
	if err != nil {
		return Time{}, err
	}

	return Time{
		Minutes:      min,
		Seconds:      sec,
		Milliseconds: ms,
	}, nil
}

// TODO: perhaps add display name? Could maybe derive it from the spreadsheet for BKT as well
func (t Time) String() string {
	return strconv.Itoa(t.Minutes) + ":" + strconv.Itoa(t.Seconds) + "." + strconv.Itoa(t.Milliseconds)
}
