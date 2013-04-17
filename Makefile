all: install

install: build
	mv main bin/
	mv bin/main bin/pngdiff

build: main.go
	go build main.go

test: install
	bin/pngdiff fixtures/large/base.png fixtures/large/target.png
