package handler

import (
	"log/slog"
	"net/http"
	"simple-tasks/internal/service"
)

type TaskHandler struct {
	log     *slog.Logger
	service *service.TaskService
}

func NewTaskHandler(log *slog.Logger, service *service.TaskService) *TaskHandler {
	return &TaskHandler{
		log:     log,
		service: service,
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	h.log.Info("CreateTask")
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	h.log.Info("GetTasks")
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	h.log.Info("GetTaskById")
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	h.log.Info("UpdateTask")
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	h.log.Info("DeleteTask")
}
