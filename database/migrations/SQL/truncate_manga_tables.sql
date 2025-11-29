-- ========================================
-- TRUNCATE MANGA, GENRES, AND MANGA_GENRES TABLES
-- ========================================
-- This script removes all data from manga-related tables
-- and resets their primary key sequences to 1
-- Run this BEFORE scraping data from MangaDex
-- ========================================

-- Disable triggers temporarily to avoid cascading issues
SET session_replication_role = 'replica';

-- Truncate tables in correct order (child tables first)
-- CASCADE will automatically truncate dependent tables
TRUNCATE TABLE manga_genres CASCADE;
TRUNCATE TABLE manga CASCADE;
TRUNCATE TABLE genres CASCADE;

-- Re-enable triggers
SET session_replication_role = 'origin';

-- Reset sequences to start from 1
ALTER SEQUENCE manga_id_seq RESTART WITH 1;
ALTER SEQUENCE genres_id_seq RESTART WITH 1;
ALTER SEQUENCE manga_genres_id_seq RESTART WITH 1;

-- Verify truncation
SELECT 'manga' AS table_name, COUNT(*) AS row_count FROM manga
UNION ALL
SELECT 'genres', COUNT(*) FROM genres
UNION ALL
SELECT 'manga_genres', COUNT(*) FROM manga_genres;

-- Verify sequence reset
SELECT 'manga_id_seq' AS sequence_name, last_value FROM manga_id_seq
UNION ALL
SELECT 'genres_id_seq', last_value FROM genres_id_seq
UNION ALL
SELECT 'manga_genres_id_seq', last_value FROM manga_genres_id_seq;

-- Success message
SELECT 'Tables truncated successfully! Ready to scrape MangaDex data.' AS status;
