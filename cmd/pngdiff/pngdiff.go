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
	label int
	X1    int `json:"x1"`
	Y1    int `json:"y1"`
	X2    int `json:"x2"`
	Y2    int `json:"y2"`
}

// Relative luminance
func relativeLuminance(pixel color.Color) float64 {
	r, g, b, _ := pixel.RGBA()
	return (0.2126 * float64(r)) + (0.7152 * float64(g)) + (0.0722 * float64(b))
}

// DetectRegions finds regions
// Uses Connected-component labeling https://en.wikipedia.org/wiki/Connected-component_labeling
func DetectRegions(imageURL string) (regions []*Region, err error) {
	sourceImage, err := loadImage(imageURL)
	if err != nil {
		return
	}

	imageWidth := sourceImage.Bounds().Dx()
	imageHeight := sourceImage.Bounds().Dy()
	var pos int

	blobMap := make([][]int, imageHeight)
	label := 1
	labelTable := []int{0}

	// Label every pixel as 0
	for y := 0; y < imageHeight; y++ {
		blobRow := make([]int, imageWidth)
		blobMap[y] = blobRow
	}

	// Variables for neigboring pixels
	imageData := sourceImage.(*image.NRGBA)
	var nn, nw, ne, ww, ee, sw, ss, se, minIndex int
	isVisible := false

	// Need to make two passes
	// First to identify all of the blob candidates
	// Second pass merges any blobs that the first pass failed to merge
	nIter := 2
	for nIter > 0 {
		// Leave a 1 pixel border which is ignored so we do not get array out of
		// bound errors
		for y := 1; y < imageHeight-1; y++ {
			for x := 1; x < imageWidth-1; x++ {
				pos = (y*imageWidth + x) * 4

				// We're only looking at the alpa channel in this case
				if imageData.Pix[pos+3] > 127 {
					isVisible = true
				} else {
					isVisible = false
				}

				if isVisible {
					nw = blobMap[y-1][x-1]
					nn = blobMap[y-1][x-0]
					ne = blobMap[y-1][x+1]
					ww = blobMap[y-0][x-1]
					ee = blobMap[y-0][x+1]
					sw = blobMap[y+1][x-1]
					ss = blobMap[y+1][x-0]
					se = blobMap[y+1][x+1]
					minIndex = ww

					if 0 < ww && ww < minIndex {
						minIndex = ww
					}

					if 0 < ee && ee < minIndex {
						minIndex = ee
					}

					if 0 < nn && nn < minIndex {
						minIndex = nn
					}

					if 0 < ne && ne < minIndex {
						minIndex = ne
					}

					if 0 < nw && nw < minIndex {
						minIndex = nw
					}

					if 0 < ss && ss < minIndex {
						minIndex = ss
					}

					if 0 < se && se < minIndex {
						minIndex = se
					}

					if 0 < sw && sw < minIndex {
						minIndex = sw
					}

					if minIndex == 0 {
						blobMap[y][x] = label
						labelTable = append(labelTable, label)
						label++
					} else {
						if minIndex < labelTable[nw] {
							labelTable[nw] = minIndex
						}

						if minIndex < labelTable[nn] {
							labelTable[nn] = minIndex
						}

						if minIndex < labelTable[ne] {
							labelTable[ne] = minIndex
						}

						if minIndex < labelTable[ww] {
							labelTable[ww] = minIndex
						}

						if minIndex < labelTable[ee] {
							labelTable[ee] = minIndex
						}

						if minIndex < labelTable[sw] {
							labelTable[sw] = minIndex
						}

						if minIndex < labelTable[ss] {
							labelTable[ss] = minIndex
						}

						if minIndex < labelTable[se] {
							labelTable[se] = minIndex
						}

						blobMap[y][x] = minIndex
					}
				} else {
					blobMap[y][x] = 0
				}
			}

		}
		nIter--
	}

	// Compress the table of labels so that every location refers to only 1
	// matching location
	for i := range labelTable {
		label = labelTable[i]
		for label != labelTable[label] {
			label = labelTable[label]
		}
		labelTable[i] = label
	}

	// Merge the blobs with multiple labels
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			// fmt.Printf("x=%d y=%d\n", x, y)
			label = blobMap[y][x]
			if label == 0 {
				continue
			}

			for label != labelTable[label] {
				label = labelTable[label]
			}
			blobMap[y][x] = label
		}
	}

	// Conver the blobs to minimized labels
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			label = blobMap[y][x]
			blobMap[y][x] = labelTable[label]
		}
	}

	blobs := map[int]*Region{}
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			l := blobMap[y][x]
			if l <= 0 {
				continue
			}

			b := blobs[l]
			if b != nil {
				// fmt.Printf("l=%d X1=%d Y1=%d X2=%d Y2=%d, x=%d y=%d\n", l, b.X1, b.Y1, b.X2, b.Y2, x, y)

				if b.X1 > x {
					b.X1 = x
				}

				if b.X2 < x {
					b.X2 = x
				}

				if b.Y2 < y {
					b.Y2 = y
				}
			} else {
				blobs[l] = &Region{
					label: l,
					X1:    x,
					Y1:    y,
					X2:    x,
					Y2:    y,
				}
			}
		}
	}

	for key := range blobs {
		regions = append(regions, blobs[key])
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
