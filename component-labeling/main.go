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
	ImageURL          string `json:"image_url"`
	MinimumRegionArea int    `json:"minimum_region_area`
}

// Response handles the response.
type Response struct {
	Regions []*pngdiff.Region `json:"regions`
}

// ComponentLabeling finds components in an image and labels them.
func ComponentLabeling(ctx context.Context, event Payload) (Response, error) {
	if !validURL(event.ImageURL) {
		return Response{}, errors.New("missing valid image_url")
	}

	image, err := pngdiff.DownloadImage(event.ImageURL)
	if err != nil {
		return Response{}, errors.New("could not download image_url")
	}

	regions, err := pngdiff.DetectRegions(image)
	if err != nil {
		return Response{}, err
	}

	filteredRegions := []*pngdiff.Region{}
	for _, r := range regions {
		if r.Area() < event.MinimumRegionArea {
			continue
		}

		filteredRegions = append(filteredRegions, r)
	}

	return Response{
		Regions: filteredRegions,
	}, nil
}

func main() {
	lambda.Start(ComponentLabeling)
}
