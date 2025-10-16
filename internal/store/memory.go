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

var NotFoundError = errors.New("task not found")

type TaskRepository interface {
	SaveTask(*model.Task)
	GetTasks(*model.GetTasksRequest) *model.GetTasksResponse
	GetTaskById(uuid.UUID) (model.Task, error)
	UpdateTask(*model.Task) error
	DeleteTask(uuid.UUID) error
}

type InMemoryTaskRepository struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]model.Task
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		mu:    sync.RWMutex{},
		tasks: make(map[uuid.UUID]model.Task),
	}
}

func (r *InMemoryTaskRepository) SaveTask(task *model.Task) {
	r.mu.Lock()
	r.tasks[task.Id] = *task
	r.mu.Unlock()
}

func (r *InMemoryTaskRepository) GetTasks(request *model.GetTasksRequest) *model.GetTasksResponse {
	slices.Sort(request.Tags)
	tasks := make([]model.Task, 0)

	r.mu.RLock()
	for _, task := range r.tasks {
		if request.Status != "" && task.Status != request.Status {
			continue
		}

		if len(request.Tags) > 0 {
			tagsCopy := make([]string, len(task.Tags))
			copy(tagsCopy, task.Tags)
			slices.Sort(tagsCopy)
			containsTag := slices.ContainsFunc(request.Tags, func(requestTag string) bool {
				return slices.Contains(tagsCopy, requestTag)
			})
			if !containsTag {
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

func (r *InMemoryTaskRepository) GetTaskById(id uuid.UUID) (model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if task, ok := r.tasks[id]; ok {
		return task, nil
	}

	return model.Task{}, NotFoundError
}

func (r *InMemoryTaskRepository) UpdateTask(newTask *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[newTask.Id]; !ok {
		return NotFoundError
	}

	r.tasks[newTask.Id] = *newTask

	return nil
}

func (r *InMemoryTaskRepository) DeleteTask(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return NotFoundError
	}

	delete(r.tasks, id)

	return nil
}
