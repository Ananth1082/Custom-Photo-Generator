package server

import (
	models "CustomPhotoGenerator/m-v0/Models"
	"image"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var Sr models.ShareRequest
var BaseImg image.Image
var Err error

func Server() {
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})
	Routes(router)
	log.Println("Server started at :8080")
	log.Fatal(router.Run())
}
