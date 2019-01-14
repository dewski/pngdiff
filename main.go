package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/dewski/pngdiff/cmd/pngdiff"
)

func render500(rw http.ResponseWriter, err error) {
	rw.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(rw, "{\"error\": \"%s\"}", err)
}

func validURL(input string) bool {
	if input == "" {
		return false
	}

	_, err := url.Parse(input)
	return err == nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1339"
	}

	http.HandleFunc("/process", func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw.Header().Set("Content-Type", "application/json")

		values := r.URL.Query()
		baseURL := values.Get("base_url")
		compareURL := values.Get("compare_url")

		if !validURL(baseURL) || !validURL(compareURL) {
			fmt.Printf("path=/process duration=400 base_url=%s compare_url=%s\n", baseURL, compareURL)
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, "{\"error\": \"Missing valid base_url and or compare_url\"}")
			return
		}

		baseImage, err := pngdiff.DownloadImage(baseURL)
		if err != nil {
			fmt.Printf("path=/process duration=500 base_url=%s compare_url=%s\n", baseURL, compareURL)
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, "{\"error\": \"Could not load base_url image\"}")
			return
		}

		compareImage, err := pngdiff.DownloadImage(compareURL)
		if err != nil {
			fmt.Printf("path=/process duration=500 base_url=%s compare_url=%s\n", baseURL, compareURL)
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, "{\"error\": \"Could not load compare_url image\"}")
			return
		}

		additions, deletions, diffs, changes, err := pngdiff.Diff(baseImage, compareImage)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("path=/process status=500 took=%s\n", duration)

			render500(rw, err)
		} else {
			fmt.Printf("path=/process duration=200 took=%s base_url=%s compare_url=%s\n", duration, baseURL, compareURL)

			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, "{\"additions\": %d, \"deletions\": %d, \"diffs\": %d, \"changes\": %f}", additions, deletions, diffs, changes)
		}
	})

	http.HandleFunc("/bounds", func(rw http.ResponseWriter, r *http.Request) {
		minimumRegionArea := pngdiff.MinimumRegionArea
		start := time.Now()
		values := r.URL.Query()
		imageURL := values.Get("image_url")
		ra := values.Get("minimum_region_area")

		if ra != "" {
			var err error
			minimumRegionArea, err = strconv.Atoi(ra)
			if err != nil {
				fmt.Printf("path=/bounds duration=400 minimum_region_area=%s\n", ra)
				rw.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(rw, "{\"error\": \"Invalid minimum_region_area\"}")
				return
			}
		}

		if !validURL(imageURL) {
			fmt.Printf("path=/bounds duration=400 image_url=%s\n", imageURL)
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, "{\"error\": \"Missing valid image_url\"}")
			return
		}

		image, err := pngdiff.DownloadImage(imageURL)
		if err != nil {
			return
		}

		regions, err := pngdiff.DetectRegions(image)
		duration := time.Since(start)

		filteredRegions := []*pngdiff.Region{}
		for _, r := range regions {
			if r.Area() < minimumRegionArea {
				continue
			}

			filteredRegions = append(filteredRegions, r)
		}

		if err != nil {
			fmt.Printf("path=/bounds status=500 took=%s\n", duration)

			render500(rw, err)
		} else {
			fmt.Printf("path=/bounds duration=200 took=%s regions=%d filteredRegions=%d image_url=%s\n", duration, len(regions), len(filteredRegions), imageURL)

			enc := json.NewEncoder(rw)
			err = enc.Encode(&filteredRegions)
			if err != nil {
				render500(rw, err)
			}
		}
	})

	http.HandleFunc("/_ping", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "OK - %s", time.Now())
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
