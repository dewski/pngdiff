package pngdiff

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type Pixel struct {
	X, Y int
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Couldn't load the file.")
	}

	loadedImage, err := png.Decode(file)
	if err != nil {
		fmt.Println("Couldn't decode the PNG.")
	}

	return loadedImage, nil
}

func maxHeight(baseImage, targetImage image.Image) int {
	baseHeight := float64(baseImage.Bounds().Dy())
	targetHeight := float64(targetImage.Bounds().Dy())
	return int(math.Max(baseHeight, targetHeight))
}

func samePixel(basePixel color.Color, comparePixel color.RGBA) bool {
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

func Diff(basePath string, targetPath string) {
	// additions := []int{}
	// deletions := []int{}
	// diffs     := []int{}

	baseImage, err := loadImage(basePath)
	if err != nil {
	}
	targetImage, err := loadImage(targetPath)
	if err != nil {
	}

	maxHeight := maxHeight(baseImage, targetImage)

	for y := 0; y < maxHeight; y++ {
		if emptyPixel(baseImage.At(0, y)) {
			fmt.Println("baseImage does not have y offset ", y)
		}
	}
	// fmt.Println(maxHeight)

	// fmt.Println(basePath)
	// fmt.Println(baseImage.Bounds().Size().X)
	// fmt.Println(baseImage.At(1401, 1164))

	// fmt.Println(targetPath)
	// fmt.Println(targetImage.Bounds().Size().X)
}
