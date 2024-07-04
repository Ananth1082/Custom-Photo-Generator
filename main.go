package main

import (
	models "CustomPhotoGenerator/m-v0/Models"
	telegramclient "CustomPhotoGenerator/m-v0/TelegramClient"
	"fmt"
	"image"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	env.SetENV()
}

func main() {

	r := gin.Default()
	var sr models.ShareRequest
	var baseImg image.Image
	var err error
	r.POST("/uploadImage", func(ctx *gin.Context) {
		file, handler, err := ctx.Request.FormFile("image")
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Sprintf("Failed to get file: %v", err))
			return
		}
		defer file.Close()
		baseImg, _, err = image.Decode(file)
		if err != nil {
			log.Fatal("Image not received")
			ctx.String(http.StatusInternalServerError, "Error uploading image")
		}
		ctx.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully.", handler.Filename))
	})

	// Process the image with text boxes
	r.POST("/sendTextBoxes", func(ctx *gin.Context) {
		var textData models.IshareRequest
		if err := ctx.ShouldBindJSON(&textData); err != nil {
			ctx.String(http.StatusBadRequest, fmt.Sprintf("Invalid request data: %v", err))
			return
		}
		sr.Varimg, err = printVarContent(baseImg, textData)
		if err != nil {
			log.Printf("Failed to add variable content: %v", err)
		}
		ctx.String(http.StatusOK, "Processing started successfully.")
	})

	// Endpoint to initiate Telegram authentication
	r.POST("/initAuth", func(c *gin.Context) {

		if err := c.BindJSON(&sr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		fmt.Println(sr.PhoneNumber, sr.Contacts)
		telegramclient.OtpState.WaitGroup.Add(1) // Increment the wait group counter before starting the goroutine

		go func() {
			log.Println("Starting authentication and message sending process.")
			err := telegramclient.AuthenticateAndSend(sr)
			if err != nil {
				log.Println("Error during authentication and sending message:", err)
			}
		}()

		// Respond immediately, informing the client to send the OTP
		c.JSON(http.StatusOK, gin.H{"status": "Authentication initiated, please submit the OTP using /share endpoint."})
	})

	// Endpoint to submit OTP
	r.POST("/share", func(c *gin.Context) {
		var otpRequest struct {
			OTP string `json:"otp"`
		}

		if err := c.ShouldBindJSON(&otpRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		telegramclient.OtpState.Lock()
		telegramclient.OtpState.Otp = otpRequest.OTP
		telegramclient.OtpState.Unlock()

		log.Println("OTP received from client:", otpRequest.OTP)

		telegramclient.OtpState.WaitGroup.Done() // Release the wait on OTP reception
		c.JSON(http.StatusOK, gin.H{"status": "OTP received"})
	})

	log.Println("Server started at :8080")
	log.Fatal(r.Run())
}
