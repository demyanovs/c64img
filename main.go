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

	_ "golang.org/x/lint"
)

type colorRGBA struct {
	R uint32
	G uint32
	B uint32
	A uint32
}

var colorPallet = []color.Color{
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

const (
	imgWidth          = 40
	imgHeight         = 25
	imgOutNameDefault = "out"
	header            = "" +
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
}

func main() {
	inputImage := flag.String("i", "", "path to the source image (required)")
	outputImage := flag.String("o", "", "path to the output image (default out.***)")
	outputFile := flag.String("f", "img.basic", "path to the export file with basic program (will be generated)")
	dither := flag.Bool("dither", false, "use Floyd–Steinberg dithering algorithm (default \"false\")")

	flag.Parse()

	if *inputImage == "" {
		println("-i flag is required. Type -help for help")
		flag.PrintDefaults()
		os.Exit(1)
	}

	imageFile, err := os.Open(*inputImage)

	if err != nil {
		fmt.Println("file not found!")
		os.Exit(1)
	}

	defer imageFile.Close()

	imageConfig, _, err := image.DecodeConfig(imageFile)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *outputImage == "" {
		*outputImage = fmt.Sprintf("%s%s", imgOutNameDefault, filepath.Ext(*inputImage))
	}

	if imageConfig.Width != imgWidth || imageConfig.Height != imgHeight {
		fmt.Printf("Wrong image size. Expected %dx%d, got: %dx%d\n",
			imgWidth, imgHeight, imageConfig.Width, imageConfig.Height)
		os.Exit(1)
	}

	_, err = imageFile.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(imageFile)
	if err != nil {
		log.Fatal(err)
	}

	dst := image.NewPaletted(img.Bounds(), colorPallet)
	drawer := draw.Drawer(draw.Src)
	if *dither {
		drawer = draw.FloydSteinberg
	}
	drawer.Draw(dst, dst.Bounds(), img, img.Bounds().Min)

	dstFile, err := os.Create(*outputImage)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	err = png.Encode(dstFile, dst)
	if err != nil {
		log.Fatal(err)
	}

	points := getPoints(*dst)

	err = writeImageCode(points, *outputFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("finished")
}

func getPoints(dst image.Paletted) []int {
	var points []int

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			colorCode := detectPixelColorCode(dst.At(x, y))
			points = append(points, colorCode)
		}
	}

	return points
}

func detectPixelColorCode(color color.Color) int {
	colorFromPixel := transformRGBAToRGBAColor(color)

	for i, colorP := range colorPallet {
		colorFromPallet := transformRGBAToRGBAColor(colorP)
		if colorFromPixel == colorFromPallet {
			return i
		}
	}

	return 0
}

func transformRGBAToRGBAColor(color color.Color) colorRGBA {
	r, g, b, a := color.RGBA()
	rgbaColor := colorRGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}

	return rgbaColor
}

func writeImageCode(points []int, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(header)
	if err != nil {
		return err
	}

	rowSize := 20

	var rows [][]int
	for i := 0; i < len(points); i += rowSize {
		end := i + rowSize

		if end > len(points) {
			end = len(points)
		}

		rows = append(rows, points[i:end])
	}

	lineNum := 1000
	for _, row := range rows {
		colorSeq := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(row)), ","), "[]")
		_, err = file.WriteString(fmt.Sprintf("%d data %s \n", lineNum, colorSeq))
		if err != nil {
			fmt.Printf("error writing string: %v", err)
		}

		lineNum += 10
	}

	return nil
}
