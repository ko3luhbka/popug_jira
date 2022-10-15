package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ko3luhbka/auth/rest/model"
	"github.com/segmentio/kafka-go"
)

const (
	kafkaHost        = "localhost:29092"

	UsersCUDTopic       = "usersStreaming"
	UserCreatedEvent = "userCreated"
	UserUpdatedEvent = "userUpdated"
	UserDeletedEvent = "userDeleted"
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
		reader *kafka.Reader
		writer *kafka.Writer
	}
	UserEvent struct {
		Name string `json:"name"`
		Data *model.Assignee    `json:"data"`
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
		client.reader = reader
	}

	if cfg.Producer {
		w := &kafka.Writer{
			Addr:     kafka.TCP(kafkaHost),
			Topic:    cfg.WriteTopic,
			Balancer: &kafka.LeastBytes{},
		}
		client.writer = w
	}

	return &client
}

func (c *Client) Produce(ctx context.Context, e *UserEvent) error {
	msgValue, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal Kafka event: %v", err)
	}
	msg := kafka.Message{
		Key:   nil,
		Value: msgValue,
	}

	if err := c.writer.WriteMessages(ctx, msg); err != nil {
		log.Fatal("failed to write message:", err)
	}
	return nil
}
