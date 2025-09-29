# Makefile for rex

# Default target
all: test build

# Build the rex binary
build:
	go build ./pkg/...
	go build -o rex .

test:
	go test ./pkg/...

# Remove the built binary
clean:
	rm -f rex

.PHONY: all build clean test
