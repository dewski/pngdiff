all: install

install: build
	mkdir -p bin
	mv main bin/pngdiff

build: main.go
	go build main.go

run: install
	bin/pngdiff
