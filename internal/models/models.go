package models

import (
	"time"
)

// Seat represents a theater seat
type Seat struct {
	ID     int    `json:"id"`
	Row    string `json:"row"`
	Number int    `json:"number"`
	Type   string `json:"type"` // "regular", "vip", "sponsored"
	Status string `json:"status"` // "available", "booked", "not available"
}

// Booking represents a customer booking
type Booking struct {
	ID            int     `json:"id"`
	CustomerName  string  `json:"customerName"`
	CustomerEmail string  `json:"customerEmail"`
	CustomerPhone string  `json:"customerPhone"`
	PaymentMethod string  `json:"paymentMethod"` // "bank_transfer", "cash"
	TotalAmount   float64 `json:"totalAmount"`
	Status        string  `json:"status"` // "pending", "confirmed", "cancelled"
	CreatedAt     string  `json:"createdAt"`
}

// BookingRequest represents an incoming booking request
type BookingRequest struct {
	CustomerName  string  `json:"customerName"`
	CustomerEmail string  `json:"customerEmail"`
	CustomerPhone string  `json:"customerPhone"`
	PaymentMethod string  `json:"paymentMethod"`
	SeatIDs       []int   `json:"seatIds"`
	TotalAmount   float64 `json:"totalAmount"`
}

// BookingResponse represents a response with booking details
type BookingResponse struct {
	ID            int       `json:"id"`
	CustomerName  string    `json:"customerName"`
	CustomerEmail string    `json:"customerEmail"`
	CustomerPhone string    `json:"customerPhone"`
	PaymentMethod string    `json:"paymentMethod"`
	TotalAmount   float64   `json:"totalAmount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	Seats         []string  `json:"seats"`
}

// AdminLoginRequest represents an admin login request
type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// VerifyPaymentRequest represents a payment verification request
type VerifyPaymentRequest struct {
	BookingID int    `json:"bookingId"`
	Status    string `json:"status"` // "confirmed", "cancelled"
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}