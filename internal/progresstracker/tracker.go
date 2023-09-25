package progresstracker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tvanriel/cloudsdk/amqp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NewProgressTrackerParams struct {
	fx.In
	Amqp *amqp.Connection
	Log  *zap.Logger
}

func NewProgressTracker(p NewProgressTrackerParams) *ProgressTracker {
	return &ProgressTracker{
		Amqp:     p.Amqp,
		Progress: make(map[string]*Progress),
		Log:      p.Log,
	}
}

type ProgressTracker struct {
	Amqp     *amqp.Connection
	Ch       *amqp.Channel
	Progress map[string]*Progress
	Log      *zap.Logger
}

type Progress struct {
	Current int
	Max     int
	Track   string
}

func (s *ProgressTracker) Connect() error {
	err := s.Amqp.Reconnect()
	if err != nil {
		return err
	}

	s.Ch, err = s.Amqp.Channel()
	return err
}

func (s *ProgressTracker) Consume(guildId string) (chan *Progress, error) {
	err := s.Connect()
	if err != nil {
		return nil, err
	}
	_, err = s.Ch.QueueDeclare("progress-"+guildId, false, false, false, true, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := s.Ch.Consume("progress-"+guildId, "", true, false, false, true, nil)

	out := make(chan *Progress)

	if err == nil {
		s.Log.Info("Consuming")
		go func() {
			for {
				err := s.Connect()
				if err != nil {
					s.Log.Error("Cannot connect to AQMP", zap.Error(err))
					continue
				}
				for m := range msgs {
					message := &Progress{}
					err := json.Unmarshal(m.Body, message)
					if err != nil {
						s.Log.Error("cannot unmarshal message from queue", zap.Error(err))
						continue
					}
					out <- message
				}
			}
		}()
	}
	return out, err
}

func (p *ProgressTracker) Publish() {
	go func() {

		t := time.NewTicker(500 * time.Millisecond)
		for range t.C {
			err := p.Connect()
			if err != nil {
				p.Log.Error("cannot connect to Queue", zap.Error(err))
				return
			}
			for guildId := range p.Progress {
				_, err = p.Ch.QueueDeclare("progress-"+guildId, false, false, false, true, nil)

				body, err := json.Marshal(p.Progress[guildId])
				if err != nil {
					p.Log.Error("cannot Marshal progress",
						zap.String("guildId", guildId),
						zap.Error(err),
					)
					break
				}

				err = p.Ch.PublishWithContext(
					context.Background(),
					"",
					"progress-"+guildId,
					false,
					false,
					amqp091.Publishing{
						Body: body,
					},
				)
				if err != nil {
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
