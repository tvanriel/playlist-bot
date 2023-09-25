package youtubedl

import (
	"errors"

	"github.com/mitaka8/playlist-bot/internal/musicstore"
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
}

type YoutubeDL interface {
	Save(source, guildId, uuid string)
}

func NewYouTubeDL(p NewYouTubeDLParams) (YoutubeDL, error) {
	switch p.Configuration.Implementation {
	case "kubernetes":
		return &KubernetesYouTubeDL{
			Configuration: p.Configuration,
			Kubernetes:    p.Kubernetes,
			Log:           p.Log,
		}, nil
	case "exec":
		return &ExecYouTubeDL{
			Configuration: p.Configuration,
			Log:           p.Log,
			MusicStore:    p.MusicStore,
		}, nil
	default:
		return nil, errors.New("invalid youtubedl implementation, expected exec or kubernetes")
	}
}
