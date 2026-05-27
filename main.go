package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mkw_tt_event_discord_bot/internal/commands"
	"mkw_tt_event_discord_bot/internal/commands/settime"
	"mkw_tt_event_discord_bot/internal/export"
	"mkw_tt_event_discord_bot/internal/messagehandlers"
	"mkw_tt_event_discord_bot/internal/services"
	"mkw_tt_event_discord_bot/internal/timetrial"
	"mkw_tt_event_discord_bot/pkg/httpclient"
	l "mkw_tt_event_discord_bot/pkg/logger"
)

func main() {
	l.Init()
	l.Log.Info("Starting...")

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		l.Log.Error("No API key")
		panic("No API key found")
	}

	guildId := os.Getenv("GUILD_ID")
	if guildId == "" {
		l.Log.Error("No guild ID")
		panic("No guild ID found")
	}

	shouldResetCommands := flag.Bool("reset-commands", false, "Whether to delete and readd discord commands once at startup")
	flag.Parse()

	var client = httpclient.NewClient(15 * time.Second)

	standings := export.Load()

	var timeTrialsChannels services.TimeTrialChannels
	var times []timetrial.Time = nil
	var currentTrack string = ""
	var rrExpert *timetrial.Time = nil
	var bkt *timetrial.Time = nil
	var ftreExpert *timetrial.Time = nil
	if standings != nil {
		timeTrialsChannels = services.FromExportChannels(standings.TimeTrialChannels)
		times = standings.Times
		currentTrack = standings.CurrentTrack
		rrExpert = standings.RRExpert
		bkt = standings.BKT
		ftreExpert = standings.FtreExpert
	}

	timeTrialChannelService := services.NewTimeTrialChannelsService(&timeTrialsChannels)
	lorenziService := services.NewLorenzi(client)
	bktService := services.NewBKTService(client)
	standingsService := services.NewStandings(times, currentTrack, rrExpert, bkt, ftreExpert)

	bot, err := NewBot(
		apiKey,
		guildId,
		*shouldResetCommands,
		timeTrialChannelService,
		[]commands.Command{
			commands.NewNewSeasonCommand(timeTrialChannelService, standingsService),
			commands.NewEsnCommand(
				bktService, standingsService, lorenziService, timeTrialChannelService),

			commands.NewResetCommand(standingsService, timeTrialChannelService, lorenziService),
			commands.NewPurgeCommand(standingsService, timeTrialChannelService, lorenziService),

			settime.NewSetFtreCommand(standingsService, lorenziService, timeTrialChannelService),
			settime.NewSetRRExpertCommand(standingsService, lorenziService, timeTrialChannelService),
			settime.NewSetBktCommand(standingsService, lorenziService, timeTrialChannelService),

			commands.NewTrackListCommand(bktService),
		},
		[]messagehandlers.MessageHandler{
			messagehandlers.NewTimeSubmissionMessageHandler(
				standingsService, timeTrialChannelService, lorenziService),
		},
	)
	defer bot.Close()

	if err != nil {
		l.Log.Error("Could not create bot", err)
		panic(err)
	}

	l.Log.Info("Bot is running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
