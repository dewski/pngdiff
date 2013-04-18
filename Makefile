all: install

install: build
	mv main bin/
	mv bin/main bin/pngdiff

build: main.go
	go build main.go

run: install
	bin/pngdiff
