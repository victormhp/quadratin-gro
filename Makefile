build-api:
	@echo "Building api..."
	@go build -o bin/api ./cmd/api

build-scraper:
	@echo "Building scraper..."
	@go build -o bin/scraper ./cmd/scraper

build-db:
	@echo "Building sqlite3 database..."
	@sqlite3 data/quadratin.db < data/schema.sql

run-api:
	@go run ./cmd/api

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
	@echo "  make build-api      - Build the api binary"
	@echo "  make build-scraper  - Build the scraper binary"
	@echo "  make build-db       - Build sqlite3 quadratin database"
	@echo "  make run-api        - Run the api"
	@echo "  make run-scraper    - Run the scraper"
	@echo "  make clean          - Remove binaries"
	@echo "  make clean-data     - Remove data CSV files"
	@echo "  make help           - Show available commands"

.PHONY: all build buld-api build-scraper build-db run-api run-scraper clean clean-data help
