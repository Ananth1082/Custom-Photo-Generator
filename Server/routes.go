package server

import (
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine) {
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(200, "Pong")
	})
	//Endpoint to upload images to be shared
	r.POST("/uploadImage", uploadImageController)

	// Process the image with text boxes
	r.POST("/sendTextBoxes", sendTextboxesController)

	// Endpoint to initiate Telegram authentication
	r.POST("/initAuth", initiateTelegramAuthController)

	// Endpoint to submit OTP
	r.POST("/share", shareToContactsController)
}
