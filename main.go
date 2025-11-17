package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type colorRGBA struct {
	R uint32
	G uint32
	B uint32
	A uint32
}

var colorPalette = []color.Color{
	color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}, // Black       - 0
	color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, // White       - 1
	color.RGBA{R: 0x9f, G: 0x4e, B: 0x44, A: 0xff}, // Red         - 2
	color.RGBA{R: 0x6a, G: 0xbf, B: 0xc6, A: 0xff}, // Cyan        - 3
	color.RGBA{R: 0xa0, G: 0x57, B: 0xa3, A: 0xff}, // Purple      - 4
	color.RGBA{R: 0x5c, G: 0xab, B: 0x5e, A: 0xff}, // Green       - 5
	color.RGBA{R: 0x50, G: 0x45, B: 0x9b, A: 0xff}, // Blue        - 6
	color.RGBA{R: 0xc9, G: 0xd4, B: 0x87, A: 0xff}, // Yellow      - 7
	color.RGBA{R: 0xa1, G: 0x68, B: 0x3c, A: 0xff}, // Orange      - 8
	color.RGBA{R: 0x6d, G: 0x54, B: 0x12, A: 0xff}, // Brown       - 9
	color.RGBA{R: 0xcb, G: 0x7e, B: 0x75, A: 0xff}, // Light Red   - 10
	color.RGBA{R: 0x62, G: 0x62, B: 0x62, A: 0xff}, // Dark Gray   - 11
	color.RGBA{R: 0x89, G: 0x89, B: 0x89, A: 0xff}, // Mid-Gray    - 12
	color.RGBA{R: 0x9a, G: 0xe2, B: 0x9b, A: 0xff}, // Light Green - 13
	color.RGBA{R: 0x88, G: 0x7e, B: 0xcb, A: 0xff}, // Light Blue  - 14
	color.RGBA{R: 0xad, G: 0xad, B: 0xad, A: 0xff}, // Light Gray  - 15
}

var colorToIndex map[colorRGBA]int

const (
	imgWidth          = 40
	imgHeight         = 25
	imgOutNameDefault = "out"

	// BASIC program constants
	dataRowSize        = 20   // Number of color values per DATA line
	basicStartLine     = 1000 // Starting line number for DATA statements
	basicLineIncrement = 10   // Increment between DATA line numbers

	header = "" +
		"10 for y = 0 to 24\n" +
		"20 for x = 0 to 39\n" +
		"30 o = 40 * y + x\n" +
		"40 poke 1024 + o, 160\n" +
		"45 read c\n" +
		"50 poke 55296 + o, c\n" +
		"60 next x,y\n" +
		"70 goto 70\n"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	colorToIndex = make(map[colorRGBA]int, len(colorPalette))
	for i, c := range colorPalette {
		colorToIndex[transformRGBAToRGBAColor(c)] = i
	}
}

func main() {
	inputImage := flag.String("i", "", "path to the source image (required)")
	outputImage := flag.String("o", "", "path to the output image (default out.***)")
	outputFile := flag.String("f", "img.basic", "path to the export file with basic program (will be generated)")
	dither := flag.Bool("dither", false, "use Floyd–Steinberg dithering algorithm (default \"false\")")

	flag.Parse()

	if *inputImage == "" {
		fmt.Println("-i flag is required. Type -help for help")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := processImage(*inputImage, *outputImage, *outputFile, *dither); err != nil {
		log.Fatal(err)
	}

	fmt.Println("finished")
}

func processImage(inputPath, outputPath, basicPath string, useDither bool) error {
	imageFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer imageFile.Close()

	imageConfig, _, err := image.DecodeConfig(imageFile)
	if err != nil {
		return fmt.Errorf("failed to decode image config: %w", err)
	}

	if imageConfig.Width != imgWidth || imageConfig.Height != imgHeight {
		return fmt.Errorf("wrong image size: expected %dx%d, got %dx%d",
			imgWidth, imgHeight, imageConfig.Width, imageConfig.Height)
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf("%s%s", imgOutNameDefault, filepath.Ext(inputPath))
	}

	if _, err := imageFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to start of file: %w", err)
	}

	img, _, err := image.Decode(imageFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	dst := image.NewPaletted(img.Bounds(), colorPalette)
	drawer := draw.Drawer(draw.Src)
	if useDither {
		drawer = draw.FloydSteinberg
	}
	drawer.Draw(dst, dst.Bounds(), img, img.Bounds().Min)

	if err := saveImage(dst, outputPath); err != nil {
		return err
	}

	points := pointsFromImage(*dst)
	if err := generateBASICProgram(points, basicPath); err != nil {
		return err
	}

	return nil
}

func saveImage(img *image.Paletted, path string) error {
	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output image: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

func pointsFromImage(dst image.Paletted) []int {
	points := make([]int, 0, imgWidth*imgHeight)

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			colorCode := pixelColorCode(dst.At(x, y))
			points = append(points, colorCode)
		}
	}

	return points
}

func pixelColorCode(c color.Color) int {
	colorFromPixel := transformRGBAToRGBAColor(c)

	if index, exists := colorToIndex[colorFromPixel]; exists {
		return index
	}

	return 0
}

func transformRGBAToRGBAColor(c color.Color) colorRGBA {
	r, g, b, a := c.RGBA()
	return colorRGBA{R: r, G: g, B: b, A: a}
}

func generateBASICProgram(points []int, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create BASIC file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	rows := splitIntoRows(points, dataRowSize)

	lineNum := basicStartLine
	for _, row := range rows {
		colorSeq := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(row)), ","), "[]")
		if _, err := file.WriteString(fmt.Sprintf("%d data %s\n", lineNum, colorSeq)); err != nil {
			return fmt.Errorf("failed to write data line %d: %w", lineNum, err)
		}
		lineNum += basicLineIncrement
	}

	return nil
}

func splitIntoRows(points []int, rowSize int) [][]int {
	var rows [][]int
	for i := 0; i < len(points); i += rowSize {
		end := i + rowSize
		if end > len(points) {
			end = len(points)
		}
		rows = append(rows, points[i:end])
	}
	return rows
}
