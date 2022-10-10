package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/ko3luhbka/task_tracker/db"
	"github.com/ko3luhbka/task_tracker/mq"
	"github.com/ko3luhbka/task_tracker/rest/model"
)

type Service struct {
	taskRepo     *db.TaskRepo
	assigneeRepo *db.AssigneeRepo
	Mq *mq.Client
}

func NewService(tr *db.TaskRepo, ar *db.AssigneeRepo, mq *mq.Client) *Service {
	return &Service{
		taskRepo: tr,
		assigneeRepo: ar,
		Mq: mq,
	}
}

func (s Service) CreateTask(ctx context.Context, t model.Task) (*db.Task, error) {
	assignee, err := s.getRandomAssignee(ctx)
	if err != nil {
		return nil, err
	}
	t.AssigneeID = assignee.ID

	created, err := s.taskRepo.Create(ctx, *t.ToEntity())
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s Service) GetTaskByID(ctx context.Context, uuid string) (*db.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s Service) GetAllTasks(ctx context.Context) ([]db.Task, error) {
	tasks, err := s.taskRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s Service) UpdateTask(ctx context.Context, t model.Task) (*db.Task, error) {
	t.RemoveAssignee()
	updated, err := s.taskRepo.Update(ctx, *t.ToEntity())
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s Service) DeleteTask(ctx context.Context, uuid string) error {
	return s.taskRepo.Delete(ctx, uuid)
}

func (s Service) ReassignTasks(ctx context.Context) error {
	

	tasks, err := s.taskRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		err := fmt.Errorf("no tasks to reassign")
		log.Println(err)
		return err
	}

	for _, task := range tasks {
		assignee, err := s.getRandomAssignee(ctx)
		if err != nil {
			return err
		}
		task.AssigneeID = assignee.ID
		_, err = s.taskRepo.Update(ctx, task)
		if err != nil {
			err = fmt.Errorf("failed to reassign task %s: %v", task.ID, err)
			return err
		}
	}

	return nil
}

func (s Service) getRandomAssignee(ctx context.Context) (*db.Assignee, error) {
	assignees, err := s.assigneeRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(assignees) == 0 {
		err := fmt.Errorf("no assignee found to assign tasks to")
		log.Println(err)
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(assignees))
	return &assignees[idx], nil
}

func (s Service) ConsumeMsg(errCh chan error) {
	go func(errCh chan<- error) {
		var err error
		defer func() {
			errCh <- err
			if err := s.Mq.Reader.Close(); err != nil {
				log.Printf("failed to close Reader: %v\n", err)
			}
		}()

		for {
			m, err := s.Mq.Reader.ReadMessage(context.Background())
			if err != nil {
				err = fmt.Errorf("failed to read message from topic: %v", err)
				log.Println(err)
				return
			}
			ctx := context.Background()
			if err := s.handleEvent(ctx, &m); err != nil {
				log.Printf("failed to handle event: %v", err)
			}
		}
	}(errCh)
}

func (s Service) ProduceMsg(ctx context.Context, e *mq.UserEvent) error {
	msgValue, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal Kafka event: %v", err)
	}
	msg := kafka.Message{
		Key:   nil,
		Value: msgValue,
	}

	if err := s.Mq.Writer.WriteMessages(ctx, msg); err != nil {
		log.Fatal("failed to write message:", err)
	}
	return nil
}


func (s Service) handleEvent(ctx context.Context, msg *kafka.Message) error {
	log.Printf("message at offset %d: %s = %s\n", msg.Offset, string(msg.Key), string(msg.Value))

	var e mq.UserEvent
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return err
	}

	switch e.Name {
	case mq.UserCreatedEvent:
		_, err := s.assigneeRepo.Create(ctx, e.Data)
		if err != nil {
			return fmt.Errorf("failed to create incoming assignee: %v", err)
		}
	case mq.UserUpdatedEvent:
		_, err := s.assigneeRepo.Update(ctx, e.Data)
		if err != nil {
			return fmt.Errorf("failed to update incoming assignee: %v", err)
		}
	case mq.UserDeletedEvent:
		err := s.assigneeRepo.Delete(ctx, e.Data.ID)
		if err != nil {
			return fmt.Errorf("failed to delete incoming assignee: %v", err)
		}
	default:
		return fmt.Errorf("unknown event name: %v", e.Name)
	}
	return nil
}
