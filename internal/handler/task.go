package handler

import (
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"simple-tasks/internal/model"
	"simple-tasks/internal/service"
	"strconv"
)

var validate *validator.Validate = validator.New()

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
	w.Header().Set("Content-Type", "application/json")

	var newTask model.Task
	if err := json2.NewDecoder(r.Body).Decode(&newTask); err != nil {
		h.log.ErrorContext(r.Context(), "invalid json", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusBadRequest)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorInvalidJson, err))
		return
	}

	if err := validate.Struct(newTask); err != nil {
		h.log.ErrorContext(r.Context(), "invalid task", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorValidation, err))
		return
	}

	createdTask := h.service.CreateTask(r.Context(), &newTask)

	w.Header().Set("Location", fmt.Sprintf("/tasks/%s", createdTask.Id))
	w.WriteHeader(http.StatusCreated)
	_ = json2.NewEncoder(w).Encode(createdTask)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	req := &model.GetTasksRequest{
		Status: query.Get("status"),
		Tags:   query["tag"],
		Q:      query.Get("q"),
		Sort:   query.Get("sort"),
	}

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = &page
		} else {
			h.log.ErrorContext(r.Context(), "invalid page", slog.String("error", err.Error()))
		}
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = &pageSize
		} else {
			h.log.ErrorContext(r.Context(), "invalid pageSize", slog.String("error", err.Error()))
		}
	}

	if err := validate.Struct(req); err != nil {
		h.log.ErrorContext(r.Context(), "invalid request", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorValidation, err))
		return
	}

	tasks := h.service.GetTasks(r.Context(), req)

	w.WriteHeader(http.StatusOK)
	_ = json2.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.log.ErrorContext(r.Context(), "invalid id", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorValidation, err))
		return
	}

	task, err := h.service.GetTaskById(r.Context(), id)
	if errors.Is(err, service.NotFoundError) {
		h.log.ErrorContext(r.Context(), "task not found", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusNotFound)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorNotFound, err))
		return
	}
	if err != nil {
		h.log.ErrorContext(r.Context(), "internal error", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusInternalServerError)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorInternal, err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json2.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.log.ErrorContext(r.Context(), "invalid id", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorValidation, err))
		return
	}

	var req model.UpdateTaskRequest
	if err := json2.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.ErrorContext(r.Context(), "invalid json", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusBadRequest)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorInvalidJson, err))
		return
	}

	if err := validate.Struct(req); err != nil {
		h.log.ErrorContext(r.Context(), "invalid task", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorValidation, err))
		return
	}

	newTask, err := h.service.UpdateTask(r.Context(), id, &req)
	if errors.Is(err, service.NotFoundError) {
		h.log.ErrorContext(r.Context(), "task update failed", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusNotFound)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorNotFound, err))
		return
	}
	if err != nil {
		h.log.ErrorContext(r.Context(), "task update failed", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusInternalServerError)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorInternal, err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json2.NewEncoder(w).Encode(newTask)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.log.ErrorContext(r.Context(), "invalid id", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorValidation, err))
		return
	}

	err = h.service.DeleteTask(r.Context(), id)
	if errors.Is(err, service.NotFoundError) {
		h.log.ErrorContext(r.Context(), "task delete failed", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusNotFound)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorNotFound, err))
		return
	}
	if err != nil {
		h.log.ErrorContext(r.Context(), "internal error", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusInternalServerError)
		_ = json2.NewEncoder(w).Encode(newError(r.Context(), errorInternal, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
