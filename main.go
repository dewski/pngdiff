package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"./pngdiff"
)

func render404(rw http.ResponseWriter, err error) {
	fmt.Fprintf(rw, "{\"error\": \"%s\"}", err)
}

func main() {
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
	log.Fatal(http.ListenAndServe(":1339", nil))
}
