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
		AllowOrigins: []string{"*", "http://localhost:5173", "http://172.16.25.167:5173", "https://test2-liart-seven.vercel.app/"}, // Add allowed origins
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))
	router.OPTIONS("/*cors", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.JSON(204, nil)
	})

	Routes(router)
	log.Println("Server started at :8080")
	log.Fatal(router.Run())
}
