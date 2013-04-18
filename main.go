package main

import (
	"./pngdiff"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		basePath := values["base"][0]
		comparePath := values["compare"][0]
		additions, deletions, diffs, err := pngdiff.Diff(basePath, comparePath)

		if err == nil {
			fmt.Fprintf(w, "{\"additions\": %d, \"deletions\": %d, \"diffs\": %d}", additions, deletions, diffs)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
