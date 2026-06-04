package main

import (
	"api/config"
	"api/internal/database"
	"api/internal/handlers"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	log.Println("Starting Subscription Service")
	log.Println("Version: 1.0.0")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Configuration loaded successfully")
	log.Printf("Server will listen on: %s", cfg.ServerAddress())
	log.Printf("Environment: %s", getEnv("APP_ENV", "development"))

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()

	// Initialize store and handlers
	subscriptionStore := database.NewSubscriptionStore(db)
	handler := handlers.NewHandlers(subscriptionStore)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(cfg.ReadTimeoutDuration()))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes
	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/", handler.GetAllSubscriptions)
		r.Post("/", handler.CreateSubscription)
		r.Get("/summary", handler.GetTotalSpent)
		r.Get("/{id}", handler.GetSubscriptionByID)
		r.Put("/{id}", handler.UpdateSubscription)
		r.Delete("/{id}", handler.DeleteSubscription)
	})

	// Start server
	server := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeoutDuration(),
		WriteTimeout: cfg.WriteTimeoutDuration(),
		IdleTimeout:  cfg.IdleTimeoutDuration(),
	}

	log.Printf("Server starting on http://%s", cfg.ServerAddress())
	log.Println("Waiting for requests...")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
