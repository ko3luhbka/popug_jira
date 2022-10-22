package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/ko3luhbka/popug_schema_registry/validator"

	"github.com/ko3luhbka/task_tracker/db"
	"github.com/ko3luhbka/task_tracker/mq"
	"github.com/ko3luhbka/task_tracker/rest/model"
)

const taskSchemaType = "task"

type Service struct {
	taskRepo     *db.TaskRepo
	assigneeRepo *db.AssigneeRepo
	Mq           *mq.Client
}

func NewService(tr *db.TaskRepo, ar *db.AssigneeRepo, mq *mq.Client) *Service {
	return &Service{
		taskRepo:     tr,
		assigneeRepo: ar,
		Mq:           mq,
	}
}

func (s Service) CreateTask(ctx context.Context, t model.Task) (*model.Task, error) {
	assignee, err := s.getRandomAssignee(ctx)
	if err != nil {
		return nil, err
	}
	t.AssigneeID = assignee.ID
	t.Status = model.TaskStatusAssigned

	created, err := s.taskRepo.Create(ctx, *t.ToEntity())
	if err != nil {
		return nil, err
	}

	e := mq.TaskEvent{
		Name: mq.TaskAssignedEvent,
		Version: 2,
		Data: *model.TaskEntityToTaskInfo(created),
	}
	if err := s.ProduceMsg(ctx, e); err != nil {
		return nil, err
	}

	m := new(model.Task)
	m.FromEntity(created)
	return m, nil
}

func (s Service) GetTaskByID(ctx context.Context, uuid string) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	m := new(model.Task)
	m.FromEntity(task)
	return m, nil
}

func (s Service) GetAllTasks(ctx context.Context) ([]model.Task, error) {
	tasks, err := s.taskRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	tasksModel := make([]model.Task, len(tasks))
	for i, task := range tasks {
		m := new(model.Task)
		m.FromEntity(&task)
		tasksModel[i] = *m
	}
	return tasksModel, nil
}

func (s Service) UpdateTask(ctx context.Context, t model.Task) (*model.Task, error) {
	t.RemoveAssignee()
	updated, err := s.taskRepo.Update(ctx, *t.ToEntity())
	if err != nil {
		return nil, err
	}

	if updated.Status == model.TaskStatusCompleted {
		e := mq.TaskEvent{
			Name: mq.TaskCompleted,
			Version: 2,
			Data: *model.TaskEntityToTaskInfo(updated),
		}
		if err := s.ProduceMsg(ctx, e); err != nil {
			return nil, err
		}
	}

	m := new(model.Task)
	m.FromEntity(updated)
	return m, nil
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

	reassignedTaskEvents := make([]mq.TaskEvent, len(tasks))
	for i, task := range tasks {
		assignee, err := s.getRandomAssignee(ctx)
		if err != nil {
			return err
		}
		task.AssigneeID = assignee.ID
		updated, err := s.taskRepo.Update(ctx, task)
		if err != nil {
			err = fmt.Errorf("failed to reassign task %s: %v", task.ID, err)
			return err
		}
		reassignedTaskEvents[i] = mq.TaskEvent{
			Name: mq.TaskAssignedEvent,
			Version: 2,
			Data: model.TaskInfo{
				ID:         updated.ID,
				AssigneeID: assignee.ID,
			},
		}
	}

	if err := s.ProduceMsg(ctx, reassignedTaskEvents...); err != nil {
		return err
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
			ctx := context.Background()
			m, err := s.Mq.Reader.ReadMessage(ctx)
			log.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
			if err != nil {
				log.Println(fmt.Errorf("failed to read message from topic: %v", err))
			}
			if err := s.handleEvent(ctx, &m); err != nil {
				log.Println(fmt.Errorf("failed to handle event: %v", err))
			}
		}
	}(errCh)
}

func (s Service) ProduceMsg(ctx context.Context, events ...mq.TaskEvent) error {
	
	msgs := make([]kafka.Message, len(events))
	for i, e := range events {
		if err := validator.Validate(e, taskSchemaType, 2); err != nil {
			log.Println(err)
			return fmt.Errorf("invalid event: %v", err)

		}

		msgValue, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("failed to marshal Kafka event: %v", err)
		}
		msg := kafka.Message{
			Key:   nil,
			Value: msgValue,
		}
		msgs[i] = msg
	}

	if err := s.Mq.Writer.WriteMessages(ctx, msgs...); err != nil {
		log.Fatal("failed to write message:", err)
	}
	return nil
}

func (s Service) handleEvent(ctx context.Context, msg *kafka.Message) error {
	var e mq.UserEvent
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return err
	}

	switch e.Name {
	case mq.UserCreatedEvent:
		_, err := s.assigneeRepo.Create(ctx, *e.Data.ToEntity())
		if err != nil {
			return fmt.Errorf("failed to create incoming assignee: %v", err)
		}
	case mq.UserUpdatedEvent:
		_, err := s.assigneeRepo.Update(ctx, *e.Data.ToEntity())
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
