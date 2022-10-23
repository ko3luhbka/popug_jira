package model

import (
	"time"

	"github.com/ko3luhbka/accounting/db"
)

type (
	UserInfo struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	TaskInfo struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		JiraID     string `json:"jira_id"`
		AssigneeID string `json:"assignee_id"`
	}
	Account struct {
		ID         int       `json:"id"`
		AssigneeID string    `json:"assignee_id"`
		Credit     int       `json:"credit"`
		Debit      int       `json:"debit"`
		Created    time.Time `json:"created"`
	}
	Audit struct {
		ID         int       `json:"id"`
		EventName  string    `json:"event_name"`
		AssigneeID string    `json:"assignee_id"`
		TaskID     string    `json:"task_id"`
		TaskTitle  string    `json:"task_title"`
		JiraID     string    `json:"jira_id"`
		Amount     int       `json:"amount"`
		Created    time.Time `json:"created"`
	}
	UserIncome struct {
	}
)

func (acc *Account) ToEntity() *db.Account {
	return &db.Account{
		ID:         acc.ID,
		AssigneeID: acc.AssigneeID,
		Credit:     acc.Credit,
		Debit:      acc.Debit,
		Created:    acc.Created,
	}
}

func (acc *Account) FromEntity(e *db.Account) {
	acc.ID = e.ID
	acc.AssigneeID = e.AssigneeID
	acc.Credit = e.Credit
	acc.Debit = e.Debit
	acc.Created = e.Created
}

func (aud *Audit) ToEntity() *db.Audit {
	return &db.Audit{
		ID:         aud.ID,
		EventName:  aud.EventName,
		AssigneeID: aud.AssigneeID,
		TaskID:     aud.TaskID,
		TaskTitle:  aud.TaskTitle,
		JiraID:     aud.JiraID,
		Created:    aud.Created,
	}
}

func (aud *Audit) FromEntity(e *db.Audit) {
	aud.ID = e.ID
	aud.EventName = e.EventName
	aud.AssigneeID = e.AssigneeID
	aud.TaskID = e.TaskID
	aud.TaskTitle = e.TaskTitle
	aud.JiraID = e.JiraID
	aud.Amount = e.Amount
	aud.Created = e.Created
}
