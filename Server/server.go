package server

import (
	models "CustomPhotoGenerator/m-v0/Models"
	"image"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var Sr models.ShareRequest
var BaseImg image.Image
var Err error

func Server() {
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://172.16.25.167:5173", "https://test2-liart-seven.vercel.app", "https://vite-project-psi-ivory.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true, // Allow credentials
		MaxAge:           12 * time.Hour,
	}))

	// Explicitly handle OPTIONS requests for CORS
	router.OPTIONS("/*cors", func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "https://test2-liart-seven.vercel.app" || origin == "http://localhost:5173" || origin == "http://172.16.25.167:5173" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Status(204)
	})

	// Define your routes
	Routes(router)

	// Start the server
	log.Println("Server started at :8080")
	log.Fatal(router.Run())
}
