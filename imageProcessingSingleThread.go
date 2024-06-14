// package main

// import (
// 	"fmt"
// 	"image"
// 	"image/color"
// 	"image/draw"
// 	"image/png"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang/freetype"
// 	"github.com/golang/freetype/truetype"
// 	_ "golang.org/x/image/font"
// )

// type ShareRequest struct {
// 	MergedImage    string
// 	ConstTextBoxes []ConstTextBox
// 	VarTextBoxes   []VarTextBox
// }

// type TextDetails struct {
// 	Name     string
// 	IsItalic bool
// 	IsBold   bool
// 	Color    color.RGBA
// 	Size     int
// }

// type ConstTextBox struct {
// 	MetaDetails  TextDetails
// 	ConstContent string
// 	Location     image.Point
// }
// type VarTextBox struct {
// 	MetaDetails TextDetails
// 	VarContent  []string
// 	Location    image.Point
// }

// func loadImage(filename string) (image.Image, error) {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	img, err := png.Decode(file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return img, nil
// }

// func saveImage(img image.Image, filename string) error {
// 	outFile, err := os.Create(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer outFile.Close()

// 	return png.Encode(outFile, img)
// }

// func loadFont(fontDetails TextDetails) *truetype.Font {
// 	fontSpecifier := fontDetails.Name + "-"
// 	if fontDetails.IsBold {
// 		fontSpecifier += "Bold"
// 	}
// 	if fontDetails.IsItalic {
// 		fontSpecifier += "Italic"
// 	}
// 	if !fontDetails.IsBold && !fontDetails.IsItalic {
// 		fontSpecifier += "Regular"
// 	}

// 	fontBytes, err := os.ReadFile(fmt.Sprintf("./Fonts/%s/%s.ttf", fontDetails.Name, fontSpecifier))
// 	if err != nil {
// 		log.Fatalf("failed to load font: %v", err)
// 	}
// 	font, err := freetype.ParseFont(fontBytes)
// 	if err != nil {
// 		log.Fatalf("failed to parse font: %v", err)
// 	}
// 	return font
// }

// func addLabel(img image.Image, td TextDetails, location image.Point, content string) (image.Image, error) {
// 	canvas := image.NewRGBA(img.Bounds())
// 	draw.Draw(canvas, canvas.Bounds(), img, image.Point{}, draw.Src)
// 	font := loadFont(td)

// 	c := freetype.NewContext()
// 	c.SetDPI(72)
// 	c.SetFont(font)
// 	c.SetFontSize(float64(td.Size))
// 	c.SetClip(canvas.Bounds())
// 	c.SetDst(canvas)
// 	c.SetSrc(&image.Uniform{td.Color})

// 	pt := freetype.Pt(location.X, location.Y+int(c.PointToFixed(24)>>6))
// 	_, err := c.DrawString(content, pt)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return canvas, nil
// }

// func PrintVarContent(img image.Image, textData ShareRequest) {
// 	if len(textData.VarTextBoxes) == 0 {
// 		saveImage(img, "./OUT/output.png")
// 		return
// 	}
// 	var err error
// 	var labeledImg image.Image

// 	for i := 0; i < len(textData.VarTextBoxes[0].VarContent); i++ {
// 		labeledImg = img
// 		for _, tb := range textData.VarTextBoxes {
// 			labeledImg, err = addLabel(labeledImg, tb.MetaDetails, tb.Location, tb.VarContent[i])
// 			if err != nil {
// 				log.Fatal("An error occured while printing")
// 			}
// 			go saveImage(labeledImg, fmt.Sprintf("./OUT/output_%d.png", i))
// 		}
// 	}

// 	log.Println("Image saved successfully.")
// }

// func PrintConstContent(img image.Image, textData ShareRequest) image.Image {
// 	if len(textData.ConstTextBoxes) == 0 {
// 		return img
// 	}
// 	var err error
// 	var labeledImg image.Image = img
// 	for _, tb := range textData.ConstTextBoxes {

// 		labeledImg, err = addLabel(labeledImg, tb.MetaDetails, tb.Location, tb.ConstContent)
// 		if err != nil {
// 			log.Fatalf("failed to add label: %v", err)
// 		}

// 	}
// 	err = saveImage(labeledImg, "./OUT/const_output.png")
// 	if err != nil {
// 		log.Fatalf("failed to save image: %v", err)
// 	}

// 	log.Println("Image saved successfully.")
// 	return labeledImg
// }

// func main() {

// 	//example request body
// 	// textData := ShareRequest{
// 	// 	ConstTextBoxes: []ConstTextBox{
// 	// 		{MetaDetails: TextDetails{Name: "Poppins", IsItalic: false, IsBold: true, Color: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Size: 70},
// 	// 			ConstContent: "Jai Hind",
// 	// 			Location:     image.Pt(400, 950),
// 	// 		},
// 	// 		{MetaDetails: TextDetails{Name: "Roboto", IsItalic: false, IsBold: false, Color: color.RGBA{R: 0, G: 255, B: 0, A: 255}, Size: 50},
// 	// 			ConstContent: "Indian Army",
// 	// 			Location:     image.Pt(50, 100),
// 	// 		},
// 	// 	},
// 	// 	VarTextBoxes: []VarTextBox{{
// 	// 		VarContent:  []string{"Ananth contact:+91 9354821328", "Len contact:+91 6987451554", "Varun contact:+91 8542555421", "Kavan contact:+91 7841254574", "Chaddi gopal contact:+9169696969"},
// 	// 		MetaDetails: TextDetails{Name: "Roboto", IsItalic: false, IsBold: false, Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}, Size: 30},
// 	// 		Location:    image.Pt(10, 1050),
// 	// 	},
// 	// 	},
// 	// }
// 	r := gin.Default()
// 	//option 1 upload file in first request and save it in local device,
// 	// then in the second request send the metadata
// 	r.POST("/uploadImage", func(ctx *gin.Context) {

// 		// Parse the multipart form, limit the maximum memory to 10 MB
// 		if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
// 			ctx.String(http.StatusBadRequest, "File is too large")
// 			return
// 		}

// 		// Get the file from the form input name 'file'
// 		file, handler, err := ctx.Request.FormFile("image")
// 		if err != nil {
// 			ctx.String(http.StatusBadRequest, fmt.Sprintf("Failed to get file: %v", err))
// 			return
// 		}
// 		defer file.Close()

// 		// Create the uploads folder if it doesn't exist
// 		if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
// 			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create uploads directory: %v", err))
// 			return
// 		}

// 		// Create a new file in the uploads directory
// 		dst, err := os.Create(filepath.Join("./uploads", handler.Filename))
// 		if err != nil {
// 			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create file: %v", err))
// 			return
// 		}

// 		// Copy the uploaded file to the destination file
// 		if _, err := io.Copy(dst, file); err != nil {
// 			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save file: %v", err))
// 			return
// 		}
// 		ctx.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully.", handler.Filename))

// 	})

// 	r.POST("/sendTextBoxes", func(ctx *gin.Context) {
// 		var textData ShareRequest
// 		err := ctx.BindJSON(&textData)
// 		if err != nil {
// 			log.Panic("Error", err)
// 			ctx.String(400, "Error Occured")
// 		}
// 		fmt.Println(textData.MergedImage)
// 		img, err := loadImage("./uploads/" + textData.MergedImage)
// 		if err != nil {
// 			fmt.Println("Error occured while loading Image")
// 		}

// 		PrintVarContent(img, textData)
// 		ctx.String(200, "Successfully created")
// 	})

// 	//option 2 send the json meta data as a field in the form

// 	r.Run()
// }

package main
