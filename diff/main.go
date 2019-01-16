package main

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dewski/pngdiff/cmd/pngdiff"
)

func validURL(input string) bool {
	if input == "" {
		return false
	}

	_, err := url.Parse(input)
	return err == nil
}

// Payload handles the incoming request.
type Payload struct {
	BaseURL    string `json:"base_url"`
	CompareURL string `json:"compare_url"`
}

// Response handles the diff response.
type Response struct {
	Additions int     `json:"additions"`
	Deletions int     `json:"deletions"`
	Diffs     int     `json:"diffs"`
	Changes   float64 `json:"changes"`
}

// ProcessDiff handles processing images.
func ProcessDiff(req Payload) (Response, error) {
	if !validURL(req.BaseURL) {
		return Response{}, fmt.Errorf("missing valid base_url got \"%s\"", req.BaseURL)
	}

	if !validURL(req.CompareURL) {
		return Response{}, fmt.Errorf("missing valid compare_url got \"%s\"", req.CompareURL)
	}

	baseImage, err := pngdiff.DownloadImage(req.BaseURL)
	if err != nil {
		return Response{}, errors.New("could not download base_url")
	}

	compareImage, err := pngdiff.DownloadImage(req.CompareURL)
	if err != nil {
		return Response{}, errors.New("could not download compare_url")
	}

	additions, deletions, diffs, changes, err := pngdiff.Diff(baseImage, compareImage)

	if err != nil {
		return Response{}, err
	}

	return Response{
		Additions: additions,
		Deletions: deletions,
		Diffs:     diffs,
		Changes:   changes,
	}, nil
}

func main() {
	lambda.Start(ProcessDiff)
}
