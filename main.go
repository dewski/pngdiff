package main

import (
	"./pngdiff"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/process", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		values := r.URL.Query()
		basePath := values["base"][0]
		comparePath := values["compare"][0]
		additions, deletions, diffs, err := pngdiff.Diff(basePath, comparePath)

		if err == nil {
			fmt.Fprintf(rw, "{\"additions\": %d, \"deletions\": %d, \"diffs\": %d}", additions, deletions, diffs)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
