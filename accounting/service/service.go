package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/ko3luhbka/accounting/db"
	"github.com/ko3luhbka/accounting/mq"
	"github.com/ko3luhbka/accounting/rest/model"
)

type Service struct {
	accountRepo *db.AccountRepo
	auditRepo   *db.AuditRepo
	Mq          *mq.Client
}

func NewService(accr *db.AccountRepo, audr *db.AuditRepo, mq *mq.Client) *Service {
	return &Service{
		accountRepo: accr,
		auditRepo:   audr,
		Mq:          mq,
	}
}

func (s Service) GetUserBalance(ctx context.Context, userUUID string) (int, error) {
	balance, err := s.accountRepo.GetUserBalance(ctx, userUUID)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (s Service) DeleteUserAccount(ctx context.Context, uuid string) error {
	return s.accountRepo.DeleteByUser(ctx, uuid)
}

func (s Service) PayToUser(ctx context.Context, uuid string) (int, error) {
	acc := &db.Account{
		AssigneeID: uuid,
		Debit:      getRandNumInRange(20, 40),
	}
	created, err := s.accountRepo.CreateRecord(ctx, acc)
	if err != nil {
		return 0, err
	}
	return created.Debit, nil
}

func (s Service) WithdrawUser(ctx context.Context, uuid string) (int, error) {
	acc := &db.Account{
		AssigneeID: uuid,
		Credit:     -getRandNumInRange(10, 20),
	}
	created, err := s.accountRepo.CreateRecord(ctx, acc)
	if err != nil {
		return 0, err
	}
	return created.Credit, nil
}

func (s Service) GetManagementIncome(ctx context.Context) (int, error) {
	income, err := s.accountRepo.GetManagementTodayIncome(ctx)
	if err != nil {
		return 0, err
	}
	return income, nil
}

func getRandNumInRange(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func (s Service) CreateAuditRecord(ctx context.Context, aud *db.Audit) (*model.Audit, error) {
	audit, err := s.auditRepo.Create(ctx, aud)
	if err != nil {
		return nil, err
	}

	m := new(model.Audit)
	m.FromEntity(audit)
	return m, nil
}

func (s Service) GetUserAuditLog(ctx context.Context, uuid string) ([]model.Audit, error) {
	auditLog, err := s.auditRepo.GetUserAudit(ctx, uuid)
	if err != nil {
		return nil, err
	}
	auditLogModel := make([]model.Audit, len(auditLog))
	for i, aud := range auditLog {
		m := new(model.Audit)
		m.FromEntity(&aud)
		auditLogModel[i] = *m
	}
	return auditLogModel, nil
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

func (s Service) handleEvent(ctx context.Context, msg *kafka.Message) error {
	switch msg.Topic {
	case mq.UsersCUDTopic:
		return s.handleUserEvents(ctx, msg)
	case mq.TasksTopic:
		return s.handleTaskEvents(ctx, msg)
	default:
		return fmt.Errorf("unknownn topic: %s", msg.Topic)
	}
}

func (s Service) handleUserEvents(ctx context.Context, msg *kafka.Message) error {
	var e mq.UserEvent
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return err
	}

	user := e.Data.ID
	switch e.Name {
	case mq.UserDeletedEvent:
		if err := s.DeleteUserAccount(ctx, user); err != nil {
			return fmt.Errorf("failed to delete account of user %s: %v", user, err)
		}
		log.Printf("account of user %s was deleted", user)
		return nil
	default:
		return fmt.Errorf("unknown event name: %v", e.Name)
	}
}

func (s Service) handleTaskEvents(ctx context.Context, msg *kafka.Message) error {
	var e mq.TaskEvent
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return err
	}

	user := e.Data.AssigneeID
	switch e.Name {
	case mq.TaskAssignedEvent:
		amount, err := s.WithdrawUser(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to withdraw user %s: %v", user, err)
		}
		audit := &db.Audit{
			EventName:  mq.TaskAssignedEvent,
			AssigneeID: user,
			TaskID:     e.Data.ID,
			TaskTitle:  e.Data.Title,
			JiraID:     e.Data.JiraID,
			Amount:     amount,
		}
		_, err = s.CreateAuditRecord(ctx, audit)
		if err != nil {
			return err
		}
		log.Printf("user %s was withdrawed due to assigned task", user)
		return nil
	case mq.TaskCompletedEvent:
		amount, err := s.PayToUser(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to pay to user %s: %v", user, err)
		}
		audit := &db.Audit{
			EventName:  mq.TaskCompletedEvent,
			AssigneeID: user,
			TaskID:     e.Data.ID,
			TaskTitle:  e.Data.Title,
			JiraID:     e.Data.JiraID,
			Amount:     amount,
		}
		_, err = s.CreateAuditRecord(ctx, audit)
		if err != nil {
			return err
		}
		log.Printf("user %s was payed due to completed task", user)
		return nil
	default:
		return fmt.Errorf("unknown event name: %v", e.Name)
	}
}
