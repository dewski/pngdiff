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
)

func downloadFile(url string) (path string, err error) {
	// Create the file
	tmpfile, err := ioutil.TempFile("", "screenshot")
	if err != nil {
		return
	}
	path = tmpfile.Name()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return
	}
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
		return nil, errors.New("couldn't load the file")
	}

	loadedImage, err := png.Decode(file)
	if err != nil {
		return nil, errors.New("couldn't decode the PNG")
	}

	return loadedImage, nil
}

// DownloadImage loads an image from disk or downloads it from URL
func DownloadImage(url string) (image.Image, error) {
	// Load the image from disk if it's available
	if img, err := loadImage(url); err == nil {
		return img, nil
	}

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

// Diff is cool
func Diff(baseImage, compareImage image.Image) (additionsCount int, deletionsCount int, diffsCount int, changesCount float64, err error) {
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
		if emptyPixel(baseImage.At(0, y)) {
			start := y * realCompareWidth
			finish := start + realCompareWidth

			additions = append(additions, compareData.Pix[start:finish]...)
		} else if emptyPixel(compareImage.At(0, y)) {
			start := y * realBaseWidth
			finish := start + realBaseWidth

			deletions = append(deletions, baseData.Pix[start:finish]...)
		} else {
			compareY := y + 1
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
