package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	playlistdiscordbot "gitlab.com/mitaka8/playlist-discord-bot"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewLogger() *log.Logger {
	return log.New(os.Stdout, "", 0)
}

func dsn() string {
	var sb strings.Builder
	sb.WriteString(viper.GetString("db.user"))
	sb.WriteString(":")
	sb.WriteString(viper.GetString("db.pass"))
	sb.WriteString("@tcp(")
	sb.WriteString(viper.GetString("db.host"))
	sb.WriteString(")/")
	sb.WriteString(viper.GetString("db.name"))
	sb.WriteString("?charset=utf8&parseTime=True&loc=Local")
	log.Print(sb.String())
	return sb.String()
}

func NewDatabase(lc fx.Lifecycle) *gorm.DB {
	var db *gorm.DB
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			conn, err := db.DB()
			if err != nil {
				return err
			}
			return conn.Close()
		},
	})
	return db

}

func NewGuildRepository(db *gorm.DB) *playlistdiscordbot.GuildRepository {
	return &playlistdiscordbot.GuildRepository{
		DB:        db,
		SoundRoot: viper.GetString("path.sound_root"),
		Guilds:    map[string]*playlistdiscordbot.Guild{},
	}
}

func NewDiscordbot(lc fx.Lifecycle, guilds *playlistdiscordbot.GuildRepository, logger *log.Logger) *playlistdiscordbot.Bot {
	bot := &playlistdiscordbot.Bot{
		Guilds:         guilds,
		Logger:         logger,
		Authentication: viper.GetString("discord.bot_token"),
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go bot.Connect()

			return nil
		},
		OnStop: func(context.Context) error {
			bot.Disconnect()
			return nil
		},
	})

	return bot
}

func RegisterBotMessages(bot *playlistdiscordbot.Bot) {
	go bot.ListenForMessages()
}

func MigrateDatabase(db *gorm.DB) {
	go func() {
		for db == nil {
			// Wait for database connection
			time.Sleep(1 * time.Second)
		}
		db.AutoMigrate(&playlistdiscordbot.Guild{})
	}()
}

func ConnectToDatabase(db *gorm.DB) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn()))
}

func setupConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/playlist-bot")

	return viper.ReadInConfig()
}

func main() {
	setupConfig()
	os.Mkdir(viper.GetString("path.sound_root"), 0755)

	app := fx.New(
		fx.Provide(
			NewLogger,
			NewDatabase,
			NewGuildRepository,
			NewDiscordbot,
		),
		fx.Decorate(ConnectToDatabase),
		fx.Invoke(RegisterBotMessages, MigrateDatabase),
	)

	startCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		log.Fatal(err)
	}

	<-app.Done()

	stopCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}

}
