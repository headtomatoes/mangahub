#!/bin/bash

# MangaDex Scraping and Import Script
# This script scrapes data from MangaDex API and imports it to the database

set -e

echo "================================================"
echo "MangaDex Data Scraping and Import"
echo "================================================"
echo ""

# Check if .env file exists
if [ -f "../.env" ]; then
    echo "Loading environment variables from .env..."
    export $(cat ../.env | grep -v '^#' | xargs)
else
    echo "Warning: .env file not found, using default values"
fi

# Step 1: Run the scraper
echo ""
echo "Step 1: Scraping data from MangaDex API..."
echo "-------------------------------------------"
cd "$(dirname "$0")"

if [ ! -f "mangadex_scraper.go" ]; then
    echo "Error: mangadex_scraper.go not found!"
    exit 1
fi

go run mangadex_scraper.go

if [ $? -ne 0 ]; then
    echo "Error: Scraping failed!"
    exit 1
fi

# Check if scraped data exists
if [ ! -f "scraped_data.json" ]; then
    echo "Error: scraped_data.json not found!"
    exit 1
fi

echo ""
echo "✓ Scraping completed successfully!"

# Step 2: Import to database
echo ""
echo "Step 2: Importing data to database..."
echo "--------------------------------------"

if [ ! -f "import_to_db.go" ]; then
    echo "Error: import_to_db.go not found!"
    exit 1
fi

go run import_to_db.go scraped_data.json

if [ $? -ne 0 ]; then
    echo "Error: Database import failed!"
    exit 1
fi

echo ""
echo "✓ Database import completed successfully!"

echo ""
echo "================================================"
echo "All operations completed successfully!"
echo "================================================"
echo ""
echo "Summary:"
echo "  - Data scraped from MangaDex API"
echo "  - Data saved to scraped_data.json"
echo "  - Data imported to database tables:"
echo "    * manga"
echo "    * genres"
echo "    * manga_genres"
echo ""
