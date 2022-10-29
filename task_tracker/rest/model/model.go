package model

import (
	"fmt"
	"time"

	"github.com/ko3luhbka/task_tracker/db"
)

const (
	TaskStatusAssigned  = "Assigned"
	TaskStatusCompleted = "Completed"
)

type (
	UserInfo struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	Task struct {
		ID          string    `json:"id"`
		Title       string    `json:"title"`
		JiraID      string    `json:"jira_id"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		AssigneeID  string    `json:"assignee_id"`
		Created     time.Time `json:"created"`
	}
	TaskInfo struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		JiraID     string `json:"jira_id"`
		AssigneeID string `json:"assignee_id"`
	}
)

func (u *UserInfo) ToEntity() *db.Assignee {
	return &db.Assignee{
		ID:       u.ID,
		Username: u.Username,
	}
}

func (t *Task) ValidateCreate() error {
	if t.Title == "" {
		return fmt.Errorf("title field is empty")
	}
	if t.Description == "" {
		return fmt.Errorf("description field is empty")
	}
	return nil
}

func (t *Task) ValidateUpdate() error {
	for _, s := range []string{TaskStatusAssigned, TaskStatusCompleted} {
		if t.Status == s {
			return nil
		}
	}
	return fmt.Errorf("wrong task status: %s", t.Status)
}

// don't allow to change task assignee via REST, only with 'reassign tasks' button
func (t *Task) RemoveAssignee() {
	if t.AssigneeID != "" {
		t.AssigneeID = ""
	}
}

func (m *Task) ToEntity() *db.Task {
	return &db.Task{
		ID:          m.ID,
		Title:       m.Title,
		JiraID:      m.JiraID,
		Description: m.Description,
		Status:      m.Status,
		AssigneeID:  m.AssigneeID,
		Created:     m.Created,
	}
}

func (m *Task) FromEntity(e *db.Task) {
	m.ID = e.ID
	m.Title = e.Title
	m.JiraID = e.JiraID
	m.Description = e.Description
	m.Status = e.Status
	m.AssigneeID = e.AssigneeID
	m.Created = e.Created
}

func TaskEntityToTaskInfo(e *db.Task) *TaskInfo {
	return &TaskInfo{
		ID:         e.ID,
		Title:      e.Title,
		JiraID:     e.JiraID,
		AssigneeID: e.AssigneeID,
	}
}
