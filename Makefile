.DEFAULT_GOAL := help

.PHONY: help build clean dist distclean test test-cmd

help:
	@echo "Available targets:"
	@echo "  build      Build test program at ./bin/exiftool-go"
	@echo "  clean      Remove build artifacts (./bin)"
	@echo "  distclean  Remove distribution/build artifacts (alias: clean)"
	@echo "  test       Run module tests with coverage"

build:
	mkdir -p ./bin
	go build -o ./bin/exiftool-go ./cmd/exiftool-go

test:
	go test ./... -cover


clean:
	rm -rf ./bin

distclean: clean
