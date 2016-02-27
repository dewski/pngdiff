package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dewski/pngdiff/Godeps/_workspace/src/github.com/quipo/statsd"
	"github.com/dewski/pngdiff/cmd/pngdiff"
)

var stats *statsd.StatsdClient

func render500(rw http.ResponseWriter, err error) {
	rw.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(rw, "{\"error\": \"%s\"}", err)
}

func createStatsd() (client *statsd.StatsdClient) {
	prefix := os.Getenv("STATSD_PREFIX")
	host := os.Getenv("STATSD_HOST")
	if prefix == "" {
		prefix = "pngdiff."
	}
	if host == "" {
		host = "localhost:8126"
	}

	client = statsd.NewStatsdClient(host, prefix)
	err := client.CreateSocket()
	if nil != err {
		log.Println(err)
		os.Exit(1)
	}

	return
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

	stats = createStatsd()

	http.HandleFunc("/process", func(rw http.ResponseWriter, r *http.Request) {

		start := time.Now()
		rw.Header().Set("Content-Type", "application/json")

		values := r.URL.Query()
		baseURL := values.Get("base_url")
		compareURL := values.Get("compare_url")

		if !validURL(baseURL) || !validURL(compareURL) {
			stats.Incr("path.process.request.status.400", 1)
			fmt.Printf("path=/process duration=400 base_url=%s compare_url=%s\n", baseURL, compareURL)
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, "{\"error\": \"Missing base_url or compare_url\"}")
			return
		}

		additions, deletions, diffs, changes, err := pngdiff.Diff(baseURL, compareURL)

		if err != nil {
			fmt.Printf("path=/process status=500 took=%s\n", time.Since(start))
			stats.Incr("path.process.request.status.500", 1)

			render500(rw, err)
		} else {
			duration := time.Since(start)

			stats.Incr("path.process.request.status.200", 1)
			stats.Timing("path.process.request.duration", int64(duration.Seconds()*1000))

			fmt.Printf("path=/process duration=200 took=%s base_url=%s compare_url=%s\n", duration, baseURL, compareURL)

			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, "{\"additions\": %d, \"deletions\": %d, \"diffs\": %d, \"changes\": %f}", additions, deletions, diffs, changes)
		}
	})

	http.HandleFunc("/_ping", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "OK - %s", time.Now())
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
