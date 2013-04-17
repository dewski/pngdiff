all: install

install: build
	mv pngdiff bin/

build: pngdiff.go
	go build pngdiff.go

test: install
	bin/pngdiff fixtures/large/base.png fixtures/large/target.png
