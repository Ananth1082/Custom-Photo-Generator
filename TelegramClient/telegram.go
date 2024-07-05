package telegramclient

import (
	models "CustomPhotoGenerator/m-v0/Models"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

var OtpState = &models.OTPState{}

func AuthenticateAndSend(sr models.ShareRequest) error {
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
			OtpState.WaitGroup.Wait() // Block until OTP is received

			OtpState.Lock()
			defer OtpState.Unlock()
			return strings.TrimSpace(OtpState.Otp), nil
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
		importResult, err := ImportContacts(ctx, api, sr.Contacts)
		if err != nil {
			return err
		}
		for i := 0; i < len(importResult.Users); i++ {
			n := len(sr.Varimg)
			if i >= n {
				err := UploadAndSendPhoto(ctx, api, sr.Varimg[n-1], importResult.Users[i])
				if err != nil {
					return err
				}
			} else if err := UploadAndSendPhoto(ctx, api, sr.Varimg[i], importResult.Users[i]); err != nil {
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
func ImportContacts(ctx context.Context, api *tg.Client, contacts []string) (*tg.ContactsImportedContacts, error) {
	tgContacts := make([]tg.InputPhoneContact, len(contacts))
	for i, contact := range contacts {
		tgContacts[i] = tg.InputPhoneContact{
			ClientID: rand.Int63(),
			Phone:    contact,
			// FirstName: "John", // Optional, can be replaced with actual name
			// LastName:  "Doe",  // Optional, can be replaced with actual surname
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
func UploadAndSendPhoto(ctx context.Context, api *tg.Client, imgBytes []byte, recp tg.UserClass) error {

	// Upload the photo in parts
	fileID := rand.Int63()
	partSize := 524288 // 512 KB per part
	numParts := (len(imgBytes) + partSize - 1) / partSize

	for part := 0; part < numParts; part++ {
		start := part * partSize
		end := start + partSize
		if end > len(imgBytes) {
			end = len(imgBytes)
		}

		_, err := api.UploadSaveFilePart(ctx, &tg.UploadSaveFilePartRequest{
			FileID:   fileID,
			FilePart: part,
			Bytes:    imgBytes[start:end],
		})
		if err != nil {
			return fmt.Errorf("failed to upload file part: %w", err)
		}
	}

	// Prepare the uploaded file as an input file
	inputFile := &tg.InputFile{
		ID:    fileID,
		Parts: numParts,
		Name:  "send.png",
	}

	// Send the photo to each user

	user := recp.(*tg.User)
	_, err := api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
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
