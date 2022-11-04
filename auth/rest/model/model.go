package model

import (
	"fmt"
	"time"

	"github.com/ko3luhbka/auth/db"
)

type (
	User struct {
		ID           string    `json:"id"`
		Username     string    `json:"username"`
		Password     string    `json:"password,omitempty"`
		Role         string    `json:"role"`
		Email        string    `json:"email"`
		LastModified time.Time `json:"last_modified"`
	}
	UserLogin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	Assignee struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
)

func EntityToAssignee(e *db.User) *Assignee {
	return &Assignee{
		ID:       e.ID,
		Username: e.Username,
	}
}

func (m *User) ToEntity() *db.User {
	return &db.User{
		ID:           m.ID,
		Username:     m.Username,
		Password:     m.Password,
		Role:         m.Role,
		Email:        m.Email,
		LastModified: m.LastModified,
	}
}

func (m *User) FromEntity(e *db.User) {
	m.ID = e.ID
	m.Username = e.Username
	m.Role = e.Role
	m.Email = e.Email
	m.LastModified = e.LastModified
}

func (u *User) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("username field is empty")
	}
	if u.Password == "" {
		return fmt.Errorf("password field is empty")
	}
	if u.Role == "" {
		return fmt.Errorf("role field is empty")
	}
	return nil
}

func (u *UserLogin) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("username field is empty")
	}
	if u.Password == "" {
		return fmt.Errorf("password field is empty")
	}
	return nil
}
