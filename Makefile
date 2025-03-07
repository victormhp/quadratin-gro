all: build

build:
	@echo "Building..."
	@go build -o bin/scraper cmd/scraper/main.go

run-scraper:
	@go run ./cmd/scraper

clean:
	@echo "Cleaning..."
	@rm -rf bin

clean-data:
	@echo "Cleaning data..."
	@rm -f data/*.csv

help:
	@echo "Available commands:"
	@echo "  make build         - Build the scraper binary"
	@echo "  make run-scraper   - Run the scraper"
	@echo "  make clean         - Remove the binary"
	@echo "  make clean-data    - Remove the news data CSV file"
	@echo "  make help          - Show available commands"

.PHONY: all build run-scraper clean clean-data help
