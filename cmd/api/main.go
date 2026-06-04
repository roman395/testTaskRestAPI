package api

import (
	"api/internal/database"
	"api/internal/handlers"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Println("Starting Subscription Service")
	log.Println("Version: 1.0.0")

	// Get configuration from environment variables
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
		log.Printf("SERVER_PORT not set, using default: %s", serverPort)
	} else {
		log.Printf("SERVER_PORT loaded: %s", serverPort)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://subscriptionuser:subscriptionpass@localhost:5432/subscriptiondb?sslmode=disable"
		log.Println("DATABASE_URL not set, using default connection string")
	} else {
		log.Println("DATABASE_URL loaded from environment")
	}

	// Log configuration summary
	log.Println("Configuration loaded:")
	log.Printf("  Server Port: %s", serverPort)
	log.Printf("  Database Host: %s", extractDBHost(databaseURL))
	log.Printf("  Database Name: %s", extractDBName(databaseURL))

	// Connect to database
	log.Println("Connecting to database...")
	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()

	log.Println("Database connection established successfully")

	// Test database connection
	log.Println("Testing database connection...")
	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("Database ping successful")

	// Initialize store and handlers
	log.Println("Initializing subscription store...")
	subscriptionStore := database.NewSubscriptionStore(db)

	log.Println("Initializing HTTP handlers...")
	handler := handlers.NewHandlers(subscriptionStore)

	// Setup routes
	log.Println("Setting up HTTP routes...")
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})
	log.Println("  Registered: GET /health")

	// API routes
	mux.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ROUTE] /subscriptions - Method: %s, Remote: %s", r.Method, r.RemoteAddr)
		switch r.Method {
		case http.MethodGet:
			handler.GetAllSubscriptions(w, r)
		case http.MethodPost:
			handler.CreateSubscription(w, r)
		default:
			log.Printf("[ROUTE] Method not allowed: %s for /subscriptions", r.Method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
	log.Println("  Registered: GET /subscriptions")
	log.Println("  Registered: POST /subscriptions")

	// Summary endpoint (must be before /subscriptions/ to avoid conflict)
	mux.HandleFunc("/subscriptions/summary", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ROUTE] /subscriptions/summary - Method: %s, Remote: %s", r.Method, r.RemoteAddr)
		if r.Method != http.MethodGet {
			log.Printf("[ROUTE] Method not allowed: %s for /subscriptions/summary", r.Method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		handler.GetTotalSpent(w, r)
	})
	log.Println("  Registered: GET /subscriptions/summary")

	// Individual subscription endpoints
	mux.HandleFunc("/subscriptions/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ROUTE] /subscriptions/{id} - Method: %s, Remote: %s, Path: %s", r.Method, r.RemoteAddr, r.URL.Path)

		// Skip if it's the summary endpoint
		if r.URL.Path == "/subscriptions/summary" {
			return
		}

		switch r.Method {
		case http.MethodGet:
			handler.GetSubscriptionByID(w, r)
		case http.MethodPut:
			handler.UpdateSubscription(w, r)
		case http.MethodDelete:
			handler.DeleteSubscription(w, r)
		default:
			log.Printf("[ROUTE] Method not allowed: %s for %s", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
	log.Println("  Registered: GET /subscriptions/{id}")
	log.Println("  Registered: PUT /subscriptions/{id}")
	log.Println("  Registered: DELETE /subscriptions/{id}")

	// Apply logging middleware
	loggedMux := loggingMiddleware(mux)

	// Create server with timeouts
	serverAddr := ":" + serverPort
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      loggedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	log.Printf("Server starting on http://0.0.0.0%s", serverAddr)
	log.Println("Waiting for requests...")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// loggingMiddleware logs all incoming HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request start
		log.Printf("[REQUEST] START - Method: %s, Path: %s, Remote: %s, User-Agent: %s",
			r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		// Create response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(rw, r)

		// Log request completion
		duration := time.Since(start)
		log.Printf("[REQUEST] END - Method: %s, Path: %s, Status: %d, Duration: %v",
			r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper functions to extract info from database URL
func extractDBHost(dbURL string) string {
	// Simple extraction for logging
	if len(dbURL) > 0 {
		// postgres://user:pass@host:port/db?sslmode=disable
		start := 0
		for i := 0; i < len(dbURL); i++ {
			if dbURL[i] == '@' && i+1 < len(dbURL) {
				start = i + 1
				break
			}
		}
		end := start
		for end < len(dbURL) && dbURL[end] != ':' && dbURL[end] != '/' {
			end++
		}
		if end > start {
			return dbURL[start:end]
		}
	}
	return "localhost"
}

func extractDBName(dbURL string) string {
	// Simple extraction for logging
	lastSlash := -1
	for i := len(dbURL) - 1; i >= 0; i-- {
		if dbURL[i] == '/' {
			lastSlash = i
			break
		}
	}
	if lastSlash >= 0 && lastSlash+1 < len(dbURL) {
		end := lastSlash + 1
		for end < len(dbURL) && dbURL[end] != '?' && dbURL[end] != '&' && dbURL[end] != '/' {
			end++
		}
		if end > lastSlash+1 {
			return dbURL[lastSlash+1 : end]
		}
	}
	return "subscriptiondb"
}
