package store

import (
	"errors"
	"github.com/google/uuid"
	"math"
	"simple-tasks/internal/model"
	"slices"
	"sort"
	"strings"
	"sync"
)

type TaskRepository interface {
	SaveTask(*model.Task)
	GetTasks(*model.GetTasksRequest) *model.GetTasksResponse
	GetTaskById(uuid.UUID) (*model.Task, error)
	UpdateTask(*model.Task) error
	DeleteTask(uuid.UUID)
}

type InMemoryTaskRepository struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*model.Task
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		mu:    sync.RWMutex{},
		tasks: make(map[uuid.UUID]*model.Task),
	}
}

// TODO: errors

func (r *InMemoryTaskRepository) SaveTask(task *model.Task) {
	r.mu.Lock()
	r.tasks[task.Id] = task
	r.mu.Unlock()
}

func (r *InMemoryTaskRepository) GetTasks(request *model.GetTasksRequest) *model.GetTasksResponse {
	slices.Sort(request.Tags)
	tasks := make([]*model.Task, 0)

	r.mu.RLock()
	for _, task := range r.tasks {
		if request.Status != "" && task.Status != request.Status {
			continue
		}

		if len(request.Tags) > 0 {
			slices.Sort(task.Tags)
			if !slices.Equal(request.Tags, task.Tags) {
				continue
			}
		}

		if request.Q != "" {
			containsQ := strings.Contains(task.Title, request.Q) ||
				strings.Contains(task.Content, request.Q)
			if !containsQ {
				continue
			}
		}

		tasks = append(tasks, task)
	}
	r.mu.RUnlock()

	total := len(tasks)
	var totalPages *int

	if request.PageSize != nil && request.Page != nil {
		pageSize := *request.PageSize
		page := *request.Page
		if pageSize != 0 {
			totalPagesVal := int(math.Ceil(float64(total) / float64(pageSize)))
			totalPages = &totalPagesVal
		}

		start := (page - 1) * pageSize
		if start > total {
			start = total
		}
		end := start + pageSize
		if end > total {
			end = total
		}

		tasks = tasks[start:end]
	}

	switch request.Sort {
	case model.SortPriority:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority < tasks[j].Priority
		})
	}

	return &model.GetTasksResponse{
		Tasks:      tasks,
		Page:       request.Page,
		PageSize:   request.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

func (r *InMemoryTaskRepository) GetTaskById(id uuid.UUID) (*model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if task, ok := r.tasks[id]; ok {
		return task, nil
	}

	return nil, errors.New("task not found")
}

func (r *InMemoryTaskRepository) UpdateTask(newTask *model.Task) error {
	r.mu.RLock()
	if _, ok := r.tasks[newTask.Id]; !ok {
		r.mu.RUnlock()
		return errors.New("task not found")
	}
	r.mu.RUnlock()

	r.mu.Lock()
	r.tasks[newTask.Id] = newTask
	r.mu.Unlock()

	return nil
}

func (r *InMemoryTaskRepository) DeleteTask(id uuid.UUID) {
	r.mu.Lock()
	delete(r.tasks, id)
	r.mu.Unlock()
}
