package utils

import (
	models "CustomPhotoGenerator/m-v0/Models"
	"fmt"
	"image"
	"log"

	"sync"

	_ "golang.org/x/image/font"
)

func PrintVarContent(img image.Image, textData models.IshareRequest) ([][]byte, error) {
	noOfDetails := len(textData.VarTextBoxes[0].VarContent)
	imagesBytes := make([][]byte, noOfDetails) // Initialize slice with fixed length

	if len(textData.VarTextBoxes) == 0 && len(textData.VarImageBoxes) == 0 {
		imgBytes, err := ImageToBytes(img)
		if err != nil {
			return nil, err
		}
		imagesBytes[0] = imgBytes // Place the single image at index 0
		return imagesBytes, nil
	}

	var mu sync.Mutex                       // Mutex to protect access to imagesBytes
	errors := make(chan error, noOfDetails) // Channel to collect errors
	numWorkers := 10
	var wg sync.WaitGroup

	// Worker function
	worker := func(jobs <-chan int) {
		for i := range jobs {
			labeledImg := img
			var err error
			for _, tb := range textData.VarTextBoxes {
				labeledImg, err = AddLabel(labeledImg, tb.MetaDetails, tb.Location, tb.VarContent[i])
				if err != nil {
					errors <- fmt.Errorf("error adding label: %w", err)
					return
				}
			}
			for _, itb := range textData.VarImageBoxes {
				log.Println("Gimme the image...")
				foregroundImage, err := GetImage(itb.ImageLink[i])
				log.Println("Got the image...")
				if err != nil {
					errors <- fmt.Errorf("error getting image: %w", err)
					return
				}
				foregroundImage = FitImage(foregroundImage, itb.MetaDetails.Size.Height, itb.MetaDetails.Size.Width)
				labeledImg = AddImage(labeledImg, foregroundImage, itb.Location)
			}
			imgBytes, err := ImageToBytes(labeledImg)
			if err != nil {
				errors <- fmt.Errorf("error saving image: %w", err)
				return
			}
			mu.Lock()
			imagesBytes[i] = imgBytes // Store image bytes at the correct index
			mu.Unlock()
		}
		wg.Done()
	}

	// Create a channel to send jobs to workers
	jobs := make(chan int, noOfDetails)

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(jobs)
	}

	// Send jobs to workers
	for i := 0; i < noOfDetails; i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
	close(errors)

	// Collect any errors
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	log.Println("All variable images processed successfully.")
	return imagesBytes, nil
}
