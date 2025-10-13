package service

import (
	"context"
	"errors"
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

func (s *TaskService) GetTaskById(ctx context.Context, uuid uuid.UUID) (*model.Task, error) {
	return s.repo.GetTaskById(uuid)
}

func (s *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, request *model.UpdateTaskRequest) (*model.Task, error) {
	task, err := s.GetTaskById(ctx, id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	if request.Title != "" {
		task.Title = request.Title
	}
	if request.Content != "" {
		task.Content = request.Content
	}
	if request.Status != "" {
		task.Status = request.Status
	}
	if request.Priority != "" {
		task.Priority = request.Priority
	}
	if len(request.Tags) != 0 {
		task.Tags = request.Tags
	}
	if request.DueDate != nil {
		task.DueDate = request.DueDate
	}

	task.UpdatedAt = time.Now()

	if err := s.repo.UpdateTask(task); err != nil {
		return nil, err
	}
	return task, nil
}
