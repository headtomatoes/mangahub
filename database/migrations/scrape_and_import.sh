#!/bin/bash

set -e

echo "========================================"
echo "MangaDex Scraper and Database Importer"
echo "========================================"
echo ""

# Step 1: Scrape data
echo "Step 1: Scraping data from MangaDex..."
echo "--------------------------------------"
cd Scrape
go run mangadex_scraper.go types.go
if [ $? -ne 0 ]; then
    echo "❌ Scraping failed!"
    exit 1
fi
cd ..
echo "✓ Scraping completed!"
echo ""

# Step 2: Import to database
echo "Step 2: Importing data to database..."
echo "--------------------------------------"
cd import
go run import_to_db.go types.go
if [ $? -ne 0 ]; then
    echo "❌ Import failed!"
    exit 1
fi
cd ..
echo "✓ Import completed!"
echo ""

echo "========================================"
echo "✓ Process completed successfully!"
echo "========================================"