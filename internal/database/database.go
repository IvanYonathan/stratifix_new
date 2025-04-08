package database

import (
	"database/sql"
	"fmt"
	"os"
	"stratifix/internal/models"
	
	_ "github.com/lib/pq"
)

type Database struct {
	*sql.DB
}

func InitDB() (*Database, error) {
	// Get database URL from environment variable (for Railway deployment)
	// Falls back to local development connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Local development connection
		dbURL = "postgres://postgres:1234@localhost:5432/stratifix?sslmode=disable"
	}
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	
	// Create tables if they don't exist
	if err = createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}
	
	// Initialize admin user if it doesn't exist
	if err = initAdminUser(db); err != nil {
		return nil, fmt.Errorf("failed to initialize admin user: %v", err)
	}
	
	// Initialize seats if they don't exist
	if err = initSeats(db); err != nil {
		return nil, fmt.Errorf("failed to initialize seats: %v", err)
	}
	
	return &Database{db}, nil
}

func createTables(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(100) NOT NULL,
			is_admin BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	if err != nil {
		return err
	}
	
	// Create seats table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS seats (
			id SERIAL PRIMARY KEY,
			row CHAR(1) NOT NULL,
			number INT NOT NULL,
			type VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'available',
			UNIQUE (row, number)
		)
	`)
	if err != nil {
		return err
	}
	
	// Create bookings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bookings (
			id SERIAL PRIMARY KEY,
			customer_name VARCHAR(100) NOT NULL,
			customer_email VARCHAR(100) NOT NULL,
			customer_phone VARCHAR(20) NOT NULL,
			payment_method VARCHAR(20) NOT NULL,
			total_amount DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}
	
	// Create booking_seats join table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS booking_seats (
			booking_id INT REFERENCES bookings(id),
			seat_id INT REFERENCES seats(id),
			PRIMARY KEY (booking_id, seat_id)
		)
	`)
	
	return err
}

func initAdminUser(db *sql.DB) error {
	// Check if admin user exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return err
	}
	
	// If admin doesn't exist, create one
	if count == 0 {
		// In a real app, you'd hash the password properly
		_, err = db.Exec("INSERT INTO users (username, password_hash, is_admin) VALUES ('admin', 'admin123', TRUE)")
		if err != nil {
			return err
		}
	}
	
	return nil
}

func initSeats(db *sql.DB) error {
	// Check if seats exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM seats").Scan(&count)
	if err != nil {
		return err
	}
	
	// If no seats exist, create them
	if count == 0 {
		// Initialize all seats based on the theater layout
		
		// Regular seats
		// A1-13 & A18-30
		for i := 1; i <= 30; i++ {
			if i <= 13 || i >= 18 {
				_, err = db.Exec("INSERT INTO seats (row, number, type, status) VALUES ('A', $1, 'regular', 'available')", i)
				if err != nil {
					return err
				}
			}
		}
		
		// (B,C,D,E,F,G,H,I,J,K) rows with appropriate gaps
		rows := []string{"B", "C", "D", "E", "F", "G", "H", "I", "J", "K"}
		for _, row := range rows {
			for i := 1; i <= 30; i++ {
				// Skip seats B13-B18, C13-C18, and D13-D18
				if (row == "B" || row == "C" || row == "D") && i >= 13 && i <= 18 {
					continue
				}
				
				seatType := "regular"
				status := "available"
				
				// VIP seats: (E,F,G,H,I,J,K)13-15 & (E,F,G,H,I,J,K)16-18
				if (row >= "E" && row <= "K") && ((i >= 13 && i <= 15) || (i >= 16 && i <= 18)) {
					seatType = "vip"
				}
				
				_, err = db.Exec("INSERT INTO seats (row, number, type, status) VALUES ($1, $2, $3, $4)", 
					row, i, seatType, status)
				if err != nil {
					return err
				}
			}
		}
		
		// L1-7 & L24-30 (regular)
		// L8-23 (sponsored)
		for i := 1; i <= 30; i++ {
			seatType := "regular"
			status := "available"
			
			if i >= 8 && i <= 23 {
				seatType = "sponsored"
				status = "sponsored"
			}
			
			_, err = db.Exec("INSERT INTO seats (row, number, type, status) VALUES ('L', $1, $2, $3)", 
				i, seatType, status)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// GetAllSeats retrieves all seats from the database
func (db *Database) GetAllSeats() ([]models.Seat, error) {
	rows, err := db.Query("SELECT id, row, number, type, status FROM seats ORDER BY row, number")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	seats := []models.Seat{}
	for rows.Next() {
		var seat models.Seat
		if err := rows.Scan(&seat.ID, &seat.Row, &seat.Number, &seat.Type, &seat.Status); err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}
	
	return seats, nil
}

// BookSeats books seats for a customer
func (db *Database) BookSeats(booking models.Booking, seatIDs []int) (int, error) {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	
	// Insert booking
	var bookingID int
	err = tx.QueryRow(`
		INSERT INTO bookings 
		(customer_name, customer_email, customer_phone, payment_method, total_amount, status) 
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id
	`, booking.CustomerName, booking.CustomerEmail, booking.CustomerPhone, 
	   booking.PaymentMethod, booking.TotalAmount).Scan(&bookingID)
	if err != nil {
		return 0, err
	}
	
	// Update seat status to 'booked'
	for _, seatID := range seatIDs {
		_, err = tx.Exec("UPDATE seats SET status = 'booked' WHERE id = $1", seatID)
		if err != nil {
			return 0, err
		}
		
		// Link seat to booking
		_, err = tx.Exec("INSERT INTO booking_seats (booking_id, seat_id) VALUES ($1, $2)", 
			bookingID, seatID)
		if err != nil {
			return 0, err
		}
	}
	
	// Commit transaction
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	
	return bookingID, nil
}

// VerifyAdmin checks if username and password match an admin
func (db *Database) VerifyAdmin(username, password string) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE username = $1 AND password_hash = $2 AND is_admin = TRUE
	`, username, password).Scan(&count)
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// GetPendingBookings retrieves all pending bookings
func (db *Database) GetPendingBookings() ([]models.BookingResponse, error) {
	rows, err := db.Query(`
		SELECT b.id, b.customer_name, b.customer_email, b.customer_phone, 
		       b.payment_method, b.total_amount, b.status, b.created_at
		FROM bookings b
		WHERE status = 'pending'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	bookings := []models.BookingResponse{}
	for rows.Next() {
		var booking models.BookingResponse
		if err := rows.Scan(
			&booking.ID, &booking.CustomerName, &booking.CustomerEmail, &booking.CustomerPhone,
			&booking.PaymentMethod, &booking.TotalAmount, &booking.Status, &booking.CreatedAt,
		); err != nil {
			return nil, err
		}
		
		// Get seats for this booking
		seatRows, err := db.Query(`
			SELECT s.row, s.number, s.type
			FROM seats s
			JOIN booking_seats bs ON s.id = bs.seat_id
			WHERE bs.booking_id = $1
		`, booking.ID)
		if err != nil {
			return nil, err
		}
		
		for seatRows.Next() {
			var seatRow string
			var seatNumber int
			var seatType string
			if err := seatRows.Scan(&seatRow, &seatNumber, &seatType); err != nil {
				seatRows.Close()
				return nil, err
			}
			booking.Seats = append(booking.Seats, fmt.Sprintf("%s%d (%s)", seatRow, seatNumber, seatType))
		}
		seatRows.Close()
		
		bookings = append(bookings, booking)
	}
	
	return bookings, nil
}

// UpdateBookingStatus updates a booking's status and related seats
func (db *Database) UpdateBookingStatus(bookingID int, status string) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Update booking status
	_, err = tx.Exec("UPDATE bookings SET status = $1 WHERE id = $2", status, bookingID)
	if err != nil {
		return err
	}
	
	// Update seat statuses based on booking status
	seatStatus := "booked"
	if status == "confirmed" {
		seatStatus = "not-available"
	} else if status == "cancelled" {
		seatStatus = "available"
	}
	
	_, err = tx.Exec(`
		UPDATE seats 
		SET status = $1 
		WHERE id IN (
			SELECT seat_id FROM booking_seats WHERE booking_id = $2
		)
	`, seatStatus, bookingID)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}