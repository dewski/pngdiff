.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./diff/diff

build:
	GOOS=linux GOARCH=amd64 go build -o diff/diff ./diff
