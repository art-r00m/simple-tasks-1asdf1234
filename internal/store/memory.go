package store

import (
	"errors"
	"github.com/google/uuid"
	"simple-tasks/internal/model"
	"sync"
)

type TaskRepository interface {
	SaveTask(model.Task)
	GetTasks() []model.Task
	GetTaskById(uuid.UUID) (model.Task, error)
	UpdateTask(model.Task) error
	DeleteTask(uuid.UUID)
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

// TODO: errors

func (r *InMemoryTaskRepository) SaveTask(task model.Task) {
	r.mu.Lock()
	r.tasks[task.Id] = task
	r.mu.Unlock()
}

func (r *InMemoryTaskRepository) GetTasks() []model.Task {
	tasks := make([]model.Task, 0, len(r.tasks))
	r.mu.RLock()
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	r.mu.RUnlock()

	return tasks
}

func (r *InMemoryTaskRepository) GetTaskById(id uuid.UUID) (model.Task, error) {
	r.mu.RLock()
	if task, ok := r.tasks[id]; ok {
		return task, nil
	}
	r.mu.RUnlock()

	return model.Task{}, errors.New("task not found")
}

func (r *InMemoryTaskRepository) UpdateTask(newTask model.Task) error {
	r.mu.RLock()
	if _, ok := r.tasks[newTask.Id]; !ok {
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
