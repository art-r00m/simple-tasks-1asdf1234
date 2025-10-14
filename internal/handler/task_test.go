package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"simple-tasks/internal/service"
	"simple-tasks/internal/store"
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
			expectedStatus: 201,
		},
		{
			name:           "missing title",
			requestBody:    `{"content":"No title"}`,
			expectedStatus: 422,
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
			//if resp.Header.Get("Content-Type") != "application/json" {
			//	t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			//}
		})
	}
}
