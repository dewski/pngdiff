package pngdiff

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

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

func maxHeight(baseImage, targetImage image.Image) int {
	baseHeight := float64(baseImage.Bounds().Dy())
	targetHeight := float64(targetImage.Bounds().Dy())
	return int(math.Max(baseHeight, targetHeight))
}

func Diff(basePath string, comparePath string) {
	baseImage, err := loadImage(basePath)
	if err != nil {
	}
	compareImage, err := loadImage(comparePath)
	if err != nil {
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
	// diffs := []uint8{}

	maxHeight := maxHeight(baseImage, compareImage)

	fmt.Println(baseData.PixOffset(1, 0))
	fmt.Println(len(compareData.Pix))

	for y := 0; y < maxHeight; y++ {
		compareY := y + 1

		// If the compare image is taller than the base image.
		if emptyPixel(baseImage.At(0, y)) {
			start := y * realCompareWidth
			finish := start + realCompareWidth

			additions = append(additions, compareData.Pix[start:finish]...)
			// If the base image is taller than the target image.
		} else if emptyPixel(compareImage.At(0, y)) {
			start := y * realBaseWidth
			finish := start + realBaseWidth

			deletions = append(deletions, baseData.Pix[start:finish]...)
		} else {
			startPixel := realBaseWidth * y
			endPixel := startPixel + realBaseWidth
			x := 0

			for i := startPixel; i < endPixel; i++ {
				realX := x + 1

				// The base image is wider than the comparison image.
				if realX > realCompareWidth {
					start := startPixel + realX
					finish := startPixel + (realBaseWidth - realX)
					deletions = append(deletions, baseData.Pix[start:finish]...)
				} else if realX == realBaseWidth && compareWidth > baseWidth {
					start := compareY * realBaseWidth
					finish := compareY * realCompareWidth
					additions = append(additions, compareData.Pix[start:finish]...)
				} else {
					basePixel := baseImage.At(x, y)
					comparePixel := compareImage.At(x, y)
					if !samePixel(currentBasePixel, currentComparePixel) {
						fmt.Println("diff")
					}
				}

				x++
			}
			// fmt.Println("baseImage and compareImage have y offset", y)
		}
	}

	fmt.Println(len(additions) / 4)
	fmt.Println(len(additions))

	fmt.Println(len(deletions) / 4)
	fmt.Println(len(deletions))
	// fmt.Println(len(test))
}
