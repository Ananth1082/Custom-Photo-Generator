package server

import (
	models "CustomPhotoGenerator/m-v0/Models"
	"image"
	"log"

	"github.com/gin-gonic/gin"
)

var Sr models.ShareRequest
var BaseImg image.Image
var Err error

func Server() {
	router := gin.Default()
	Routes(router)
	log.Println("Server started at :8080")
	log.Fatal(router.Run())
}
