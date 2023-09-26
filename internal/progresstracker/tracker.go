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
		Log:      p.Log.Named("progress-tracker"),
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

	out := make(chan *Progress)

	if err == nil {
		s.Log.Info("Consuming")
		go func() {
			msgs, err := s.Ch.Consume("progress-"+guildId, "", true, false, false, true, nil)
			if err != nil {
                                close(out)
				s.Log.Error("cannot consume from channel", zap.Error(err))
                                return
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
		}()
	}
	return out, err
}

func (p *ProgressTracker) Publish() {
	go func() {

		err := p.Connect()
		if err != nil {
			p.Log.Fatal("cannot connect to Queue", zap.Error(err))
		}

		t := time.NewTicker(500 * time.Millisecond)
		for range t.C {
			for guildId := range p.Progress {
				_, err = p.Ch.QueueDeclare("progress-"+guildId, false, false, false, true, nil)
                                if err != nil {
                                        p.Log.Error("cannot declare queue", zap.String("guildId", guildId), zap.Error(err))
                                }

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
