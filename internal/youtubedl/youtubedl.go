package youtubedl

import (
	"errors"

	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Configuration struct {
	Implementation string
}

type NewYouTubeDLParams struct {
	fx.In

	Kubernetes    *kubernetes.KubernetesClient
	Configuration Configuration
	Log           *zap.Logger
	MusicStore    *musicstore.MusicStore
	PlaylistStore *playliststore.PlaylistStore
}

type YouTubeDLParams struct {
	Source       string
	GuildID      string
	PlaylistName string
}

func (p YouTubeDLParams) ZapFields() []zap.Field {
	return []zap.Field{
		zap.String("source", p.Source),
		zap.String("guildId", p.GuildID),
		zap.String("playlistName", p.PlaylistName),
	}
}

type YoutubeDL interface {
	Save(params YouTubeDLParams) error
}

func NewYouTubeDL(p NewYouTubeDLParams) (YoutubeDL, error) {
	switch p.Configuration.Implementation {
	case "kubernetes":
		return &KubernetesYouTubeDL{
			Configuration: p.Configuration,
			Kubernetes:    p.Kubernetes,
			Log:           p.Log.Named("ytdl-kubernetes"),
		}, nil
	case "exec":
		return &ExecYouTubeDL{
			Configuration: p.Configuration,
			Log:           p.Log.Named("ytdl-exec"),
			MusicStore:    p.MusicStore,
			PlaylistStore: p.PlaylistStore,
		}, nil
	default:
		return nil, errors.New("invalid youtubedl implementation, expected exec or kubernetes")
	}
}
