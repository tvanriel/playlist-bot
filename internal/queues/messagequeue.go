package queues

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MessageQueue struct {
	Conn *amqp091.Connection
	Log  *zap.Logger

	Ch *amqp091.Channel

	Exchange string
	Prefix   string
}

type NewMessageQueueParams struct {
	fx.In
	Log    *zap.Logger
	Amqp   *amqp091.Connection
	Config Configuration
}

type MessageQueueBody struct {
	Channel string `json:"channel"`
	Content string `json:"content"`
}

func NewMessageQueue(p NewMessageQueueParams) *MessageQueue {
	return &MessageQueue{
		Conn: p.Amqp,
		Log:  p.Log.Named("messagequeue"),
	}
}

func (m *MessageQueue) Connect() error {
	if m.Conn != nil && !m.Ch.IsClosed() {
		return nil
	}

	ch, err := m.Conn.Channel()

	if err != nil {
		return err
	}

	m.Ch = ch
	return nil
}

func (m *MessageQueue) QueueName() string {
	return strings.Join([]string{
		m.Prefix,
		"messages",
	}, "")
}

func (m *MessageQueue) Append(channelId string, message string) error {

	body := &MessageQueueBody{
		Channel: channelId,
		Content: message,
	}

	err := m.Connect()
	if err != nil {
		return err
	}

	_, err = m.Ch.QueueDeclare(m.QueueName(), false, false, false, true, nil)

	marshalled, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return m.Ch.PublishWithContext(context.Background(), m.Exchange, m.QueueName(), false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        marshalled,
	})

}

func (m *MessageQueue) Consume() (chan MessageQueueBody, error) {
	err := m.Connect()

	if err != nil {
		return nil, err
	}
	out := make(chan MessageQueueBody)

	_, err = m.Ch.QueueDeclare(m.QueueName(), false, false, false, true, nil)

	if err != nil {
		return nil, err
	}

	msgs, err := m.Ch.Consume(m.QueueName(), "", true, false, false, true, nil)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			err := m.Connect()
			if err != nil {
				m.Log.Error("Cannot connect to AMQP channel", zap.Error(err))
				continue
			}

			for msg := range msgs {
				message := MessageQueueBody{}
				err := json.Unmarshal(msg.Body, &message)
				if err != nil {
					m.Log.Error("cannot unmarshal message from queue", zap.Error(err))
					continue
				}
				out <- message
			}
		}
	}()

	return out, nil
}
