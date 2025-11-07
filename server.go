package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// URLMapping represents a shortened URL and its original URL
type URLMapping struct {
	ID          int64  `json:"id"`           // Database ID (auto-increment)
	ShortCode   string `json:"short_code"`   // The shortened code (e.g., "abc123")
	OriginalURL string `json:"original_url"` // The full original URL
}

// ShortenRequest represents the JSON payload for creating a short URL
type ShortenRequest struct {
	URL string `json:"url" validate:"required"` // The URL to be shortened
}

// ShortenResponse represents the JSON response after creating a short URL
type ShortenResponse struct {
	ShortCode string `json:"short_code"` // The generated short code
	ShortURL  string `json:"short_url"`  // The complete shortened URL
}

// ErrorResponse represents an error message response
type ErrorResponse struct {
	Message string `json:"message"`
}

// Database holds the database connection
type Database struct {
	conn *sql.DB
}

// NewDatabase creates a new database connection
// It expects a PostgreSQL connection string like:
// "postgres://username:password@localhost:5432/dbname?sslmode=disable"
func NewDatabase(connectionString string) (*Database, error) {
	// Open database connection
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ Database connected successfully")

	return &Database{conn: db}, nil
}

// InitSchema creates the necessary database tables if they don't exist
func (db *Database) InitSchema() error {
	// Create the urls table with an auto-incrementing ID
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,          -- Auto-incrementing ID
			short_code VARCHAR(20) UNIQUE NOT NULL,  -- The Base62 code
			original_url TEXT NOT NULL,     -- The original long URL
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- When it was created
		);

		-- Create an index on short_code for faster lookups
		CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return err
	}

	log.Println("✅ Database schema initialized")
	return nil
}

// SaveURL inserts a new URL mapping into the database
// Returns the auto-generated ID from the database
func (db *Database) SaveURL(shortCode, originalURL string) (int64, error) {
	query := `
		INSERT INTO urls (short_code, original_url) 
		VALUES ($1, $2) 
		RETURNING id
	`

	var id int64
	err := db.conn.QueryRow(query, shortCode, originalURL).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetURL retrieves the original URL by short code
// Returns the URL mapping and a boolean indicating if it was found
func (db *Database) GetURL(shortCode string) (*URLMapping, bool, error) {
	query := `
		SELECT id, short_code, original_url 
		FROM urls 
		WHERE short_code = $1
	`

	var mapping URLMapping
	err := db.conn.QueryRow(query, shortCode).Scan(
		&mapping.ID,
		&mapping.ShortCode,
		&mapping.OriginalURL,
	)

	// If no rows found, return false for "exists"
	if err == sql.ErrNoRows {
		return nil, false, nil
	}

	// If other error occurred, return the error
	if err != nil {
		return nil, false, err
	}

	// Successfully found the URL
	return &mapping, true, nil
}

// GetNextID returns the next available ID from the database sequence
// This is used to generate the short code
func (db *Database) GetNextID() (int64, error) {
	// Get the next value from the PostgreSQL sequence
	// SERIAL columns automatically create a sequence named tablename_columnname_seq
	query := `SELECT nextval('urls_id_seq')`

	var id int64
	err := db.conn.QueryRow(query).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}

// Base62 character set: 0-9, a-z, A-Z (62 characters total)
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateShortCode converts an integer ID to a Base62 string
// Base62 uses: 0-9 (10 chars) + a-z (26 chars) + A-Z (26 chars) = 62 total
// Examples:
//   - 0 -> "0"
//   - 61 -> "Z"
//   - 62 -> "10"
//   - 15432 -> "3dE"
func generateShortCode(id int64) string {
	// Handle the edge case of 0
	if id == 0 {
		return string(base62Chars[0])
	}

	// Build the result string in reverse order
	result := ""

	// Keep dividing by 62 and taking remainders
	for id > 0 {
		// Get the remainder when dividing by 62
		remainder := id % 62

		// Prepend the corresponding character to our result
		// (prepending because we're building the string backwards)
		result = string(base62Chars[remainder]) + result

		// Divide by 62 for the next iteration
		id = id / 62
	}

	return result
}

func main() {
	// Get database connection string from environment variable
	// Default to local PostgreSQL if not set
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable"
		log.Println("⚠️  DATABASE_URL not set, using default:", dbURL)
	}

	// Initialize database connection
	db, err := NewDatabase(dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize database schema (create tables)
	if err := db.InitSchema(); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}

	// Initialize Echo framework
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())  // Logs all HTTP requests
	e.Use(middleware.Recover()) // Recovers from panics

	// CORS middleware to allow cross-origin requests
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))

	// Routes
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// POST /shorten - Create a shortened URL
	e.POST("/shorten", func(c echo.Context) error {
		// Parse the request body
		req := new(ShortenRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "Invalid request body",
			})
		}

		// Validate that URL is provided
		if req.URL == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "URL is required",
			})
		}

		// Get the next sequential ID from the database
		id, err := db.GetNextID()
		if err != nil {
			log.Println("Error getting next ID:", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "Failed to generate short code",
			})
		}

		// Generate a short code by encoding the ID in Base62
		shortCode := generateShortCode(id)

		// Save the mapping to the database
		_, err = db.SaveURL(shortCode, req.URL)
		if err != nil {
			log.Println("Error saving URL:", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "Failed to save URL",
			})
		}

		// Build the full shortened URL
		// In production, you'd use your actual domain
		shortURL := "http://localhost:8080/" + shortCode

		// Return the response
		return c.JSON(http.StatusCreated, ShortenResponse{
			ShortCode: shortCode,
			ShortURL:  shortURL,
		})
	})

	// GET /:shortCode - Redirect to original URL
	e.GET("/:shortCode", func(c echo.Context) error {
		// Get the short code from URL parameter
		shortCode := c.Param("shortCode")

		// Look up the original URL from database
		mapping, exists, err := db.GetURL(shortCode)
		if err != nil {
			log.Println("Error retrieving URL:", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "Database error",
			})
		}

		// If not found, return 404
		if !exists {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "Short URL not found",
			})
		}

		// Redirect to the original URL with 301 (permanent redirect)
		return c.Redirect(http.StatusMovedPermanently, mapping.OriginalURL)
	})

	// GET /api/stats/:shortCode - Get URL information (bonus endpoint)
	e.GET("/api/stats/:shortCode", func(c echo.Context) error {
		shortCode := c.Param("shortCode")

		// Look up the original URL from database
		mapping, exists, err := db.GetURL(shortCode)
		if err != nil {
			log.Println("Error retrieving URL:", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "Database error",
			})
		}

		if !exists {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "Short URL not found",
			})
		}

		// Return the mapping information
		return c.JSON(http.StatusOK, mapping)
	})

	// Start the server on port 8080
	log.Println("🚀 Server starting on http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
