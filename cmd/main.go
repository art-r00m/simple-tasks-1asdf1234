package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"simple-tasks/internal/handler"
	"simple-tasks/internal/middleware"
	"simple-tasks/internal/service"
	"simple-tasks/internal/store"
	"strconv"
)

type config struct {
	port int
}

func getConfig() config {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		println("invalid port: %v", err)
		os.Exit(1)
	}

	return config{
		port: port,
	}
}

func (c *config) String() string {
	return fmt.Sprintf("port: %d", c.port)
}

func main() {
	config := getConfig()

	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	requestIdHandler := middleware.NewHandlerMiddleware(textHandler)
	log := slog.New(requestIdHandler)
	log.Info("starting server", slog.String("config", config.String()))

	taskRepo := store.NewInMemoryTaskRepository()
	taskService := service.NewTaskService(log, taskRepo)
	taskHandler := handler.NewTaskHandler(log, taskService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /tasks", taskHandler.GetTasks)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.GetTaskById)
	mux.HandleFunc("PATCH /tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTask)

	logMiddleware := func(h http.Handler) http.Handler {
		return middleware.LogMiddleware(log, h)
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.port),
		Handler: middleware.RequestIdMiddleware(logMiddleware(mux)),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("%v", err)
	}
}
