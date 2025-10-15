package handler

import (
	json2 "encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"simple-tasks/internal/model"
	"simple-tasks/internal/service"
	"simple-tasks/internal/store"
	"slices"
	"strings"
	"testing"
)

var log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

var handler = NewTaskHandler(log,
	service.NewTaskService(log, store.NewInMemoryTaskRepository()))

func TestCreateTask(t *testing.T) {
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

func addTasks() []string {
	allTitles := make([]string, 0, len(tasks))
	for _, json := range tasks {
		var task model.Task
		_ = json2.NewDecoder(strings.NewReader(json)).Decode(&task)
		allTitles = append(allTitles, task.Title)
		req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(json))
		w := httptest.NewRecorder()
		handler.CreateTask(w, req)
	}
	slices.Sort(allTitles)
	return allTitles
}

func TestGetTasks(t *testing.T) {
	allTitles := addTasks()

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
