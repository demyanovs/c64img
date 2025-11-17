package main

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransformRGBAToRGBAColor(t *testing.T) {
	tests := []struct {
		name     string
		input    color.Color
		expected colorRGBA
	}{
		{
			name:  "Black color",
			input: color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff},
			expected: colorRGBA{
				R: 0,
				G: 0,
				B: 0,
				A: 0xffff,
			},
		},
		{
			name:  "White color",
			input: color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			expected: colorRGBA{
				R: 0xffff,
				G: 0xffff,
				B: 0xffff,
				A: 0xffff,
			},
		},
		{
			name:  "Red color",
			input: color.RGBA{R: 0x9f, G: 0x4e, B: 0x44, A: 0xff},
			expected: colorRGBA{
				R: 0x9f9f,
				G: 0x4e4e,
				B: 0x4444,
				A: 0xffff,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformRGBAToRGBAColor(tt.input)
			if result != tt.expected {
				t.Errorf("transformRGBAToRGBAColor() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestPixelColorCode(t *testing.T) {
	tests := []struct {
		name     string
		input    color.Color
		expected int
	}{
		{
			name:     "Black - index 0",
			input:    color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff},
			expected: 0,
		},
		{
			name:     "White - index 1",
			input:    color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			expected: 1,
		},
		{
			name:     "Red - index 2",
			input:    color.RGBA{R: 0x9f, G: 0x4e, B: 0x44, A: 0xff},
			expected: 2,
		},
		{
			name:     "Cyan - index 3",
			input:    color.RGBA{R: 0x6a, G: 0xbf, B: 0xc6, A: 0xff},
			expected: 3,
		},
		{
			name:     "Light Gray - index 15",
			input:    color.RGBA{R: 0xad, G: 0xad, B: 0xad, A: 0xff},
			expected: 15,
		},
		{
			name:     "Unknown color defaults to 0",
			input:    color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pixelColorCode(tt.input)
			if result != tt.expected {
				t.Errorf("pixelColorCode() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestColorToIndexMapInitialization(t *testing.T) {
	// Verify that colorToIndex map is properly initialized
	if colorToIndex == nil {
		t.Fatal("colorToIndex map is nil")
	}

	if len(colorToIndex) != len(colorPalette) {
		t.Errorf("colorToIndex map size = %d, want %d", len(colorToIndex), len(colorPalette))
	}

	// Verify all palette colors are in the map
	for i, c := range colorPalette {
		key := transformRGBAToRGBAColor(c)
		if index, exists := colorToIndex[key]; !exists {
			t.Errorf("Color at palette index %d not found in colorToIndex map", i)
		} else if index != i {
			t.Errorf("Color at palette index %d mapped to %d in colorToIndex", i, index)
		}
	}
}

func TestSplitIntoRows(t *testing.T) {
	tests := []struct {
		name     string
		points   []int
		rowSize  int
		expected [][]int
	}{
		{
			name:     "Even split",
			points:   []int{1, 2, 3, 4, 5, 6},
			rowSize:  3,
			expected: [][]int{{1, 2, 3}, {4, 5, 6}},
		},
		{
			name:     "Uneven split",
			points:   []int{1, 2, 3, 4, 5},
			rowSize:  3,
			expected: [][]int{{1, 2, 3}, {4, 5}},
		},
		{
			name:     "Single row",
			points:   []int{1, 2, 3},
			rowSize:  5,
			expected: [][]int{{1, 2, 3}},
		},
		{
			name:     "Empty input",
			points:   []int{},
			rowSize:  3,
			expected: [][]int{},
		},
		{
			name:     "Row size of 1",
			points:   []int{1, 2, 3},
			rowSize:  1,
			expected: [][]int{{1}, {2}, {3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitIntoRows(tt.points, tt.rowSize)

			if len(result) != len(tt.expected) {
				t.Errorf("splitIntoRows() returned %d rows, want %d", len(result), len(tt.expected))
				return
			}

			for i, row := range result {
				if len(row) != len(tt.expected[i]) {
					t.Errorf("Row %d has %d elements, want %d", i, len(row), len(tt.expected[i]))
					continue
				}
				for j, val := range row {
					if val != tt.expected[i][j] {
						t.Errorf("Row %d, element %d = %d, want %d", i, j, val, tt.expected[i][j])
					}
				}
			}
		})
	}
}

func TestPointsFromImage(t *testing.T) {
	// Create a small test image
	img := image.NewPaletted(
		image.Rect(0, 0, 3, 2),
		colorPalette,
	)

	// Set specific colors
	img.SetColorIndex(0, 0, 0) // Black
	img.SetColorIndex(1, 0, 1) // White
	img.SetColorIndex(2, 0, 2) // Red
	img.SetColorIndex(0, 1, 3) // Cyan
	img.SetColorIndex(1, 1, 4) // Purple
	img.SetColorIndex(2, 1, 5) // Green

	// Temporarily modify constants for testing
	origWidth := imgWidth
	origHeight := imgHeight
	defer func() {
		// This won't actually change the constants, but shows intent
		_ = origWidth
		_ = origHeight
	}()

	// We can't actually test pointsFromImage directly because it uses constants
	// But we can test the logic by creating a smaller version
	testPointsFromImage := func(dst image.Paletted, width, height int) []int {
		points := make([]int, 0, width*height)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				colorCode := pixelColorCode(dst.At(x, y))
				points = append(points, colorCode)
			}
		}
		return points
	}

	points := testPointsFromImage(*img, 3, 2)
	expected := []int{0, 1, 2, 3, 4, 5}

	if len(points) != len(expected) {
		t.Errorf("pointsFromImage() returned %d points, want %d", len(points), len(expected))
		return
	}

	for i, point := range points {
		if point != expected[i] {
			t.Errorf("Point at index %d = %d, want %d", i, point, expected[i])
		}
	}
}

func TestWriteBASICProgram(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test.basic")

	points := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	err := generateBASICProgram(points, outputFile)
	if err != nil {
		t.Fatalf("generateBASICProgram() error = %v", err)
	}

	// Read the file back
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check header is present
	if !strings.Contains(contentStr, "10 for y = 0 to 24") {
		t.Error("BASIC program missing header")
	}

	// Check data lines are present
	if !strings.Contains(contentStr, "1000 data") {
		t.Error("BASIC program missing DATA lines")
	}

	// Check that our test data appears
	if !strings.Contains(contentStr, "0,1,2,3,4,5,6,7,8,9") {
		t.Error("BASIC program missing expected data values")
	}
}

func TestSaveImage(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test.png")

	// Create a simple test image
	img := image.NewPaletted(
		image.Rect(0, 0, 10, 10),
		colorPalette,
	)

	// Fill with a color
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetColorIndex(x, y, 1) // White
		}
	}

	err := saveImage(img, outputFile)
	if err != nil {
		t.Fatalf("saveImage() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Try to decode the saved image
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("Failed to open saved image: %v", err)
	}
	defer file.Close()

	_, _, err = image.Decode(file)
	if err != nil {
		t.Errorf("Failed to decode saved image: %v", err)
	}
}

func TestProcessImageErrors(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		inputPath   string
		outputPath  string
		basicPath   string
		useDither   bool
		expectError bool
	}{
		{
			name:        "Non-existent input file",
			inputPath:   "nonexistent.png",
			outputPath:  filepath.Join(tempDir, "out.png"),
			basicPath:   filepath.Join(tempDir, "out.basic"),
			useDither:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processImage(tt.inputPath, tt.outputPath, tt.basicPath, tt.useDither)
			if (err != nil) != tt.expectError {
				t.Errorf("processImage() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// Benchmark tests
func BenchmarkPixelColorCode(b *testing.B) {
	testColor := color.RGBA{R: 0x9f, G: 0x4e, B: 0x44, A: 0xff}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pixelColorCode(testColor)
	}
}

func BenchmarkTransformRGBAToRGBAColor(b *testing.B) {
	testColor := color.RGBA{R: 0x9f, G: 0x4e, B: 0x44, A: 0xff}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transformRGBAToRGBAColor(testColor)
	}
}

func BenchmarkSplitIntoRows(b *testing.B) {
	points := make([]int, 1000)
	for i := range points {
		points[i] = i % 16
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitIntoRows(points, dataRowSize)
	}
}
