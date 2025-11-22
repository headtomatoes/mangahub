-- ============================================
-- MangaHub Database Health Check
-- ============================================
-- Run with: docker exec -i mangahub_db psql -U mangahub -d mangahub < database/health-check.sql

\echo '================================='
\echo 'MANGAHUB DATABASE HEALTH CHECK'
\echo '================================='
\echo ''

-- Connection test
\echo '1. Connection Test'
SELECT '✅ Database connected successfully' AS status;
\echo ''

-- Table count check
\echo '2. Table Count'
SELECT 
    CASE 
        WHEN COUNT(*) = 11 THEN '✅ All 11 tables exist'
        ELSE '❌ Missing tables: expected 11, found ' || COUNT(*)::text
    END AS table_check
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_type = 'BASE TABLE';
\echo ''

-- List all tables
\echo '3. Tables Found:'
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_type = 'BASE TABLE'
ORDER BY table_name;
\echo ''

-- Check role column (CRITICAL)
\echo '4. Critical Column Check'
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'users' AND column_name = 'role'
        ) THEN '✅ users.role column exists'
        ELSE '❌ users.role column MISSING - Run migration 003!'
    END AS role_check;
\echo ''

-- Check token indexes
\echo '5. Index Check'
SELECT 
    CASE 
        WHEN COUNT(*) >= 2 THEN '✅ Token indexes exist (' || COUNT(*)::text || ' found)'
        ELSE '❌ Token indexes missing: ' || COUNT(*)::text || ' of 2'
    END AS index_check
FROM pg_indexes 
WHERE tablename = 'refresh_tokens' 
  AND indexname IN ('idx_refresh_tokens_revoked', 'idx_refresh_tokens_expires_at');
\echo ''

-- Foreign key constraints
\echo '6. Foreign Key Constraints'
SELECT 
    COUNT(*) || ' foreign key constraints found' AS fk_count
FROM information_schema.table_constraints 
WHERE constraint_type = 'FOREIGN KEY' 
  AND table_schema = 'public';
\echo ''

-- Check constraints
\echo '7. Check Constraints'
SELECT 
    table_name,
    constraint_name,
    check_clause
FROM information_schema.table_constraints tc
JOIN information_schema.check_constraints cc
  ON tc.constraint_name = cc.constraint_name
WHERE tc.table_schema = 'public'
ORDER BY table_name;
\echo ''

-- Index list
\echo '8. Important Indexes'
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE schemaname = 'public'
  AND (
    indexname LIKE 'idx_%'
    OR indexname LIKE '%_pkey'
  )
ORDER BY tablename, indexname;
\echo ''

-- Database size
\echo '9. Database Size'
SELECT 
    pg_size_pretty(pg_database_size(current_database())) AS database_size;
\echo ''

-- Record counts
\echo '10. Record Counts'
SELECT 
    (SELECT COUNT(*) FROM users) AS users,
    (SELECT COUNT(*) FROM manga) AS manga,
    (SELECT COUNT(*) FROM genres) AS genres,
    (SELECT COUNT(*) FROM refresh_tokens) AS tokens,
    (SELECT COUNT(*) FROM user_progress) AS progress,
    (SELECT COUNT(*) FROM notifications) AS notifications;
\echo ''

-- Connection stats
\echo '11. Active Connections'
SELECT 
    COUNT(*) AS active_connections,
    COUNT(*) FILTER (WHERE state = 'active') AS active_queries,
    COUNT(*) FILTER (WHERE state = 'idle') AS idle_connections
FROM pg_stat_activity 
WHERE datname = current_database();
\echo ''

\echo '================================='
\echo 'HEALTH CHECK COMPLETE'
\echo '================================='
