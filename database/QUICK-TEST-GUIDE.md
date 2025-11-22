# Quick Database Testing Guide

**Quick reference for testing your PostgreSQL setup**

## Option 1: Using Docker CLI (Fastest)

### Basic Connection Test
```powershell
# Connect to PostgreSQL
docker exec -it mangahub_db psql -U mangahub -d mangahub

# Or single command to list tables:
docker exec -it mangahub_db psql -U mangahub -d mangahub -c "\dt"
```

### Essential Commands Inside psql

```sql
-- List all tables
\dt

-- Show table structure
\d users
\d refresh_tokens
\d manga

-- Check if role column exists (CRITICAL TEST)
SELECT column_name, data_type, column_default
FROM information_schema.columns
WHERE table_name = 'users'
ORDER BY ordinal_position;

-- Verify indexes
\di

-- Check foreign keys
\d+ users

-- Count records
SELECT 'users' as table, COUNT(*) FROM users
UNION ALL
SELECT 'manga', COUNT(*) FROM manga
UNION ALL
SELECT 'refresh_tokens', COUNT(*) FROM refresh_tokens;

-- Exit
\q
```

### Quick Health Check Script

```powershell
# Run this to verify everything
docker exec -it mangahub_db psql -U mangahub -d mangahub << 'EOF'
SELECT '=== DATABASE HEALTH CHECK ===' AS status;

-- Check tables exist
SELECT COUNT(*) || ' tables found (expected: 11)' AS table_count
FROM information_schema.tables 
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

-- Check role column exists
SELECT CASE 
    WHEN EXISTS (SELECT 1 FROM information_schema.columns 
                 WHERE table_name = 'users' AND column_name = 'role')
    THEN '✅ users.role column EXISTS'
    ELSE '❌ users.role column MISSING - Run migration 003'
END AS role_check;

-- Check token indexes
SELECT COUNT(*) || ' token indexes found (expected: 2)' AS index_count
FROM pg_indexes 
WHERE tablename = 'refresh_tokens' 
  AND indexname LIKE 'idx_refresh%';

-- List all tables
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;
EOF
```

## Option 2: Using pgAdmin 4

### Add pgAdmin to Your docker-compose.yml

```yaml
# Add this service to docker-compose.yml
pgadmin:
  image: dpage/pgadmin4:latest
  container_name: mangahub-pgadmin
  restart: always
  environment:
    PGADMIN_DEFAULT_EMAIL: admin@mangahub.local
    PGADMIN_DEFAULT_PASSWORD: admin123
    PGADMIN_CONFIG_SERVER_MODE: 'False'
  ports:
    - "5050:80"
  volumes:
    - pgadmin_data:/var/lib/pgadmin
  networks:
    - mangahub-network
  depends_on:
    db:
      condition: service_healthy

# Add to volumes section:
volumes:
  pgadmin_data:  # Add this line
```

### Start pgAdmin

```powershell
# Update docker-compose and start pgAdmin
docker compose up -d pgadmin

# Wait for it to start (check logs)
docker compose logs pgadmin -f

# Open in browser:
# http://localhost:5050
# Email: admin@mangahub.local
# Password: admin123
```

### Connect to Database in pgAdmin

1. **Right-click "Servers" → "Register" → "Server"**

2. **General Tab:**
   - Name: `MangaHub DB`

3. **Connection Tab:**
   - Host: `db` (not localhost!)
   - Port: `5432`
   - Database: `mangahub`
   - Username: `mangahub`
   - Password: `mangahub_secret` (or your DB_PASS from .env)

4. **Click Save**

### Useful Queries in pgAdmin

```sql
-- Check schema version
SELECT * FROM schema_migrations; -- If golang-migrate is set up

-- Find missing columns
SELECT 
    c.table_name,
    c.column_name,
    c.data_type
FROM information_schema.columns c
WHERE c.table_schema = 'public'
  AND c.table_name IN ('users', 'refresh_tokens', 'manga')
ORDER BY c.table_name, c.ordinal_position;

-- Check constraints
SELECT
    tc.constraint_name,
    tc.table_name,
    tc.constraint_type,
    cc.check_clause
FROM information_schema.table_constraints tc
LEFT JOIN information_schema.check_constraints cc 
    ON tc.constraint_name = cc.constraint_name
WHERE tc.table_schema = 'public'
ORDER BY tc.table_name, tc.constraint_type;

-- Performance: Check index usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;
```

## Option 3: Test with Your API

### Check API is running

```powershell
# Start all services
docker compose up -d

# Wait for health checks
Start-Sleep -Seconds 10

# Test connection endpoint
curl http://localhost:8084/check-conn

# Should return: {"message":"api server running","ok":true}
```

### Test database ping endpoint

```powershell
curl http://localhost:8084/db-ping

# Should return: {"ok":true}
# If error, check: docker compose logs api-server
```

### Test auth registration (full database test)

```powershell
# Register a test user
$body = @{
    username = "testuser"
    password = "password123"
    email = "test@example.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8084/auth/register `
    -Method POST `
    -ContentType "application/json" `
    -Body $body

# Should return user details with user_id
# If it fails with "role column" error, your migration didn't run
```

## Option 4: Run Integration Tests

### Setup test environment

```powershell
# Set environment variables
$env:DATABASE_URL = "postgres://mangahub:mangahub_secret@localhost:5432/mangahub?sslmode=disable"
$env:JWT_SECRET = "test-secret-key-must-be-at-least-32-chars-long-for-security"

# Run tests
go test -v ./test/auth_integration_test.go
```

### Expected Results

**If migrations are complete:**
```
=== RUN   TestAuthIntegrationTestSuite
=== RUN   TestAuthIntegrationTestSuite/TestRegister_Success
=== RUN   TestAuthIntegrationTestSuite/TestLogin_Success
...
--- PASS: TestAuthIntegrationTestSuite (2.34s)
PASS
```

**If role column is missing:**
```
ERROR: column "role" of relation "users" does not exist
```

**If test database isn't accessible:**
```
--- SKIP: TestAuthIntegrationTestSuite (0.00s)
    Failed to open test database (is Docker running?)
```

## Common Issues & Quick Fixes

### Issue: "role does not exist"
```sql
-- Check if role column exists
docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT column_name FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role';"

-- If empty, manually add it:
docker exec -it mangahub_db psql -U mangahub -d mangahub << 'EOF'
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(255) DEFAULT 'user';
ALTER TABLE users ALTER COLUMN role SET NOT NULL;
EOF
```

### Issue: "password authentication failed"
```powershell
# Check your password in docker-compose.yml
docker compose config | Select-String "POSTGRES_PASSWORD"

# Restart with fresh database
docker compose down -v
docker compose up -d
```

### Issue: "Connection refused"
```powershell
# Check if database is running
docker ps | Select-String mangahub_db

# Check database logs
docker compose logs db

# Test connection from host
Test-NetConnection -ComputerName localhost -Port 5432
```

### Issue: "Too many open connections"
```sql
-- Check current connections
docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'mangahub';"

-- See connection details
docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT pid, usename, application_name, client_addr, state FROM pg_stat_activity WHERE datname = 'mangahub';"

-- Kill idle connections (if needed)
docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'mangahub' AND state = 'idle';"
```

## Verify Migration Status

### Check what migrations have run

```powershell
# If using golang-migrate (after implementing it)
docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT * FROM schema_migrations ORDER BY version;"

# Check what files are mounted
docker exec -it mangahub_db ls -la /docker-entrypoint-initdb.d/

# Check PostgreSQL logs for initialization
docker compose logs db | Select-String "init"
```

### Manual migration check

```sql
-- Connect to database
docker exec -it mangahub_db psql -U mangahub -d mangahub

-- Check each migration's effect:

-- 001_init.up.sql - Should have created all tables
\dt

-- 003_add_user_role.up.sql - Should have added role column
SELECT column_name FROM information_schema.columns 
WHERE table_name = 'users' AND column_name = 'role';

-- 004_add_index_token.up.sql - Should have created indexes
SELECT indexname FROM pg_indexes 
WHERE tablename = 'refresh_tokens' 
  AND indexname LIKE 'idx_refresh%';
```

## Database Reset (Nuclear Option)

**WARNING: This deletes all data!**

```powershell
# Stop and remove containers with volumes
docker compose down -v

# Remove database data directory (if needed)
docker volume rm mangahub_postgres_data

# Restart fresh
docker compose up -d

# Wait for initialization
Start-Sleep -Seconds 15

# Verify tables created
docker exec -it mangahub_db psql -U mangahub -d mangahub -c "\dt"
```

## Performance Testing

### Check query performance

```sql
-- Enable query timing
\timing on

-- Test user lookup (should be fast with index)
SELECT * FROM users WHERE username = 'testuser';

-- Test token validation (should use index)
SELECT * FROM refresh_tokens 
WHERE token = 'some-token' AND revoked = false;

-- Explain query plan
EXPLAIN ANALYZE 
SELECT * FROM refresh_tokens 
WHERE revoked = false AND expires_at > NOW();
```

### Monitor active queries

```sql
-- See what's running
SELECT pid, age(clock_timestamp(), query_start), usename, query 
FROM pg_stat_activity 
WHERE query != '<IDLE>' AND query NOT ILIKE '%pg_stat_activity%'
ORDER BY query_start DESC;
```

## Next Steps

After verifying your database setup:

1. **Fix Critical Issues:**
   - ✅ Ensure role column exists
   - ✅ Verify all indexes are created
   - ✅ Confirm foreign key constraints work

2. **Implement golang-migrate:**
   - See DATABASE-SETUP-REVIEW.md Section 2

3. **Setup separate test database:**
   - See DATABASE-SETUP-REVIEW.md Section 4

4. **Add monitoring:**
   - Consider adding pg_stat_statements extension
   - Set up regular backups

---

**Pro Tip:** Add this health check script to your Makefile:

```makefile
## db-check: Quick database health check
db-check:
	@echo "🔍 Checking database health..."
	@docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT '✅ Database connected successfully' AS status;"
	@docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT COUNT(*) || ' tables' AS table_count FROM information_schema.tables WHERE table_schema = 'public';"
	@docker exec -it mangahub_db psql -U mangahub -d mangahub -c "SELECT CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role') THEN '✅ Role column exists' ELSE '❌ Role column MISSING' END AS role_status;"
```

Then run:
```powershell
make db-check
```
