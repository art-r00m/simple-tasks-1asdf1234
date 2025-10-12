package service

import (
	"github.com/google/uuid"
	"log/slog"
	"simple-tasks/internal/model"
	"simple-tasks/internal/store"
	"time"
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

func (s *TaskService) CreateTask(t model.Task) (model.Task, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		s.log.Error("Failed to create UUID", "error", err)
		return model.Task{}, err
	}

	t.Id = id
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	t.SetDefaultValues()

	s.repo.SaveTask(t)
	return t, nil
}
