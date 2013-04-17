package main

import (
  "os"
  "./pngdiff"
)

func main() {
    baseImagePath := os.Args[1]
    targetImagePath := os.Args[2]

    pngdiff.Diff(baseImagePath, targetImagePath)
}
