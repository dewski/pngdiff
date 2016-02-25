package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dewski/pngdiff/cmd/pngdiff"
)

func render404(rw http.ResponseWriter, err error) {
	fmt.Fprintf(rw, "{\"error\": \"%s\"}", err)
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
		baseURL := values["base_url"][0]
		compareURL := values["compare_url"][0]
		additions, deletions, diffs, changes, err := pngdiff.Diff(baseURL, compareURL)

		if err != nil {
			fmt.Printf("path=/process status=404 took=%s\n", time.Since(start))
			render404(rw, err)
		} else {
			fmt.Printf("path=/process status=200 took=%s base_url=%s compare_url=%s\n", time.Since(start), baseURL, compareURL)
			fmt.Fprintf(rw, "{\"additions\": %d, \"deletions\": %d, \"diffs\": %d, \"changes\": %f}", additions, deletions, diffs, changes)
		}
	})

	http.HandleFunc("/_ping", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "OK - %s", time.Now())
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
