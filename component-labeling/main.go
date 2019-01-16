package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
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

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	minimumRegionArea := pngdiff.MinimumRegionArea
	ra := request.QueryStringParameters["minimum_region_area"]

	if ra != "" {
		var err error
		minimumRegionArea, err = strconv.Atoi(ra)
		if err != nil {
			return events.APIGatewayProxyResponse{}, fmt.Errorf("invalid minimum_region_area must be an integer")
		}
	}

	imageURL := request.QueryStringParameters["image_url"]
	if !validURL(imageURL) {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("missing valid image_url got \"%s\"", imageURL)
	}

	image, err := pngdiff.DownloadImage(imageURL)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("could not download %s", imageURL)
	}

	regions, err := pngdiff.DetectRegions(image)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	filteredRegions := []*pngdiff.Region{}
	for _, r := range regions {
		if r.Area() < minimumRegionArea {
			continue
		}

		filteredRegions = append(filteredRegions, r)
	}

	json, err := json.Marshal(filteredRegions)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body: string(json),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
