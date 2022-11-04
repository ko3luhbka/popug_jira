package mq

import (
	"github.com/ko3luhbka/task_tracker/rest/model"
	"github.com/segmentio/kafka-go"
)

const (
	kafkaHost = "localhost:29092"
	GroupID   = "usersConsumer"

	UsersCUDTopic    = "usersStreaming"
	UserCreatedEvent = "userCreated"
	UserUpdatedEvent = "userUpdated"
	UserDeletedEvent = "userDeleted"

	TasksTopic           = "tasks"
	TaskAssignedEvent    = "taskAssigned"
	TaskCompleted        = "taskCompleted"
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
		Name string         `json:"name"`
		Data model.UserInfo `json:"data"`
	}
	TaskEvent struct {
		Name string         `json:"name"`
		Version int `json:"version"`
		Data model.TaskInfo `json:"data"`
	}
)

func NewMQClient(cfg *Config) *Client {
	var client Client
	if cfg.Consumer {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{kafkaHost},
			Topic:    cfg.ReadTopic,
			GroupID:  GroupID,
			MinBytes: 1,
			MaxBytes: 10e6,
		})
		client.Reader = reader
	}

	if cfg.Producer {
		w := &kafka.Writer{
			Addr:                   kafka.TCP(kafkaHost),
			Topic:                  cfg.WriteTopic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
			RequiredAcks:           1,
		}
		client.Writer = w
	}

	return &client
}
