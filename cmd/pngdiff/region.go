package pngdiff

import (
	"image"
	"image/color"
)

// Region is an area
type Region struct {
	label int
	X1    int `json:"x1"`
	Y1    int `json:"y1"`
	X2    int `json:"x2"`
	Y2    int `json:"y2"`
}

// Width calculates the region's width
func (r *Region) Width() int {
	return r.X2 - r.X1
}

// Height calculates the region's height
func (r *Region) Height() int {
	return r.Y2 - r.Y1
}

// Area calculates the region's total size
func (r *Region) Area() int {
	return r.Width() * r.Height()
}

// Relative luminance
func relativeLuminance(pixel color.Color) float64 {
	r, g, b, _ := pixel.RGBA()
	return (0.2126 * float64(r)) + (0.7152 * float64(g)) + (0.0722 * float64(b))
}

// MinimumRegionArea defines how big a region must be.
const MinimumRegionArea = 25

// DetectRegions finds regions
// Uses Connected-component labeling https://en.wikipedia.org/wiki/Connected-component_labeling
func DetectRegions(imageURL string) (regions []*Region, err error) {
	sourceImage, err := loadImage(imageURL)
	if err != nil {
		return
	}

	imageWidth := sourceImage.Bounds().Dx()
	imageHeight := sourceImage.Bounds().Dy()

	imageData := sourceImage.(*image.NRGBA)
	var nn, nw, ne, ww, ee, sw, ss, se, minIndex int
	var pos int

	// Keeps track of label keys
	blobMap := make([][]int, imageHeight)
	labelCounter := 1
	labels := []int{0}

	// Variables for neigboring pixels

	isVisible := false

	// Label every pixel as 0
	for y := 0; y < imageHeight; y++ {
		blobMap[y] = make([]int, imageWidth)
	}

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

				// Don't want faintly visible pixels to start the region
				if imageData.Pix[pos+3] > 127 {
					isVisible = true
				} else {
					isVisible = false
				}

				if isVisible {
					nw = blobMap[y-1][x-1] // top left
					nn = blobMap[y-1][x-0] // above
					ne = blobMap[y-1][x+1] // top right
					ww = blobMap[y-0][x-1] // left
					ee = blobMap[y-0][x+1] // right
					sw = blobMap[y+1][x-1] // bottom left
					ss = blobMap[y+1][x-0] // beneath
					se = blobMap[y+1][x+1] // bottom right
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
						blobMap[y][x] = labelCounter
						labels = append(labels, labelCounter)
						labelCounter++
					} else {
						if minIndex < labels[nw] {
							labels[nw] = minIndex
						}

						if minIndex < labels[nn] {
							labels[nn] = minIndex
						}

						if minIndex < labels[ne] {
							labels[ne] = minIndex
						}

						if minIndex < labels[ww] {
							labels[ww] = minIndex
						}

						if minIndex < labels[ee] {
							labels[ee] = minIndex
						}

						if minIndex < labels[sw] {
							labels[sw] = minIndex
						}

						if minIndex < labels[ss] {
							labels[ss] = minIndex
						}

						if minIndex < labels[se] {
							labels[se] = minIndex
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
	for i, label := range labels {
		for label != labels[label] {
			label = labels[label]
		}

		labels[i] = label
	}

	// Merge the blobs with multiple labels
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			label := blobMap[y][x]
			if label == 0 {
				continue
			}

			for label != labels[label] {
				label = labels[label]
			}
			blobMap[y][x] = label
		}
	}

	// Since the same label key:value pair are appended to the labelTable map,
	// we can just lookup. Not sure if necessary in practice?
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			label := blobMap[y][x]
			if label == 0 || label == labels[label] {
				continue
			}

			blobMap[y][x] = labels[label]
		}
	}

	blobs := map[int]*Region{}
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			label := blobMap[y][x]
			if label <= 0 {
				continue
			}

			if b := blobs[label]; b != nil {
				// The start of the region is before the first recorded label
				if b.X1 > x {
					b.X1 = x
				}

				// The known south-east region has been extended
				if b.X2 < x {
					b.X2 = x
				}

				// The known south-east region has been extended
				if b.Y2 < y {
					b.Y2 = y
				}
			} else {
				// Encountered a label for the first time, establish the region with
				// kwnon coordinates.
				blobs[label] = &Region{
					label: label,
					X1:    x,
					Y1:    y,
					X2:    x,
					Y2:    y,
				}
			}
		}
	}

	for _, r := range blobs {
		regions = append(regions, r)
	}

	return
}
