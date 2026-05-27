package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"mkw_tt_event_discord_bot/internal/export"
	l "mkw_tt_event_discord_bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

const sevenDaysInMinutes = 24 * 7 * 60

type TimeTrialChannelsService struct {
	mu                         sync.Mutex
	Channels                   TimeTrialChannels
	timeTrialCategoryChannelId string
}

type TimeTrialChannels struct {
	table                    TableChannel
	submissionsId            string
	customTrackSubmissionsId string
	discussionId             string
}

type TableChannel struct {
	channelId          string
	standingsMessageId string
}

func (ttcs *TimeTrialChannelsService) ToExportChannels() export.TimeTrialChannelsExport {
	ttcs.mu.Lock()
	defer ttcs.mu.Unlock()

	return export.TimeTrialChannelsExport{
		Table:                    *ttcs.toExportTable(ttcs.Channels.table),
		SubmissionsId:            ttcs.Channels.submissionsId,
		CustomTrackSubmissionsId: ttcs.Channels.customTrackSubmissionsId,
		DiscussionId:             ttcs.Channels.discussionId,
	}
}

func FromExportChannels(export *export.TimeTrialChannelsExport) TimeTrialChannels {
	if export == nil {
		l.Log.Info("export is nil")
		return TimeTrialChannels{}
	}

	return TimeTrialChannels{
		table:                    parseFromExportTable(export.Table),
		submissionsId:            export.SubmissionsId,
		customTrackSubmissionsId: export.CustomTrackSubmissionsId,
		discussionId:             export.DiscussionId,
	}
}

func parseFromExportTable(exportTable export.TableChannelExport) TableChannel {
	return TableChannel{
		channelId:          exportTable.ChannelId,
		standingsMessageId: exportTable.StandingsMessageId,
	}
}

func (ttcs *TimeTrialChannelsService) toExportTable(table TableChannel) *export.TableChannelExport {
	return &export.TableChannelExport{
		ChannelId:          table.channelId,
		StandingsMessageId: table.standingsMessageId,
	}
}

func (ttcs *TimeTrialChannelsService) GetSubmissionsId() string {
	ttcs.mu.Lock()
	defer ttcs.mu.Unlock()
	return ttcs.Channels.submissionsId
}

func (ttcs *TimeTrialChannelsService) GetTableChannelId() string {
	ttcs.mu.Lock()
	defer ttcs.mu.Unlock()
	return ttcs.Channels.table.channelId
}

func NewTimeTrialChannelsService(channels *TimeTrialChannels) *TimeTrialChannelsService {
	service := &TimeTrialChannelsService{}

	if channels != nil {
		service.Channels = *channels
	}

	l.Log.Info("sub " + channels.submissionsId)
	l.Log.Info("ct sub" + channels.customTrackSubmissionsId)
	l.Log.Info("dis" + channels.discussionId)
	l.Log.Info("table" + channels.table.channelId)
	l.Log.Info("table msg" + channels.table.standingsMessageId)

	l.Log.Info("sub " + service.Channels.submissionsId)
	l.Log.Info("ct sub" + service.Channels.customTrackSubmissionsId)
	l.Log.Info("dis" + service.Channels.discussionId)
	l.Log.Info("table" + service.Channels.table.channelId)
	l.Log.Info("table msg" + service.Channels.table.standingsMessageId)

	return service
}

func (ttcs *TimeTrialChannelsService) Setup(
	session *discordgo.Session,
	guildId string,
) error {
	ttChId, err := getChannelId(session, guildId, "time trials", discordgo.ChannelTypeGuildCategory)

	if err != nil {
		return err
	}

	ttcs.timeTrialCategoryChannelId = ttChId

	return nil
}

func (ttcs *TimeTrialChannelsService) CreateNewChannels(
	session *discordgo.Session,
	guildId string) (newSeasonForumId string, error error) {
	name, err := getNewSeasonForumName(session, guildId)
	if err != nil {
		return "", fmt.Errorf("Could not get new season forum name %w", err)
	}

	forum, err := session.GuildChannelCreateComplex(guildId,
		discordgo.GuildChannelCreateData{
			Name:     name,
			Type:     discordgo.ChannelTypeGuildForum,
			ParentID: ttcs.timeTrialCategoryChannelId,
		},
	)
	if err != nil {
		return "", fmt.Errorf("Could not create forum %w", err)
	}

	discussionThread, err := session.ForumThreadStartComplex(forum.ID, &discordgo.ThreadStart{
		Name:                "Discussion",
		AutoArchiveDuration: sevenDaysInMinutes,
		Type:                discordgo.ChannelTypeGuildPublicThread,
		AppliedTags:         []string{},
	}, &discordgo.MessageSend{
		Content: "Be collaborative! Discuss and share strats etc..",
	})
	if err != nil {
		return "", fmt.Errorf("Could not create discussion channel %w", err)
	}

	ctSubmissionsThread, err := session.ForumThreadStartComplex(forum.ID, &discordgo.ThreadStart{
		Name:                "Custom track submissions",
		AutoArchiveDuration: sevenDaysInMinutes,
		Type:                discordgo.ChannelTypeGuildPublicThread,
		AppliedTags:         []string{},
	}, &discordgo.MessageSend{
		Content: "To submit custom tracks, for fun",
	})

	if err != nil {
		return "", fmt.Errorf("Could not create CT submissions channel %w", err)
	}

	tableThread, err := session.ForumThreadStartComplex(forum.ID, &discordgo.ThreadStart{
		Name:                "Table",
		AutoArchiveDuration: sevenDaysInMinutes,
		Type:                discordgo.ChannelTypeGuildPublicThread,
		AppliedTags:         []string{},
	},
		&discordgo.MessageSend{
			Content: "The current standings:",
		})

	if err != nil {
		return "", fmt.Errorf("Could not create table channel %w", err)
	}

	submissionsThread, err := session.ForumThreadStartComplex(forum.ID, &discordgo.ThreadStart{
		Name:                "Submissions",
		AutoArchiveDuration: sevenDaysInMinutes,
		Type:                discordgo.ChannelTypeGuildPublicThread,
		AppliedTags:         []string{},
	}, &discordgo.MessageSend{
		Content: "Enter your submissions here. Reminder: submissions must include a picture of your run with the final time and splits visible and your ghost file",
	})

	if err != nil {
		return "", fmt.Errorf("Could not create submissions channel %w", err)
	}

	ttcs.mu.Lock()
	defer ttcs.mu.Unlock()

	ttcs.Channels.discussionId = discussionThread.ID
	ttcs.Channels.customTrackSubmissionsId = ctSubmissionsThread.ID

	ttcs.Channels.table.channelId = tableThread.ID
	ttcs.Channels.table.standingsMessageId = tableThread.LastMessageID

	ttcs.Channels.submissionsId = submissionsThread.ID
	return forum.ID, nil
}

func (ttcs *TimeTrialChannelsService) LockCurrentChannels(session *discordgo.Session, guildID string) error {
	locked := true

	ttcs.mu.Lock()
	defer ttcs.mu.Unlock()

	if ttcs.Channels.discussionId != "" {
		_, err := session.ChannelEdit(ttcs.Channels.discussionId, &discordgo.ChannelEdit{
			Locked: &locked,
		})

		if err != nil {
			return err
		}
	}

	if ttcs.Channels.customTrackSubmissionsId != "" {
		_, err := session.ChannelEdit(ttcs.Channels.customTrackSubmissionsId, &discordgo.ChannelEdit{
			Locked: &locked,
		})

		if err != nil {
			return err
		}
	}

	if ttcs.Channels.table.channelId != "" {
		_, err := session.ChannelEdit(ttcs.Channels.table.channelId, &discordgo.ChannelEdit{
			Locked: &locked,
		})

		if err != nil {
			return err
		}
	}

	if ttcs.Channels.submissionsId != "" {
		_, err := session.ChannelEdit(ttcs.Channels.submissionsId, &discordgo.ChannelEdit{
			Locked: &locked,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Create new table with standings
func (ttcs *TimeTrialChannelsService) NewTableMessage(
	session *discordgo.Session,
	text string,
	table *discordgo.File) error {

	ttcs.mu.Lock()
	defer ttcs.mu.Unlock()

	msg, err := session.ChannelMessageSendComplex(*&ttcs.Channels.table.channelId,
		&discordgo.MessageSend{
			Content: text,
			Files: []*discordgo.File{
				table,
			},
		})

	if err != nil {
		return fmt.Errorf("Could not send new table msg: %w", err)
	}

	ttcs.Channels.table.standingsMessageId = msg.ID
	return nil
}

// Update current table with new standings
func (ttcs *TimeTrialChannelsService) UpdateTableMessage(
	session *discordgo.Session,
	text string,
	table *discordgo.File) error {

	ttcs.mu.Lock()
	defer ttcs.mu.Unlock()

	// NOTE: Need to replace message if exists to update attachment.
	// Adding new message first in case removing goes wrong
	msg, err := session.ChannelMessageSendComplex(*&ttcs.Channels.table.channelId,
		&discordgo.MessageSend{
			Content: text,
			Files: []*discordgo.File{
				table,
			},
		})

	if err != nil {
		return fmt.Errorf("Could not send new table msg: %w", err)
	} else if ttcs.Channels.table.standingsMessageId != "" {
		err = session.ChannelMessageDelete(ttcs.Channels.table.channelId, ttcs.Channels.table.standingsMessageId)
		if err != nil {
			return fmt.Errorf("Could not delete msg for %s: %w", ttcs.Channels.table.standingsMessageId, err)
		}
	}

	ttcs.Channels.table.standingsMessageId = msg.ID
	return nil
}

func getNewSeasonForumName(session *discordgo.Session, guildId string) (name string, err error) {
	channels, err := session.GuildChannels(guildId)
	if err != nil {
		return "", err
	}

	const prefix = "month-"
	const suffix = "-time-trials"

	curIndex := 0

	// Retrieve the highest index
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildForum &&
			strings.HasPrefix(ch.Name, prefix) &&
			strings.HasSuffix(ch.Name, suffix) {

			curIndexRaw := ch.Name[len(prefix) : len(ch.Name)-len(suffix)]

			index, err := strconv.Atoi(curIndexRaw)

			if err != nil {
				return "", errors.New("Could not parse index to int")
			}

			if index > curIndex {
				curIndex = index
			}
		}
	}

	return prefix + strconv.Itoa(curIndex+1) + suffix, nil
}

// NOTE: this is a generic method. Move to common place if needed elsewhere
func getChannelId(
	session *discordgo.Session,
	guildId string,
	channelName string,
	channelType discordgo.ChannelType) (string, error) {
	channels, err := session.GuildChannels(guildId)

	if err != nil {
		return "", err
	}

	for _, ch := range channels {
		if channelType == ch.Type && strings.EqualFold(channelName, ch.Name) {
			return ch.ID, nil
		}
	}

	return "", err
}
