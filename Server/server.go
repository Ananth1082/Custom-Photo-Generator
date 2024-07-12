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
	config := cors.Config{
		AllowAllOrigins: true,

		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(config))
	// Define your routes
	Routes(router)

	// Start the server
	log.Println("Server started at :8080")
	log.Fatal(router.Run())
}
