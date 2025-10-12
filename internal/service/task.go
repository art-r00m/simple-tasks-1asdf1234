package service

import (
	"log/slog"
	"simple-tasks/internal/store"
)

type TaskService struct {
	repo store.TaskRepository
	log  *slog.Logger
}

func NewTaskService(log *slog.Logger, repo store.TaskRepository) *TaskService {
	return &TaskService{
		log:  log,
		repo: repo,
	}
}
