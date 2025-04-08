package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"stratifix/internal/database"
	"stratifix/internal/handlers"
)

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create handler with database connection
	h := handlers.NewHandler(db)

	// Special route for admin page
	http.HandleFunc("/poggi", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/admin.html")
	})

	// Serve static files
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/seats", h.GetSeats)
	http.HandleFunc("/api/book", h.BookSeats)
	http.HandleFunc("/api/admin/login", h.AdminLogin)
	http.HandleFunc("/api/admin/bookings", h.GetBookings)
	http.HandleFunc("/api/admin/verify", h.VerifyPayment)

	go func() {
        ticker := time.NewTicker(1 * time.Hour) // Check every hour
        defer ticker.Stop()
        
        for range ticker.C {
			if err := db.ReleaseExpiredBookings(); err != nil {
				log.Printf("Error releasing expired bookings: %v", err)
			}
		}
    }()

	// Get port from environment variable (for Railway)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for local development
	}

	// Start server
	fmt.Printf("Server running on port %s\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Server error: ", err)
	}
}