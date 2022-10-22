package main

import (
	"fmt"

	"github.com/ko3luhbka/popug_schema_registry/validator"
)

var JsonData = `{"foo":"bar", "bar":"baz"}`

type TaskEvent struct {
	Name string         `json:"name"`
	Version int `json:"version"`
	Data TaskInfo `json:"data"`
}

type TaskInfo struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	JiraID     string `json:"jira_id"`
	AssigneeID string `json:"assignee_id"`
}

var foobar = TaskEvent{
	// Name: "taskReassigned",
	Name: "taskCreated",
	Version: 2,
	Data: TaskInfo{
		ID: "51c0f139-d529-4035-9436-04b9fa0ad026",
		Title: "bar",
		JiraID: "f",
		AssigneeID: "vova",
	},
}

func main() {
	if err := validator.Validate(foobar, "task", 2); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("schema is valid")
	}
}
