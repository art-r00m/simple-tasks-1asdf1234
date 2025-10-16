package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"simple-tasks/internal/config"
	"simple-tasks/internal/handler"
	"simple-tasks/internal/middleware"
	"simple-tasks/internal/service"
	"simple-tasks/internal/store"
	"sync"
	"syscall"
	"time"
)

var (
	Log  *slog.Logger
	once sync.Once
)

func GetLog() *slog.Logger {
	once.Do(func() {
		textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		requestIdHandler := middleware.NewHandlerMiddleware(textHandler)
		Log = slog.New(requestIdHandler)
	})

	return Log
}

func main() {
	cfg := config.GetConfig()

	log := GetLog()
	log.Info("starting server", slog.String("config", cfg.String()))

	taskRepo := store.NewInMemoryTaskRepository()
	taskService := service.NewTaskService(log, taskRepo)
	taskHandler := handler.NewTaskHandler(log, taskService)

	mux := http.NewServeMux()
	mux.HandleFunc(http.MethodPost+" /tasks", taskHandler.CreateTask)
	mux.HandleFunc(http.MethodGet+" /tasks", taskHandler.GetTasks)
	mux.HandleFunc(http.MethodGet+" /tasks/{id}", taskHandler.GetTaskById)
	mux.HandleFunc(http.MethodPatch+" /tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc(http.MethodDelete+" /tasks/{id}", taskHandler.DeleteTask)

	logMiddleware := func(h http.Handler) http.Handler {
		return middleware.LogMiddleware(log, h)
	}
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           middleware.RequestIdMiddleware(logMiddleware(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error %v", slog.String("error", err.Error()))
		}
		log.Info("Stopped serving new connections")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP shutdown error: %v", slog.String("error", err.Error()))
	}
	log.Info("Graceful shutdown complete")
}
