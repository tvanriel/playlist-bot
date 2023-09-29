package app

import (
	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/commands"
	"github.com/mitaka8/playlist-bot/internal/config"
	"github.com/mitaka8/playlist-bot/internal/discord"
	"github.com/mitaka8/playlist-bot/internal/guildstore"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/player"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"github.com/mitaka8/playlist-bot/internal/progresstracker"
	"github.com/mitaka8/playlist-bot/internal/web"
	"github.com/mitaka8/playlist-bot/internal/youtubedl"
	"github.com/tvanriel/cloudsdk/amqp"
	"github.com/tvanriel/cloudsdk/http"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"github.com/tvanriel/cloudsdk/logging"
	"github.com/tvanriel/cloudsdk/mysql"
	"github.com/tvanriel/cloudsdk/redis"
	"github.com/tvanriel/cloudsdk/s3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func DiscordBot() {
	fx.New(
		fx.Provide(
			config.ViperConfiguration,
			config.S3Configuration,
			config.AmqpConfiguration,
			config.MySQLConfiguration,
			config.LoggingConfiguration,
			config.DiscordConfiguration,
			config.YoutubeDLConfiguration,
			config.KubernetesConfiguration,
			config.MusicStoreConfiguration,
			config.RedisConfiguration,
		),
		logging.FXLogger(),
		logging.Module,
		guildstore.Module,
		mysql.Module,
		player.Module,
		discord.Module,
		executor.Module,
		kubernetes.Module,
		commands.Module,
		playliststore.Module,
		youtubedl.Module,
		s3.Module,
		musicstore.Module,
		progresstracker.Module,
		amqp.Module,
		redis.Module,
		fx.Invoke(progresstracker.StartReporting),
	).Run()
}

func Web() {
	fx.New(
		fx.Provide(
			config.ViperConfiguration,
			config.MySQLConfiguration,
			config.AmqpConfiguration,
			config.LoggingConfiguration,
			config.HttpConfiguration,
			config.S3Configuration,
			config.MusicStoreConfiguration,
			config.RedisConfiguration,
		),
		logging.FXLogger(),
		logging.Module,
		http.Module,
		web.Module,
		guildstore.Module,
		mysql.Module,
		progresstracker.Module,
		amqp.Module,
		playliststore.Module,

		redis.Module,
		s3.Module,
		musicstore.Module,

		fx.Decorate(web.DecorateTemplater),
		fx.Invoke(http.Listen),
	).Run()
}

func Save(source, guildId, playlistName string) {
	fx.New(
		fx.Provide(
			config.ViperConfiguration,
			config.MySQLConfiguration,
			config.S3Configuration,
			config.AmqpConfiguration,
			config.LoggingConfiguration,
			config.YoutubeDLConfiguration,
			config.MusicStoreConfiguration,
			config.KubernetesConfiguration,
		),
		logging.FXLogger(),
		logging.Module,
		s3.Module,
		mysql.Module,
		kubernetes.Module,
		youtubedl.Module,
		musicstore.Module,
		playliststore.Module,

		fx.Invoke(func(dl youtubedl.YoutubeDL, log *zap.Logger) {
			p := youtubedl.YouTubeDLParams{
				Source:       source,
				GuildID:      guildId,
				PlaylistName: playlistName,
			}
			err := dl.Save(p)
			if err != nil {
				log.With(p.ZapFields()...).Fatal("Could not download", zap.Error(err))
			}
		}),
	)
}
