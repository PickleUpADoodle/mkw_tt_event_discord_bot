package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"mkw_tt_event_discord_bot/pkg/httpclient"
	l "mkw_tt_event_discord_bot/pkg/logger"
)

type Lorenzi struct {
	client *httpclient.Client
}

type PlayerStanding struct {
	Name   string
	Points [4]int
}

type createTableRequest struct {
	Data  string `json:"data"`
	Style *style `json:"style,omitempty"`
}

type style struct {
	Title *string `json:"title,omitempty"`
}

func NewLorenzi(client *httpclient.Client) *Lorenzi {
	return &Lorenzi{
		client: client,
	}
}

// See Lorenzi API doc: https://gb2.hlorenzi.com/apidocs
func (lrz *Lorenzi) GetTableBase64PNG(title string, playerStandings []PlayerStanding) ([]byte, error) {
	if playerStandings == nil {
		return nil, errors.New("playerStandings was null")
	}

	// NOTE: generate table post data
	var b strings.Builder

	for _, ps := range playerStandings {
		b.WriteString(ps.Name)
		b.WriteString(" ")

		ptLen := len(ps.Points)

		for i, p := range ps.Points {
			b.WriteString(strconv.Itoa(p))
			if i < ptLen-1 {
				b.WriteString("|")
			}
		}

		b.WriteString("\n")
	}

	//NOTE: retrieve the png
	postData := createTableRequest{
		Style: &style{
			Title: &title,
		},
		Data: b.String(),
	}

	postDataJson, err := json.Marshal(postData)

	if err != nil {
		fmt.Println("Could not parse json data")
		return nil, err
	}

	l.Log.Info("", "postData", postDataJson)

	res, httpCode, err := lrz.client.DoRequest(
		context.Background(),
		"POST",
		"https://gb2.hlorenzi.com/table.png",
		[]byte(postDataJson),
		map[string]string{
			"Content-Type": "application/json",
		},
	)

	if err != nil {
		l.Log.Error("Could not retrieve Lorenzi data", "httpCode", httpCode, "res", res)
		return nil, err
	}

	return res, nil
}
