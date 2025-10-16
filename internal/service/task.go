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

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "task not found"
}

type InternalError struct{}

func (e *InternalError) Error() string {
	return "internal error"
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

func (s *TaskService) GetTasks(ctx context.Context, request *model.GetTasksRequest) *model.GetTasksResponse {
	return s.repo.GetTasks(request)
}

func (s *TaskService) GetTaskById(ctx context.Context, uuid uuid.UUID) (*model.Task, error) {
	task, err := s.repo.GetTaskById(uuid)
	if errors.Is(err, &store.NotFoundError{}) {
		return nil, &NotFoundError{}
	} else if err != nil {
		return nil, &InternalError{}
	}

	return &task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, request *model.UpdateTaskRequest) (*model.Task, error) {
	task, err := s.GetTaskById(ctx, id)
	if err != nil {
		return nil, err
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

	err = s.repo.UpdateTask(task)
	if errors.Is(err, &store.NotFoundError{}) {
		return nil, &NotFoundError{}
	} else if err != nil {
		return nil, &InternalError{}
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, uuid uuid.UUID) error {
	err := s.repo.DeleteTask(uuid)
	if errors.Is(err, &store.NotFoundError{}) {
		return &NotFoundError{}
	} else if err != nil {
		return &InternalError{}
	}

	return nil
}
