package handler

import (
	json2 "encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"simple-tasks/internal/model"
	"simple-tasks/internal/service"
)

type Response struct {
	Error     string `json:"error,omitempty"`
	RequestId string `json:"requestId"`
}

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
	var newTask model.Task

	if err := json2.NewDecoder(r.Body).Decode(&newTask); err != nil {
		h.log.Error(err.Error())
		_ = json2.NewEncoder(w).Encode(Response{Error: err.Error(), RequestId: "123"})
		return
	}

	if err := validator.New().Struct(newTask); err != nil {
		validateErr := err.(validator.ValidationErrors)
		h.log.Error("invalid request", validateErr.Error())
		_ = json2.NewEncoder(w).Encode(Response{Error: validateErr.Error(), RequestId: "123"})
		return
	}

	createdTask, err := h.service.CreateTask(newTask)
	if err != nil {
		h.log.Error(err.Error())
		_ = json2.NewEncoder(w).Encode(Response{Error: err.Error(), RequestId: "123"})
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/tasks/%s", createdTask.Id))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json2.NewEncoder(w).Encode(createdTask)
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
