package main

import (
	"index-duel-backend/database"
	"index-duel-backend/handlers"
	"index-duel-backend/repository"
	"index-duel-backend/scheduler"
	"index-duel-backend/service"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Connected to database successfully")

	// Initialize repositories
	cardRepo := repository.NewCardRepository(db)

	// Initialize services
	cardService := service.NewCardService(cardRepo)

	// Initialize handlers
	cardHandler := handlers.NewCardHandler(cardService)

	// Initialize and start the weekly scheduler
	cardScheduler := scheduler.NewScheduler(cardService)
	cardScheduler.Start()

	log.Println("Weekly card synchronization scheduler initialized")

	// Setup routes
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", cardHandler.HealthCheckHandler).Methods("GET")

	// Main endpoint for mobile app synchronization
	api.HandleFunc("/cards/sync", cardHandler.SyncCardsForMobileHandler).Methods("POST")

	// Add CORS middleware
	router.Use(corsMiddleware)

	// Add logging middleware
	router.Use(loggingMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	log.Printf("Health check available at: http://localhost:%s/api/v1/health", port)
	log.Printf("Mobile sync endpoint: POST http://localhost:%s/api/v1/cards/sync", port)
	log.Printf("Weekly synchronization with Yu-Gi-Oh API enabled")

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
