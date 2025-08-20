package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/floriansw/discord-auto-publish/internal"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var channels = make(map[string]*discordgo.Channel)

func main() {
	level := slog.LevelInfo
	if _, ok := os.LookupEnv("DEBUG"); ok {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	c, err := internal.NewConfig("./config.json", logger)
	if err != nil {
		logger.Error("config", err)
		return
	}

	var s *discordgo.Session
	if c.Discord != nil {
		s, err = discordgo.New("Bot " + c.Discord.Token)
		if err != nil {
			logger.Error("discord", err)
			return
		}
	}
	if s != nil {
		s.AddHandlerOnce(func(s *discordgo.Session, e *discordgo.Ready) {
			logger.Info("ready")
		})
		s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageCreate) {
			l := logger.WithGroup("on-message").With("channelId", e.Message.ChannelID, "messageId", e.Message.ID)
			l.Info("receive")
			if !c.Channels.Contains(e.Message.ChannelID) {
				return
			}
			ch, err := channel(s, e.Message.ChannelID)
			if err != nil {
				l.Error("channel-from-message", "error", err)
				return
			}
			if ch.Type != discordgo.ChannelTypeGuildNews {
				return
			}
			_, err = s.ChannelMessageCrosspost(e.Message.ChannelID, e.Message.ID)
			if err != nil {
				l.Error("publish", "error", err)
			} else {
				l.Info("cross-posted")
			}
		})
		err = s.Open()
		if err != nil {
			logger.Error("open-session", err)
			return
		}
		defer s.Close()
	}
	defer func() {
		err = c.Save()
		if err != nil {
			logger.Error("save-config", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("graceful-shutdown")
}

func channel(s *discordgo.Session, channelId string) (*discordgo.Channel, error) {
	if ch, ok := channels[channelId]; ok {
		return ch, nil
	}
	ch, err := s.Channel(channelId)
	if err != nil {
		return nil, err
	}
	channels[channelId] = ch
	return ch, nil
}
