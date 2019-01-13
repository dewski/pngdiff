package pngdiff

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"
)

func downloadFile(url string) (path string, err error) {
	// Create the file
	tmpfile, err := ioutil.TempFile("", "screenshot")
	if err != nil {
		return
	}
	path = tmpfile.Name()

	// Get the data
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	fmt.Println(url, time.Since(start))
	defer resp.Body.Close()

	// Writer the body to file
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	size, err := tmpfile.Write(contents)
	if err != nil {
		return
	}
	fmt.Printf("url=%s size=%d tmpfile=%s\n", url, size, path)

	return
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("Couldn't load the file.")
	}

	loadedImage, err := png.Decode(file)
	if err != nil {
		return nil, errors.New("Couldn't decode the PNG.")
	}

	return loadedImage, nil
}

func fetchImage(url string) (image.Image, error) {
	path, err := downloadFile(url)
	if err != nil {
		return nil, err
	}

	defer os.Remove(path)

	image, err := loadImage(path)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func samePixel(basePixel, comparePixel color.Color) bool {
	baseR, baseG, baseB, baseA := basePixel.RGBA()
	compareR, compareG, compareB, compareA := comparePixel.RGBA()

	return baseR == compareR &&
		baseG == compareG &&
		baseB == compareB &&
		baseA == compareA
}

func emptyPixel(basePixel color.Color) bool {
	emptyPixel := color.RGBA{0, 0, 0, 0}
	return samePixel(basePixel, emptyPixel)
}

func maxHeight(baseImage, compareImage image.Image) int {
	baseHeight := float64(baseImage.Bounds().Dy())
	compareHeight := float64(compareImage.Bounds().Dy())
	return int(math.Max(baseHeight, compareHeight))
}

// Region is an area
type Region struct {
	x1 int
	y1 int
	x2 int
	y2 int
}

// Relative luminance
func relativeLuminance(pixel color.Color) float64 {
	r, g, b, _ := pixel.RGBA()
	return (0.2126 * float64(r)) + (0.7152 * float64(g)) + (0.0722 * float64(b))
}

// DetectRegions finds regions
// Uses Connected-component labeling https://en.wikipedia.org/wiki/Connected-component_labeling
func DetectRegions(imageURL string) (regions []Region, err error) {
	sourceImage, err := loadImage(imageURL)
	if err != nil {
		return
	}

	imageWidth := sourceImage.Bounds().Dx()
	imageHeight := sourceImage.Bounds().Dy()

	blobPixels := make([]int, imageWidth*imageHeight)

	data := []uint32{}
	cB := 1

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			pixel := sourceImage.At(x, y)
			_, _, _, alpha := pixel.RGBA()
			lum := relativeLuminance(pixel)

			if lum >= 127 {
				data = append(data, 255, 255, 255, alpha)
			} else {
				data = append(data, 0, 0, 0, alpha)
			}
		}
	}

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			position := (x + y*imageWidth) * 4
			pixel := data[position]

			if pixel == 255 {
				if blobPixels[position] == 0 {
					eB := 0
					for xB := -1; xB < 2; xB++ {
						if eB == 0 {
							for yB := -1; yB < 2; yB++ {
								bPos := (x + xB) + (y+yB)*imageWidth
								if blobPixels[bPos] != 0 && eB == 0 {
									eB = blobPixels[bPos]
									break
								}
							}
						} else {
							break
						}
					}

					if eB == 0 {
						blobPixels[position] = cB

					} else {
						blobPixels[position] = eB
					}
				}
			}
		}
	}

	return
}

// Diff is cool
func Diff(baseURL string, compareURL string) (additionsCount int, deletionsCount int, diffsCount int, changesCount float64, err error) {
	baseImage, err := fetchImage(baseURL)
	if err != nil {
		return 0, 0, 0, 0.0, errors.New("Couldn't decode the base image.")
	}

	compareImage, err := fetchImage(compareURL)
	if err != nil {
		return 0, 0, 0, 0.0, errors.New("Couldn't decode the comparison image.")
	}

	baseData := baseImage.(*image.NRGBA)
	compareData := compareImage.(*image.NRGBA)

	// Move this into Struct
	baseWidth := baseImage.Bounds().Dx()
	realBaseWidth := baseWidth * 4
	compareWidth := compareImage.Bounds().Dx()
	realCompareWidth := compareWidth * 4

	additions := []uint8{}
	deletions := []uint8{}
	diffs := []uint8{}

	maxHeight := maxHeight(baseImage, compareImage)

	for y := 0; y < maxHeight; y++ {
		compareY := y + 1

		if emptyPixel(baseImage.At(0, y)) {
			start := y * realCompareWidth
			finish := start + realCompareWidth

			additions = append(additions, compareData.Pix[start:finish]...)
		} else if emptyPixel(compareImage.At(0, y)) {
			start := y * realBaseWidth
			finish := start + realBaseWidth

			deletions = append(deletions, baseData.Pix[start:finish]...)
		} else {
			startPixel := baseWidth * y
			endPixel := startPixel + baseWidth
			x := 0

			for i := startPixel; i < endPixel; i++ {
				realX := x + 1

				if realX == baseWidth && compareWidth > baseWidth {
					start := compareData.PixOffset(realX, y)
					finish := compareY * realCompareWidth
					additions = append(additions, compareData.Pix[start:finish]...)
				} else if realX == compareWidth && baseWidth > compareWidth {
					start := baseData.PixOffset(realX, y)
					finish := compareY * realBaseWidth
					deletions = append(deletions, baseData.Pix[start:finish]...)
				} else {
					basePixel := baseImage.At(x, y)
					comparePixel := compareImage.At(x, y)
					if !samePixel(basePixel, comparePixel) {
						diffs = append(diffs, make([]uint8, 4)...)
					}
				}

				x++
			}
		}
	}

	additionsCount = len(additions) / 4
	deletionsCount = len(deletions) / 4
	diffsCount = len(diffs) / 4

	totalChanges := additionsCount + deletionsCount + diffsCount
	baseHeight := float64(baseImage.Bounds().Dy())
	baseArea := float64(float64(baseWidth) * baseHeight)
	changesCount = (float64(totalChanges) / baseArea) * 100

	return
}
