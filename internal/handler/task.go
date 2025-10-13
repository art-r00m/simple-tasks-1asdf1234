package handler

import (
	"context"
	json2 "encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"simple-tasks/internal/middleware"
	"simple-tasks/internal/model"
	"simple-tasks/internal/service"
	"strconv"
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

func newErrorResponse(ctx context.Context, err error) *Response {
	return &Response{Error: err.Error(), RequestId: ctx.Value(middleware.RequestId).(string)}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var newTask model.Task

	if err := json2.NewDecoder(r.Body).Decode(&newTask); err != nil {
		h.log.ErrorContext(r.Context(), "invalid json", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusBadRequest)
		_ = json2.NewEncoder(w).Encode(newErrorResponse(r.Context(), err))
		return
	}

	if err := validator.New().Struct(newTask); err != nil {
		validateErr := err.(validator.ValidationErrors)
		h.log.ErrorContext(r.Context(), "invalid task", slog.String("error", validateErr.Error()))

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newErrorResponse(r.Context(), err))
		return
	}

	createdTask := h.service.CreateTask(context.TODO(), &newTask)

	w.Header().Set("Location", fmt.Sprintf("/tasks/%s", createdTask.Id))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json2.NewEncoder(w).Encode(createdTask)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	req := &model.GetTasksRequest{
		Status: query.Get("status"),
		Tags:   query["tags"],
		Q:      query.Get("q"),
		Sort:   query.Get("sort"),
	}

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = &page
		} else {
			h.log.ErrorContext(r.Context(), "invalid page", err.Error())
		}
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = &pageSize
		} else {
			h.log.ErrorContext(r.Context(), "invalid pageSize", err.Error())
		}
	}

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)
		h.log.ErrorContext(r.Context(), "invalid request", validateErr.Error())

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newErrorResponse(r.Context(), err))
		return
	}

	tasks := h.service.GetTasks(context.TODO(), req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json2.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.log.ErrorContext(r.Context(), "invalid id", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newErrorResponse(r.Context(), err))
		return
	}

	task, err := h.service.GetTaskById(context.TODO(), id)
	if err != nil {
		h.log.ErrorContext(r.Context(), "task not found", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusNotFound)
		_ = json2.NewEncoder(w).Encode(newErrorResponse(r.Context(), err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json2.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	h.log.Info("UpdateTask")
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	h.log.Info("DeleteTask")
}
