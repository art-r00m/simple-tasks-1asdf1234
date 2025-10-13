package service

import (
	"context"
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

func (s *TaskService) CreateTask(ctx context.Context, t *model.Task) *model.Task {
	t.Id = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	t.SetDefaults()

	s.repo.SaveTask(t)

	return t
}

// TODO: Using DTO of model layer looks cringe
func (s *TaskService) GetTasks(ctx context.Context, request *model.GetTasksRequest) *model.GetTasksResponse {
	return s.repo.GetTasks(request)
}
