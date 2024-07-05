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
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Add allowed origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true, // Set to true if you need to include credentials
		MaxAge:           12 * time.Hour,
	}))
	Routes(router)
	log.Println("Server started at :8080")
	log.Fatal(router.Run())
}
