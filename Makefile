all: build

build:
	@echo "Building..."
	@go build -o main cmd/scraper/main.go

run:
	@go run cmd/scraper/main.go

clean:
	@echo "Cleaning..."
	@rm -f main

.PHONY: all build run test clean
