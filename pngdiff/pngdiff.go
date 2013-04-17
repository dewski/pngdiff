package pngdiff

import (
  "os"
  "fmt"
  "image"
  "image/png"
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

    fmt.Println(basePath)
    fmt.Println(baseImage.Bounds().Size())

    fmt.Println(targetPath)
    fmt.Println(targetImage.Bounds().Size())
}
