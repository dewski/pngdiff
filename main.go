package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
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
			fmt.Fprint(rw, "{\"error\": \"Missing base_url or compare_url\"}")
			return
		}

		additions, deletions, diffs, changes, err := pngdiff.Diff(baseURL, compareURL)

		if err != nil {
			fmt.Printf("path=/process status=500 took=%s\n", time.Since(start))

			render500(rw, err)
		} else {
			duration := time.Since(start)

			fmt.Printf("path=/process duration=200 took=%s base_url=%s compare_url=%s\n", duration, baseURL, compareURL)

			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, "{\"additions\": %d, \"deletions\": %d, \"diffs\": %d, \"changes\": %f}", additions, deletions, diffs, changes)
		}
	})

	http.HandleFunc("/bounds", func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		values := r.URL.Query()
		imageURL := values.Get("image_url")

		if !validURL(imageURL) {
			fmt.Printf("path=/bounds duration=400 image_url=%s\n", imageURL)
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, "{\"error\": \"Missing valid image_url\"}")
			return
		}

		_, err := pngdiff.DetectRegions(imageURL)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("path=/bounds status=500 took=%s\n", duration)

			render500(rw, err)
		} else {
			fmt.Printf("path=/bounds duration=200 took=%s image_url=%s\n", duration, imageURL)

			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, "{\"bounds\": %d}", 1)
		}
	})

	http.HandleFunc("/_ping", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "OK - %s", time.Now())
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
