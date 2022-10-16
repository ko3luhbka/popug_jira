package mq

import (
	"github.com/ko3luhbka/accounting/rest/model"
	"github.com/segmentio/kafka-go"
)

const (
	kafkaHost = "localhost:29092"
	GroupID   = "accountingConsumer"

	UsersCUDTopic    = "usersStreaming"
	UserDeletedEvent = "userDeleted"

	TasksTopic           = "tasks"
	TaskAssignedEvent    = "taskAssigned"
	TasksReassignedEvent = "tasksReassigned"
	TaskCompletedEvent   = "taskCompleted"
)

type (
	Config struct {
		Consumer   bool
		Producer   bool
		ReadTopics []string
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
		Data model.TaskInfo `json:"data"`
	}
)

func NewMQClient(cfg *Config) *Client {
	var client Client
	if cfg.Consumer {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{kafkaHost},
			GroupTopics: cfg.ReadTopics,
			GroupID:     GroupID,
			MinBytes:    1,
			MaxBytes:    10e6,
		})
		client.Reader = reader
	}

	return &client
}
