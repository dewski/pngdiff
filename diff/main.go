package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

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

// Response handles the diff response.
type Response struct {
	Additions int     `json:"additions"`
	Deletions int     `json:"deletions"`
	Diffs     int     `json:"diffs"`
	Changes   float64 `json:"changes"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	baseURL := request.QueryStringParameters["base_url"]
	if !validURL(baseURL) {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("invalid base_url got %s", baseURL)
	}

	compareURL := request.QueryStringParameters["compare_url"]
	if !validURL(compareURL) {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("invalid compare_url got %s", compareURL)
	}

	baseImage, err := pngdiff.DownloadImage(baseURL)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("could not download base_url at %s", baseURL)
	}

	compareImage, err := pngdiff.DownloadImage(compareURL)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("could not download compare_url at %s", compareURL)
	}

	additions, deletions, diffs, changes, err := pngdiff.Diff(baseImage, compareImage)

	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	response := Response{
		Additions: additions,
		Deletions: deletions,
		Diffs:     diffs,
		Changes:   changes,
	}

	json, err := json.Marshal(response)
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
