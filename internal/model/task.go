package model

import (
	"github.com/google/uuid"
	"time"
)

type Status = int

const (
	status_todo = iota
	status_in_proress
	status_done
)

type Priority = int

const (
	priotity_low = iota
	priotity_normal
	priotity_high
)

type Task struct {
	Id        uuid.UUID
	Title     string
	Content   string
	Status    Status
	Priority  Priority
	Tags      []string
	DueDate   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
