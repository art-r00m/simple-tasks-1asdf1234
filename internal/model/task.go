package model

import (
	"github.com/google/uuid"
	"time"
)

type Status = string

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_proress"
	StatusDone       = "done"
)

type Priority = string

const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
)

type Task struct {
	Id        uuid.UUID  `json:"id"`
	Title     string     `json:"title" validate:"required,gte=1,lte=200"`
	Content   string     `json:"content" validate:"lte=5000"`
	Status    Status     `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority  Priority   `json:"priority" validate:"omitempty,oneof=low normal high"`
	Tags      []string   `json:"tags" validate:"lte=10,dive,gte=1,lte=32"`
	DueDate   *time.Time `json:"dueDate"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (t *Task) SetDefaultValues() {
	if t.Status == "" {
		t.Status = StatusTodo
	}
	if t.Priority == "" {
		t.Priority = PriorityLow
	}
}
