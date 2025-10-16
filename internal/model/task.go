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
	DueDate   *time.Time `json:"dueDate,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (t *Task) SetDefaults() {
	if t.Status == "" {
		t.Status = StatusTodo
	}
	if t.Priority == "" {
		t.Priority = PriorityLow
	}
}

type Sort = string

const (
	SortPriority = "priority"
	SortDesc     = "desc"
)

type GetTasksRequest struct {
	Status   string
	Tags     []string
	Q        string
	Sort     Sort `validate:"omitempty,oneof=priority desc"`
	Page     *int `validate:"omitempty,gte=0"`
	PageSize *int `validate:"omitempty,gte=1,lte=100"`
}

type GetTasksResponse struct {
	Tasks      []Task `json:"items,omitempty"`
	Page       *int   `json:"page,omitempty"`
	PageSize   *int   `json:"pageSize,omitempty"`
	Total      int    `json:"total"`
	TotalPages *int   `json:"totalPages,omitempty"`
}

type UpdateTaskRequest struct {
	Title    string     `json:"title,omitempty" validate:"omitempty,gte=1,lte=200"`
	Content  string     `json:"content,omitempty" validate:"omitempty,lte=5000"`
	Status   Status     `json:"status,omitempty" validate:"omitempty,oneof=todo in_progress done"`
	Priority Priority   `json:"priority,omitempty" validate:"omitempty,oneof=low normal high"`
	Tags     []string   `json:"tags,omitempty" validate:"omitempty,lte=10,dive,gte=1,lte=32"`
	DueDate  *time.Time `json:"dueDate,omitempty"`
}
