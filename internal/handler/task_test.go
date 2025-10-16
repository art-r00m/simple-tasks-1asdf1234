package handler

import (
	json2 "encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"simple-tasks/internal/model"
	"simple-tasks/internal/service"
	"simple-tasks/internal/store"
	"slices"
	"sort"
	"strings"
	"testing"
)

func createTestHandler() *TaskHandler {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	repo := store.NewInMemoryTaskRepository()
	taskService := service.NewTaskService(log, repo)
	handler := NewTaskHandler(log, taskService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks/{id}", handler.GetTaskById)

	return handler
}

func TestCreateTask(t *testing.T) {
	handler := createTestHandler()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "valid task",
			requestBody:    `{"title":"Test task"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing title",
			requestBody:    `{"content":"No title"}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "invalid json",
			requestBody:    `{"content No title"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid status",
			requestBody:    `{"title":"Test task", "status":"qwerty"}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "invalid priority",
			requestBody:    `{"title":"Test task", "status":"my_priority"}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()
			handler.CreateTask(w, req)

			resp := w.Result()
			_, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.StatusCode)
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}

var tasks = []string{
	`{"title":"Купить молоко","content":"молоко","status":"todo","tags":["покупки","todo_tag"],"priority":"low"}`,
	`{"title":"Купить машину", "content":"машину", "status":"in_progress", "tags":["покупки"], "priority":"high"}`,
	`{"title":"Погулять с собакой", "content":"погулять", "status":"done", "tags":["прогулка"], "priority":"normal"}`,
	`{"title":"Погулять с друзьями", "content":"погулять", "status":"todo", "tags":["прогулка","todo_tag"], "priority":"normal"}`,
	`{"title":"Сходить в спортзал", "content":"тренировка ног", "status":"todo", "tags":["спорт","здоровье","рутина"], "priority":"normal"}`,
	`{"title":"Подготовить отчет", "content":"ежеквартальный отчет", "status":"in_progress", "tags":["работа","отчетность","финансы"], "priority":"high"}`,
	`{"title":"Уборка дома", "content":"генеральная уборка", "status":"todo"}`,
}

func addTasks(handler *TaskHandler) []model.Task {
	allTasks := make([]model.Task, 0, len(tasks))
	for _, json := range tasks {
		req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(json))
		w := httptest.NewRecorder()
		handler.CreateTask(w, req)

		var createdTask model.Task
		resp := w.Result()
		_ = json2.NewDecoder(resp.Body).Decode(&createdTask)
		allTasks = append(allTasks, createdTask)
	}
	sort.Slice(allTasks, func(i, j int) bool {
		return allTasks[i].Title < allTasks[j].Title
	})
	return allTasks
}

func TestGetTasks(t *testing.T) {
	handler := createTestHandler()

	allTasks := addTasks(handler)
	allTitles := make([]string, 0, len(allTasks))
	for _, task := range allTasks {
		allTitles = append(allTitles, task.Title)
	}

	tests := []struct {
		name               string
		query              string
		expectedTaskTitles []string
		expectedStatus     int
	}{
		{
			name:               "all tasks",
			query:              "",
			expectedTaskTitles: allTitles,
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "filter status",
			query:              "?status=done",
			expectedTaskTitles: []string{"Погулять с собакой"},
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "filter tags",
			query:              "?tag=покупки",
			expectedTaskTitles: []string{"Купить молоко", "Купить машину"},
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "filter q",
			query:              "?q=погулять",
			expectedTaskTitles: []string{"Погулять с друзьями", "Погулять с собакой"},
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "combined filter",
			query:              "?status=todo&tag=todo_tag",
			expectedTaskTitles: []string{"Погулять с друзьями", "Купить молоко"},
			expectedStatus:     http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tasks%s", tt.query), nil)
			w := httptest.NewRecorder()
			handler.GetTasks(w, req)
			resp := w.Result()

			var response model.GetTasksResponse
			err := json2.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}
			actualTitles := make([]string, len(response.Tasks))
			for i, task := range response.Tasks {
				actualTitles[i] = task.Title
			}

			slices.Sort(actualTitles)
			slices.Sort(tt.expectedTaskTitles)

			if !slices.Equal(tt.expectedTaskTitles, actualTitles) {
				t.Errorf("expected tasks %v, got %v", tt.expectedTaskTitles, actualTitles)
			}
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.StatusCode)
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestGetTaskById(t *testing.T) {
	handler := createTestHandler()

	allTasks := addTasks(handler)

	tests := []struct {
		name           string
		id             string
		expectedTaskId string
		expectedStatus int
	}{
		{
			name:           "success get task by id",
			id:             allTasks[0].Id.String(),
			expectedTaskId: allTasks[0].Id.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			id:             uuid.New().String(),
			expectedTaskId: uuid.Nil.String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/tasks/", nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()
			handler.GetTaskById(w, req)
			resp := w.Result()

			var actualTask model.Task
			err := json2.NewDecoder(resp.Body).Decode(&actualTask)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}

			if actualTask.Id.String() != tt.expectedTaskId {
				t.Errorf("expected task id %v, got %v", tt.expectedTaskId, actualTask.Id)
			}
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.StatusCode)
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	handler := createTestHandler()

	allTasks := addTasks(handler)

	tests := []struct {
		name           string
		updateBody     string
		taskIdToUpdate string
		expectedStatus int
	}{
		{
			name:           "update task by id",
			updateBody:     `{"status":"done","tags":["tag1","tag2"]}`,
			taskIdToUpdate: allTasks[0].Id.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found task by id",
			updateBody:     `{"status":"done","tags":["tag1","tag2"]}`,
			taskIdToUpdate: uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "validation error",
			updateBody:     `{"status":"done","tags":["tag1erkljhklejrhilojreljkhlkjrtjklr5j","tag2"]}`,
			taskIdToUpdate: uuid.New().String(),
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updateRequest model.UpdateTaskRequest
			_ = json2.NewDecoder(strings.NewReader(tt.updateBody)).Decode(&updateRequest)

			req := httptest.NewRequest(http.MethodPatch, "/tasks/", strings.NewReader(tt.updateBody))
			req.SetPathValue("id", tt.taskIdToUpdate)
			w := httptest.NewRecorder()
			handler.UpdateTask(w, req)
			resp := w.Result()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.StatusCode)
			} else if resp.StatusCode != http.StatusOK {
				return
			}

			var actualTask model.Task
			err := json2.NewDecoder(resp.Body).Decode(&actualTask)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}

			if updateRequest.Title != "" {
				if actualTask.Title != updateRequest.Title {
					t.Errorf("expected title %v, got %v", updateRequest.Title, actualTask.Title)
				}
			}
			if updateRequest.Content != "" {
				if actualTask.Content != updateRequest.Content {
					t.Errorf("expected content %v, got %v", updateRequest.Content, actualTask.Content)
				}
			}
			if updateRequest.Status != "" {
				if actualTask.Status != updateRequest.Status {
					t.Errorf("expected status %v, got %v", updateRequest.Status, actualTask.Status)
				}
			}
			if len(updateRequest.Tags) > 0 {
				slices.Sort(updateRequest.Tags)
				slices.Sort(actualTask.Tags)
				if !slices.Equal(updateRequest.Tags, actualTask.Tags) {
					t.Errorf("expected tags %v, got %v", updateRequest.Tags, actualTask.Tags)
				}
			}
			if updateRequest.Priority != "" {
				if actualTask.Priority != updateRequest.Priority {
					t.Errorf("expected priority %v, got %v", updateRequest.Priority, actualTask.Priority)
				}
			}
			if updateRequest.DueDate != nil {
				if actualTask.DueDate != updateRequest.DueDate {
					t.Errorf("expected due date %v, got %v", updateRequest.DueDate, actualTask.DueDate)
				}
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	handler := createTestHandler()

	allTasks := addTasks(handler)

	tests := []struct {
		name           string
		taskIdToDelete string
		expectedStatus int
	}{
		{
			name:           "delete task by id",
			taskIdToDelete: allTasks[0].Id.String(),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "not found task by id",
			taskIdToDelete: uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodPatch, "/tasks/", nil)
			req.SetPathValue("id", tt.taskIdToDelete)
			w := httptest.NewRecorder()
			handler.DeleteTask(w, req)
			resp := w.Result()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.StatusCode)
			} else if resp.StatusCode != http.StatusOK {
				return
			}
		})
	}
}
