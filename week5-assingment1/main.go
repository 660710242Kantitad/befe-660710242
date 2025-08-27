package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Book struct
type Book struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Author   string  `json:"author"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

// fake data
var books = []Book{
	{ID: "1", Title: "Go Programming", Author: "John Doe", Category: "Programming", Price: 500},
	{ID: "2", Title: "Python for Data Science", Author: "Jane Smith", Category: "Data Science", Price: 650},
	{ID: "3", Title: "Design Patterns", Author: "Gamma", Category: "Programming", Price: 700},
}

// ดึงหนังสือทั้งหมด หรือ filter ตาม category
func getBooks(c *gin.Context) {
	category := c.Query("category")

	if category != "" {
		filter := []Book{}
		for _, book := range books {
			if book.Category == category {
				filter = append(filter, book)
			}
		}
		c.JSON(http.StatusOK, filter)
		return
	}

	c.JSON(http.StatusOK, books)
}

// ดึงหนังสือตาม ID
func getBookByID(c *gin.Context) {
	id := c.Param("id")

	for _, book := range books {
		if book.ID == id {
			c.JSON(http.StatusOK, book)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
}

func main() {
	r := gin.Default()

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "healthy"})
	})

	// group API
	api := r.Group("/api/v1")
	{
		api.GET("/books", getBooks)        // เส้นที่ 1
		api.GET("/books/:id", getBookByID) // เส้นที่ 2
	}

	r.Run(":8080")
}
