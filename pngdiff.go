package pngdiff

import "os"
import "fmt"
import "image"
import "image/png"

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

func main() {
    baseImagePath := os.Args[1]
    targetImagePath := os.Args[2]

    baseImage, err := loadImage(baseImagePath)
    if err != nil {
    }
    targetImage, err := loadImage(targetImagePath)
    if err != nil {
    }

    fmt.Println(baseImagePath)
    fmt.Println(baseImage.Bounds().Size())

    fmt.Println(targetImagePath)
    fmt.Println(targetImage.Bounds().Size())
}
