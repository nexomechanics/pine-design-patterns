package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func handleGenerate(c *gin.Context) {
	apiKey := os.Getenv("API_KEY")
	if apiKey != "" && c.GetHeader("X-API-Key") != apiKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.resolve()
	result := generate(req)
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "X-API-Key"},
	}))

	r.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.POST("/generate", handleGenerate)
	r.Run(":8080")
}
