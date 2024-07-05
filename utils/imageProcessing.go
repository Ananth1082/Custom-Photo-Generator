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
	var imagesBytes [][]byte // To store the bytes of all generated images
	if len(textData.VarTextBoxes) == 0 && len(textData.VarImageBoxes) == 0 {
		imgBytes, err := ImageToBytes(img)
		if err != nil {
			return nil, err
		}
		imagesBytes = append(imagesBytes, imgBytes)
		return imagesBytes, nil
	}

	var mu sync.Mutex                                                         // To safely append to imagesBytes
	errors := make(chan error, len(textData.VarTextBoxes[0].VarContent))      // Channel for errors
	imagesChan := make(chan []byte, len(textData.VarTextBoxes[0].VarContent)) // Channel for image bytes

	// Number of workers
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
				foregroundImage, err := GetImage(itb.ImageLink[i])
				foregroundImage = FitImage(foregroundImage, itb.MetaDetails.Size.Height, itb.MetaDetails.Size.Width)
				if err != nil {
					log.Fatal("error getting image", err)
				}
				labeledImg = AddImage(labeledImg, foregroundImage, itb.Location)
			}
			imgBytes, err := ImageToBytes(labeledImg)
			if err != nil {
				errors <- fmt.Errorf("error saving image: %w", err)
				return
			}
			mu.Lock()
			imagesBytes = append(imagesBytes, imgBytes)
			mu.Unlock()
		}
		wg.Done()
	}

	// Create a channel to send jobs to workers
	jobs := make(chan int, len(textData.VarTextBoxes[0].VarContent))

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(jobs)
	}

	// Send jobs to workers
	for i := 0; i < len(textData.VarTextBoxes[0].VarContent); i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
	close(imagesChan)
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
