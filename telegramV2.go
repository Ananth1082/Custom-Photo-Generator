package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func init() {
	setEnv()
}

// Shared state to hold OTP and manage synchronization
type OTPState struct {
	sync.Mutex
	otp       string
	waitGroup sync.WaitGroup
}

type ShareRequest struct {
	Img         []string
	Contacts    []string
	PhoneNumber string
}

var otpState = &OTPState{}

func main() {

	r := gin.Default()

	r.POST("/uploadImage", func(ctx *gin.Context) {
		file, handler, err := ctx.Request.FormFile("image")
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Sprintf("Failed to get file: %v", err))
			return
		}
		defer file.Close()

		if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create uploads directory: %v", err))
			return
		}

		dst, err := os.Create(filepath.Join("./uploads", handler.Filename))
		if err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create file: %v", err))
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save file: %v", err))
			return
		}
		ctx.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully.", handler.Filename))
	})

	// Process the image with text boxes
	r.POST("/sendTextBoxes", func(ctx *gin.Context) {
		var textData ishareRequest
		if err := ctx.ShouldBindJSON(&textData); err != nil {
			ctx.String(http.StatusBadRequest, fmt.Sprintf("Invalid request data: %v", err))
			return
		}

		img, err := loadImage("./uploads/" + textData.MergedImage)
		if err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to load image: %v", err))
			return
		}

		// Handle constant content
		img, err = printConstContent(img, textData)
		if err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to add constant content: %v", err))
			return
		}
		// Handle variable content
		go func() {
			if err := printVarContent(img, textData); err != nil {
				log.Printf("Failed to add variable content: %v", err)
			}
		}()

		ctx.String(http.StatusOK, "Processing started successfully.")
	})

	// Endpoint to initiate Telegram authentication
	r.POST("/initAuth", func(c *gin.Context) {
		var sr ShareRequest
		if err := c.BindJSON(&sr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		fmt.Println(sr.PhoneNumber, sr.Contacts)
		otpState.waitGroup.Add(1) // Increment the wait group counter before starting the goroutine

		go func() {
			log.Println("Starting authentication and message sending process.")
			err := authenticateAndSend(sr)
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

		otpState.Lock()
		otpState.otp = otpRequest.OTP
		otpState.Unlock()

		log.Println("OTP received from client:", otpRequest.OTP)

		otpState.waitGroup.Done() // Release the wait on OTP reception
		c.JSON(http.StatusOK, gin.H{"status": "OTP received"})
	})

	log.Println("Server started at :8080")
	log.Fatal(r.Run())
}

func authenticateAndSend(sr ShareRequest) error {
	ctx := context.Background()
	var (
		apiID, _ = strconv.Atoi(os.Getenv("API_ID"))
		apiHash  = os.Getenv("API_HASH")
	)
	client := telegram.NewClient(apiID, apiHash, telegram.Options{})

	err := client.Run(ctx, func(ctx context.Context) error {
		// Define the code prompt function
		codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
			// Here we simulate waiting for OTP from client
			log.Println("Waiting for OTP from client...")
			otpState.waitGroup.Wait() // Block until OTP is received

			otpState.Lock()
			defer otpState.Unlock()
			return strings.TrimSpace(otpState.otp), nil
		}

		// Create and run the authentication flow
		flow := auth.NewFlow(
			auth.CodeOnly(sr.PhoneNumber, auth.CodeAuthenticatorFunc(codePrompt)),
			auth.SendCodeOptions{AllowFlashCall: true, CurrentNumber: true},
		)

		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Create a tg.Client to interact with the Telegram API
		api := tg.NewClient(client)

		// Import contacts
		importResult, err := importContacts(ctx, api, sr.Contacts)
		if err != nil {
			return err
		}
		for i := 0; i < len(importResult.Users); i++ {
			if err := uploadAndSendPhoto(ctx, api, sr.Img[i], importResult.Users[i]); err != nil {
				return err
			}
		}
		// Upload and send the photo

		fmt.Println("Photo sent successfully!")
		return nil
	})

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Import contacts and return the result
func importContacts(ctx context.Context, api *tg.Client, contacts []string) (*tg.ContactsImportedContacts, error) {
	tgContacts := make([]tg.InputPhoneContact, len(contacts))
	for i, contact := range contacts {
		tgContacts[i] = tg.InputPhoneContact{
			ClientID:  rand.Int63(),
			Phone:     contact,
			FirstName: "John", // Optional, can be replaced with actual name
			LastName:  "Doe",  // Optional, can be replaced with actual surname
		}
	}

	importResult, err := api.ContactsImportContacts(ctx, tgContacts)
	if err != nil {
		return nil, fmt.Errorf("failed to import contacts: %w", err)
	}

	if len(importResult.Users) == 0 {
		return nil, fmt.Errorf("no users found with provided phone numbers")
	}

	return importResult, nil
}

// Upload a photo to Telegram and send it to specified users
func uploadAndSendPhoto(ctx context.Context, api *tg.Client, imgPath string, recp tg.UserClass) error {
	// Read the photo
	photo, err := os.ReadFile(filepath.Join("./OUT", imgPath))
	if err != nil {
		return fmt.Errorf("failed to read image file: %w", err)
	}

	// Upload the photo in parts
	fileID := rand.Int63()
	partSize := 524288 // 512 KB per part
	numParts := (len(photo) + partSize - 1) / partSize

	for part := 0; part < numParts; part++ {
		start := part * partSize
		end := start + partSize
		if end > len(photo) {
			end = len(photo)
		}

		_, err := api.UploadSaveFilePart(ctx, &tg.UploadSaveFilePartRequest{
			FileID:   fileID,
			FilePart: part,
			Bytes:    photo[start:end],
		})
		if err != nil {
			return fmt.Errorf("failed to upload file part: %w", err)
		}
	}

	// Prepare the uploaded file as an input file
	inputFile := &tg.InputFile{
		ID:    fileID,
		Parts: numParts,
		Name:  filepath.Base(imgPath),
	}

	// Send the photo to each user

	user := recp.(*tg.User)
	_, err = api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer: &tg.InputPeerUser{
			UserID:     user.ID,
			AccessHash: user.AccessHash,
		},
		Media: &tg.InputMediaUploadedPhoto{
			File: inputFile,
		},
		RandomID: rand.Int63(),
	})
	if err != nil {
		return fmt.Errorf("failed to send photo to user %d: %w", user.ID, err)
	}

	return nil
}
