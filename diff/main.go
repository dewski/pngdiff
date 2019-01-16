package main

import (
	"context"
	"errors"
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
func ProcessDiff(ctx context.Context, event Payload) (Response, error) {
	if !validURL(event.BaseURL) {
		return Response{}, errors.New("missing valid base_url")
	}

	if !validURL(event.CompareURL) {
		return Response{}, errors.New("missing valid compare_url")
	}

	baseImage, err := pngdiff.DownloadImage(event.BaseURL)
	if err != nil {
		return Response{}, errors.New("could not download base_url")
	}

	compareImage, err := pngdiff.DownloadImage(event.CompareURL)
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
