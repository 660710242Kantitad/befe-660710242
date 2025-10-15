package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "week11-assignment/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Book struct (ฟิลด์หลัก)
type Book struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	ISBN   string  `json:"isbn"`
	Year   int     `json:"year"`
	Price  float64 `json:"price"`

	// ฟิลด์ใหม่
	Category      string   `json:"category"`
	OriginalPrice *float64 `json:"original_price,omitempty"`
	Discount      int      `json:"discount"`
	CoverImage    string   `json:"cover_image"`
	Rating        float64  `json:"rating"`
	ReviewsCount  int      `json:"reviews_count"`
	IsNew         bool     `json:"is_new"`
	Pages         *int     `json:"pages,omitempty"`
	Language      string   `json:"language"`
	Publisher     string   `json:"publisher"`
	Description   string   `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var db *sql.DB

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDB() {
	var err error
	host := getEnv("DB_HOST", "localhost")
	name := getEnv("DB_NAME", "mydb")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	port := getEnv("DB_PORT", "5432")

	conStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	db, err = sql.Open("postgres", conStr)

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	log.Println("✅ Successfully connected to database")
}

// ----------- HANDLERS -----------

// GET /api/v1/books?category=fiction
func getBooks(c *gin.Context) {
	category := c.Query("category")

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = db.Query(`
			SELECT id, title, author, isbn, year, price, created_at, updated_at 
			FROM books WHERE LOWER(category) = LOWER($1)
		`, category)
	} else {
		rows, err = db.Query(`
			SELECT id, title, author, isbn, year, price, created_at, updated_at
			FROM books
		`)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			log.Println("scan error:", err)
			continue
		}
		books = append(books, book)
	}
	c.JSON(http.StatusOK, books)
}

// GET /api/v1/categories
func getCategories(c *gin.Context) {
	rows, err := db.Query("SELECT DISTINCT category FROM books ORDER BY category")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		err := rows.Scan(&cat)
		if err != nil {
			log.Println("scan category error:", err)
			continue
		}
		categories = append(categories, cat)
	}

	c.JSON(http.StatusOK, categories)
}

// GET /api/v1/books/search?q=keyword
func searchBooks(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' required"})
		return
	}
	keyword = "%" + strings.ToLower(keyword) + "%"

	rows, err := db.Query(`
		SELECT id, title, author, isbn, year, price, created_at, updated_at
		FROM books
		WHERE LOWER(title) LIKE $1 OR LOWER(author) LIKE $1
	`, keyword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			log.Println("scan error:", err)
			continue
		}
		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

// GET /api/v1/books/featured (สมมติ featured คือ discount > 0 and rating >= 4)
func getFeaturedBooks(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, title, author, isbn, year, price, created_at, updated_at
		FROM books
		WHERE discount > 0 AND rating >= 4.0
		ORDER BY rating DESC, discount DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			log.Println("scan error:", err)
			continue
		}
		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

// GET /api/v1/books/new
func getNewBooks(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, title, author, isbn, year, price, created_at, updated_at
		FROM books
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			log.Println("scan error:", err)
			continue
		}
		books = append(books, book)
	}
	c.JSON(http.StatusOK, books)
}

// GET /api/v1/books/discounted (discount > 0)
func getDiscountedBooks(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, title, author, isbn, year, price, created_at, updated_at
		FROM books
		WHERE discount > 0
		ORDER BY discount DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			log.Println("scan error:", err)
			continue
		}
		books = append(books, book)
	}
	c.JSON(http.StatusOK, books)
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()
	r.Use(cors.Default())

	// Swagger UI
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/books", getBooks)
		api.GET("/categories", getCategories)
		api.GET("/books/search", searchBooks)
		api.GET("/books/featured", getFeaturedBooks)
		api.GET("/books/new", getNewBooks)
		api.GET("/books/discounted", getDiscountedBooks)
	}

	r.Run(":8080")
}
