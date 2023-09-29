package progresstracker

import (
	"context"
	"encoding/json"
	"github.com/tvanriel/cloudsdk/redis"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NewProgressTrackerParams struct {
	fx.In
	Log   *zap.Logger
	Redis *redis.RedisClient
}

func NewProgressTracker(p NewProgressTrackerParams) *ProgressTracker {
	return &ProgressTracker{
		Progress: make(map[string]*Progress),
		Log:      p.Log.Named("progress-tracker"),
		Redis:    p.Redis,
	}
}

type ProgressTracker struct {
	Progress map[string]*Progress
	Log      *zap.Logger
	Redis    *redis.RedisClient
}

type Progress struct {
	Current int
	Max     int
	Track   string
}

func (s *ProgressTracker) Consume(guildId string) chan *Progress {

	out := make(chan *Progress)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		s.getprogress(out, guildId)
		for range ticker.C {
			s.getprogress(out, guildId)
		}

	}()
	return out
}

func (s *ProgressTracker) getprogress(out chan *Progress, guildId string) {
	log := s.Log.With(zap.String("guildId", guildId))
	key := "progress"
	conn := s.Redis.Conn()
	cmd := conn.Get(context.Background(), key+guildId)
	err := cmd.Err()
	if err != nil {
		log.Error("Cannot fetch progress from redis", zap.Error(err))
		return
	}

	m, err := cmd.Bytes()
	if err != nil {
		log.Error("cannot get redis response as bytes", zap.Error(err))
		return
	}
	message := &Progress{}
	err = json.Unmarshal(m, message)
	if err != nil {
		s.Log.Error("cannot unmarshal message from redis", zap.Error(err))
		return
	}
	out <- message
}

func (p *ProgressTracker) Publish() {
	key := "progress"

	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		for range t.C {
			for guildId := range p.Progress {
				body, err := json.Marshal(p.Progress[guildId])
				if err != nil {
					p.Log.Error("cannot Marshal progress",
						zap.String("guildId", guildId),
						zap.Error(err),
					)
					break
				}

				conn := p.Redis.Conn()
				cmd := conn.Set(
					context.Background(),
					key+guildId,
					body,
					time.Hour,
				)
				if cmd.Err() != nil {
					p.Log.Error("cannot publish progress",
						zap.String("guildId", guildId),
						zap.Error(err),
					)
				}

			}

		}
	}()
}

func (p *ProgressTracker) Report(t *Progress, guildId string) {
	p.Progress[guildId] = t
}
