package handlers

import (
	"encoding/json"
	"net/http"
	"stratifix/internal/database"
	"stratifix/internal/models"
	"time"
)

type Handler struct {
	db *database.Database
}

func NewHandler(db *database.Database) *Handler {
	return &Handler{db: db}
}

// GetSeats returns all seats
func (h *Handler) GetSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	seats, err := h.db.GetAllSeats()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving seats")
		return
	}

	respondWithJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    seats,
	})
}

// BookSeats handles seat booking
func (h *Handler) BookSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var bookingReq models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&bookingReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if bookingReq.CustomerName == "" || bookingReq.CustomerEmail == "" || bookingReq.CustomerPhone == "" {
		respondWithError(w, http.StatusBadRequest, "Missing customer information")
		return
	}

	if len(bookingReq.SeatIDs) == 0 {
		respondWithError(w, http.StatusBadRequest, "No seats selected")
		return
	}

	// Create booking object
	booking := models.Booking{
		CustomerName:  bookingReq.CustomerName,
		CustomerEmail: bookingReq.CustomerEmail,
		CustomerPhone: bookingReq.CustomerPhone,
		PaymentMethod: bookingReq.PaymentMethod,
		TotalAmount:   bookingReq.TotalAmount,
		Status:        "pending",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	// Book seats
	bookingID, err := h.db.BookSeats(booking, bookingReq.SeatIDs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error booking seats")
		return
	}

	// Return booking confirmation
	respondWithJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Booking successful",
		Data: map[string]interface{}{
			"bookingId": bookingID,
		},
	})
}

// AdminLogin handles admin login
func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq models.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Verify admin credentials
	isAdmin, err := h.db.VerifyAdmin(loginReq.Username, loginReq.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error verifying credentials")
		return
	}

	if !isAdmin {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// In a real app, you would generate a JWT token here
	respondWithJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data: map[string]bool{
			"authenticated": true,
		},
	})
}

// GetBookings returns all pending bookings
func (h *Handler) GetBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In a real app, you would verify the JWT token here

	bookings, err := h.db.GetPendingBookings()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving bookings")
		return
	}

	respondWithJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    bookings,
	})
}

// VerifyPayment updates booking status
func (h *Handler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In a real app, you would verify the JWT token here

	var verifyReq models.VerifyPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&verifyReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Update booking status
	err := h.db.UpdateBookingStatus(verifyReq.BookingID, verifyReq.Status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating booking status")
		return
	}

	respondWithJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Booking status updated successfully",
	})
}

// Helper functions for JSON responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, models.APIResponse{
		Success: false,
		Message: message,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}