.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./diff/diff
	rm -rf ./component-labeling/component-labeling

build:
	GOOS=linux GOARCH=amd64 go build -o diff/diff ./diff
	GOOS=linux GOARCH=amd64 go build -o component-labeling/component-labeling ./component-labeling
