package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erickgnclvs/go-task-viewer/internal/handlers"
)

func main() {
	log.Println("Starting Go Task Viewer application...")

	// Path relative to the project root directory (where 'go run' is executed)
	tmplPath := "cmd/server/templates/index.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Fatalf("Error loading template from %s: %v", tmplPath, err)
	}
	log.Printf("Template loaded successfully from %s.", tmplPath)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Serve static files (CSS, JS) from templates directory relative to project root
	staticFilesPath := "cmd/server/templates"
	fs := http.FileServer(http.Dir(staticFilesPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register handlers from the handlers package
	mux.HandleFunc("/", handlers.HomeHandler(tmpl))
	mux.HandleFunc("/analyze", handlers.AnalyzeHandler(tmpl))
	mux.HandleFunc("/health", handlers.HealthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received
	log.Println("Shutting down server...")

	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
