package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/guildstore"
	"github.com/mitaka8/playlist-bot/internal/player"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NewDiscordBotParams struct {
	fx.In

	Config *Configuration
	Log    *zap.Logger
	Repo   *guildstore.GuildRepository
	Exe    *executor.Executor
	Player *player.Player
}

func NewDiscordBot(p NewDiscordBotParams) *DiscordBot {
	return &DiscordBot{
		botToken: p.Config.BotToken,
		log:      p.Log.Named("discord"),
		ready:    false,
		exe:      p.Exe,
		repo:     p.Repo,
		player:   p.Player,
	}
}

type DiscordBot struct {
	conn *discordgo.Session
	repo *guildstore.GuildRepository

	log      *zap.Logger
	botToken string
	ready    bool
	exe      *executor.Executor
	player   *player.Player
}

func (d *DiscordBot) AddHandlers() error {
	d.conn.AddHandler(messagehandler(d))
	d.conn.AddHandler(ready(d))
	d.conn.AddHandler(guildCreate(d))
	return nil
}

func (d *DiscordBot) ListenQueuedMessages() error {
	//	msgs, err := d.queue.Consume()
	//	if err != nil {
	//		return err
	//	}
	//	go func() {
	//
	//		for m := range msgs {
	//			d.log.Info("Send message from AMQP chan",
	//				zap.String("channel", m.ChannelID),
	//				zap.String("content", m.Content),
	//			)
	//
	//			if m.ChannelID == "" || m.Content == "" {
	//				d.log.Error("Invalid message request from AMQP chan",
	//					zap.String("channel", m.ChannelID),
	//					zap.String("content", m.Content),
	//				)
	//				return
	//			}
	//
	//			content := escapeDiscordMessage(m.Content)
	//
	//			_, err := d.conn.ChannelMessageSend(m.ChannelID, content)
	//			if err != nil {
	//				d.log.Error("error while listening to queued messages",
	//					zap.Error(err),
	//					zap.String("channel", m.ChannelID),
	//					zap.String("content", m.Content),
	//				)
	//			}
	//
	//		}
	//	}()
	return nil
}

func (d *DiscordBot) Connect() error {
	ses, err := discordgo.New("Bot " + d.botToken)

	if err != nil {
		return err
	}
	ses.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)
	d.conn = ses

	return d.conn.Open()
}

func escapeDiscordMessage(s string) string {
	s = strings.ReplaceAll(s, "@", "")
	s = strings.ReplaceAll(s, "#", "")

	return s
}
