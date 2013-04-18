package main

import (
	"./pngdiff"
	"fmt"
	"log"
	"net/http"
)

func render404(rw http.ResponseWriter, err error) {
	fmt.Fprintf(rw, "{\"error\": \"%s\"}", err)
}

func main() {
	http.HandleFunc("/process", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		values := r.URL.Query()
		basePath := values["base"][0]
		comparePath := values["compare"][0]
		additions, deletions, diffs, changes, err := pngdiff.Diff(basePath, comparePath)

		if err != nil {
			render404(rw, err)
		} else {
			fmt.Fprintf(rw, "{\"additions\": %d, \"deletions\": %d, \"diffs\": %d, \"changes\": %f}", additions, deletions, diffs, changes)
		}
	})
	log.Fatal(http.ListenAndServe(":1339", nil))
}
