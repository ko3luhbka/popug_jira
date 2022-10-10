package mq

import (
	"github.com/ko3luhbka/task_tracker/db"
	"github.com/segmentio/kafka-go"
)

const (
	kafkaHost        = "localhost:29092"
	UsersTopic       = "users"
	UserCreatedEvent = "user_created"
	UserUpdatedEvent = "user_updated"
	UserDeletedEvent = "user_deleted"
)

type (
	Config struct {
		Consumer   bool
		Producer   bool
		ReadTopic  string
		WriteTopic string
	}
	Client struct {
		config *Config
		Reader *kafka.Reader
		Writer *kafka.Writer
	}
	UserEvent struct {
		Name string      `json:"name"`
		Data db.Assignee `json:"data"`
	}
)

func NewMQClient(cfg *Config) *Client {
	var client Client
	if cfg.Consumer {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{kafkaHost},
			Topic:     cfg.ReadTopic,
			Partition: 0,
			MinBytes:  10e3, // 10KB
			MaxBytes:  10e6, // 10MB
		})
		reader.SetOffset(14)
		client.Reader = reader
	}

	if cfg.Producer {
		w := &kafka.Writer{
			Addr:     kafka.TCP(kafkaHost),
			Topic:    cfg.WriteTopic,
			Balancer: &kafka.LeastBytes{},
		}
		client.Writer = w
	}

	return &client
}
