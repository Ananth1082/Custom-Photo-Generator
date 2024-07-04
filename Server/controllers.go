package server

import (
	models "CustomPhotoGenerator/m-v0/Models"
	telegramclient "CustomPhotoGenerator/m-v0/TelegramClient"
	"CustomPhotoGenerator/m-v0/utils"
	"fmt"
	"image"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func uploadImageController(ctx *gin.Context) {
	file, handler, err := ctx.Request.FormFile("image")
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Failed to get file: %v", err))
		return
	}
	defer file.Close()
	BaseImg, _, err = image.Decode(file)
	if err != nil {
		log.Fatal("Image not received")
		ctx.String(http.StatusInternalServerError, "Error uploading image")
	}
	ctx.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully.", handler.Filename))

}

func sendTextboxesController(ctx *gin.Context) {
	var textData models.IshareRequest
	if err := ctx.ShouldBindJSON(&textData); err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Invalid request data: %v", err))
		return
	}
	Sr.Varimg, Err = utils.PrintVarContent(BaseImg, textData)
	if Err != nil {
		log.Printf("Failed to add variable content: %v", Err)
	}
	ctx.String(http.StatusOK, "Processing started successfully.")
}

func initiateTelegramAuthController(c *gin.Context) {

	if err := c.BindJSON(&Sr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	fmt.Println(Sr.PhoneNumber, Sr.Contacts)
	telegramclient.OtpState.WaitGroup.Add(1) // Increment the wait group counter before starting the goroutine

	go func() {
		log.Println("Starting authentication and message sending process.")
		err := telegramclient.AuthenticateAndSend(Sr)
		if err != nil {
			log.Println("Error during authentication and sending message:", err)
		}
	}()

	// Respond immediately, informing the client to send the OTP
	c.JSON(http.StatusOK, gin.H{"status": "Authentication initiated, please submit the OTP using /share endpoint."})
}

func shareToContactsController(c *gin.Context) {
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
}
