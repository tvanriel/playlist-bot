package saver

import (
	"github.com/minio/minio-go/v7"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Configuration struct {
	Bucket string
}

type NewSaverParams struct {
	fx.In

	Log           *zap.Logger
	Configuration Configuration
	S3            *minio.Client
}

func NewSaver(p NewSaverParams) *Saver {
	return &Saver{
		Log:    p.Log,
		Bucket: p.Configuration.Bucket,
		S3:     p.S3,
	}

}

type Saver struct {
	Log    *zap.Logger
	Bucket string
	S3     *minio.Client
}

func (s *Saver) SaveTrack(guildId, uuid string, source string) {

}
