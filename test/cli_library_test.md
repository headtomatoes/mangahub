# CLI Library Test Cases

## Prerequisites
- API server running on `http://localhost:8084`
- PostgreSQL database running with migrations applied
- CLI binary built: `go build -o mangahub cmd/cli/main.go`

## Test Data Setup
Before running tests, ensure you have:
1. A test user account (username: `testuser`, password: `Test123!`, email: `test@example.com`)
2. At least 3 manga entries in the database (IDs: 1, 2, 3)

---

## Test Suite 1: Authentication (Login Required for Library Operations)

### Test Case 1.1: Register New User
**Command:**
```bash
./mangahub auth register -u testuser -p Test123! -e test@example.com
```

**Expected Output:**
```
✓ Registration successful! Please login to continue.
Username: testuser
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 1.2: Login with Valid Credentials
**Command:**
```bash
./mangahub auth login -u testuser -p Test123!
```

**Expected Output:**
```
✓ Successfully logged in!
```

**Status:** ✅ PASS / ❌ FAIL

**Notes:** Tokens are stored in system keyring for subsequent commands.

---

### Test Case 1.3: Login with Invalid Credentials
**Command:**
```bash
./mangahub auth login -u testuser -p WrongPassword
```

**Expected Output:**
```
Error: login process failed: login failed with status: 401 Unauthorized
```

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 2: Library Add Command

### Test Case 2.1: Add Manga to Library (Not Logged In)
**Precondition:** Ensure logged out
```bash
./mangahub auth logout
```

**Command:**
```bash
./mangahub library add 1
```

**Expected Output:**
```
Not logged in. Please login first.
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 2.2: Add Manga to Library (Valid Manga ID)
**Precondition:** User must be logged in
```bash
./mangahub auth login -u testuser -p Test123!
```

**Command:**
```bash
./mangahub library add 1
```

**Expected Output:**
```
✅ Successfully added manga (ID: 1) to your library
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 2.3: Add Same Manga Twice (Duplicate)
**Command:**
```bash
./mangahub library add 1
```

**Expected Output:**
```
Failed to add manga to library: manga already in library
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 2.4: Add Manga with Invalid ID Format
**Command:**
```bash
./mangahub library add abc
```

**Expected Output:**
```
Invalid manga ID: strconv.ParseInt: parsing "abc": invalid syntax
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 2.5: Add Non-Existent Manga ID
**Command:**
```bash
./mangahub library add 99999
```

**Expected Output:**
```
Failed to add manga to library: failed to add to library: 404 Not Found
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 2.6: Add Multiple Different Manga
**Command:**
```bash
./mangahub library add 2
./mangahub library add 3
```

**Expected Output:**
```
✅ Successfully added manga (ID: 2) to your library
✅ Successfully added manga (ID: 3) to your library
```

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 3: Library List Command

### Test Case 3.1: List Library (Not Logged In)
**Precondition:** Ensure logged out
```bash
./mangahub auth logout
```

**Command:**
```bash
./mangahub library list
```

**Expected Output:**
```
Not logged in. Please login first.
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 3.2: List Empty Library
**Precondition:** 
1. Login
2. Remove all manga from library if any

**Command:**
```bash
./mangahub auth login -u testuser -p Test123!
./mangahub library list
```

**Expected Output:**
```
📚 Your library is empty
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 3.3: List Library with Single Manga
**Precondition:** 
```bash
./mangahub library add 1
```

**Command:**
```bash
./mangahub library list
```

**Expected Output:**
```
📚 Your Library (1 manga)
─────────────────────────────────────────────────────────
1. One Piece (ID: 1)
   Author: Eiichiro Oda
   Status: Ongoing
   Added: 2025-11-18 14:30

```

**Status:** ✅ PASS / ❌ FAIL

**Notes:** The exact manga details will vary based on your database content.

---

### Test Case 3.4: List Library with Multiple Manga
**Precondition:**
```bash
./mangahub library add 1
./mangahub library add 2
./mangahub library add 3
```

**Command:**
```bash
./mangahub library list
```

**Expected Output:**
```
📚 Your Library (3 manga)
─────────────────────────────────────────────────────────
1. One Piece (ID: 1)
   Author: Eiichiro Oda
   Status: Ongoing
   Added: 2025-11-18 14:30

2. Naruto (ID: 2)
   Author: Masashi Kishimoto
   Status: Completed
   Added: 2025-11-18 14:31

3. Bleach (ID: 3)
   Author: Tite Kubo
   Status: Completed
   Added: 2025-11-18 14:32

```

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 4: Library Remove Command

### Test Case 4.1: Remove Manga (Not Logged In)
**Precondition:** Ensure logged out
```bash
./mangahub auth logout
```

**Command:**
```bash
./mangahub library remove 1
```

**Expected Output:**
```
Not logged in. Please login first.
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 4.2: Remove Manga with Invalid ID Format
**Command:**
```bash
./mangahub auth login -u testuser -p Test123!
./mangahub library remove xyz
```

**Expected Output:**
```
Invalid manga ID: strconv.ParseInt: parsing "xyz": invalid syntax
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 4.3: Remove Manga Not in Library
**Command:**
```bash
./mangahub library remove 99
```

**Expected Output:**
```
✅ Successfully removed manga (ID: 99) from your library
```
**OR**
```
Failed to remove manga from library: failed to remove from library: 404 Not Found
```

**Status:** ✅ PASS / ❌ FAIL

**Notes:** Behavior depends on API implementation.

---

### Test Case 4.4: Remove Valid Manga from Library
**Precondition:**
```bash
./mangahub library add 1
```

**Command:**
```bash
./mangahub library remove 1
```

**Expected Output:**
```
✅ Successfully removed manga (ID: 1) from your library
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 4.5: Verify Removal by Listing
**Precondition:** After removing manga with ID 1

**Command:**
```bash
./mangahub library list
```

**Expected Output:**
Library should not contain manga with ID 1.

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 5: Token Expiration & Refresh

### Test Case 5.1: Auto-Login with Valid Token
**Precondition:** Already logged in

**Command:**
```bash
./mangahub auth autologin
```

**Expected Output:**
```
(No output if token is still valid)
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 5.2: Library Operations with Expired Token
**Precondition:** Wait for token to expire or manually invalidate

**Command:**
```bash
./mangahub library list
```

**Expected Output:**
```
Failed to fetch library: failed to get library: 401 Unauthorized
```

**Status:** ✅ PASS / ❌ FAIL

**Notes:** User should login again.

---

## Test Suite 6: Logout and Cleanup

### Test Case 6.1: Logout
**Command:**
```bash
./mangahub auth logout
```

**Expected Output:**
```
✓ Successfully logged out.
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 6.2: Verify Logout - Cannot Access Library
**Command:**
```bash
./mangahub library list
```

**Expected Output:**
```
Not logged in. Please login first.
```

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 7: Complete Workflow Test

### Test Case 7.1: End-to-End Library Management
**Complete Workflow:**

```bash
# 1. Register
./mangahub auth register -u workflowuser -p Test123! -e workflow@example.com

# 2. Login
./mangahub auth login -u workflowuser -p Test123!

# 3. Check empty library
./mangahub library list

# 4. Add multiple manga
./mangahub library add 1
./mangahub library add 2
./mangahub library add 3

# 5. List library
./mangahub library list

# 6. Remove one manga
./mangahub library remove 2

# 7. Verify removal
./mangahub library list

# 8. Logout
./mangahub auth logout

# 9. Verify cannot access after logout
./mangahub library list
```

**Expected:** All commands should work as expected in sequence.

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 8: Error Handling & Edge Cases

### Test Case 8.1: Missing Argument for Add Command
**Command:**
```bash
./mangahub library add
```

**Expected Output:**
```
Error: accepts 1 arg(s), received 0
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 8.2: Missing Argument for Remove Command
**Command:**
```bash
./mangahub library remove
```

**Expected Output:**
```
Error: accepts 1 arg(s), received 0
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 8.3: Help Command for Library
**Command:**
```bash
./mangahub library --help
```

**Expected Output:**
```
Add, remove, and list manga in your personal library

Usage:
  mangahubCLI library [command]

Available Commands:
  add         Add a manga to your library
  list        List all manga in your library
  remove      Remove a manga from your library

Flags:
  -h, --help   help for library

Use "mangahubCLI library [command] --help" for more information about a command.
```

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 8.4: Help for Specific Library Command
**Command:**
```bash
./mangahub library add --help
```

**Expected Output:**
```
Add a manga to your library

Usage:
  mangahubCLI library add [manga_id] [flags]

Flags:
  -h, --help   help for add
```

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 9: API Server Down Scenarios

### Test Case 9.1: Library Operations When API Server is Down
**Precondition:** Stop the API server

**Command:**
```bash
./mangahub library list
```

**Expected Output:**
```
Failed to fetch library: Get "http://localhost:8084/api/library": dial tcp [::1]:8084: connect: connection refused
```

**Status:** ✅ PASS / ❌ FAIL

---

## Test Suite 10: Custom API URL

### Test Case 10.1: Use Custom API URL via Flag
**Command:**
```bash
./mangahub --api-url https://api.mangahub.example.com auth login -u testuser -p Test123!
```

**Expected Output:**
Should attempt to connect to the specified URL.

**Status:** ✅ PASS / ❌ FAIL

---

### Test Case 10.2: Use Custom API URL via Environment Variable
**Command:**
```bash
export MANGAHUB_API_URL=https://api.mangahub.example.com
./mangahub auth login -u testuser -p Test123!
```

**Expected Output:**
Should attempt to connect to the URL from environment variable.

**Status:** ✅ PASS / ❌ FAIL

---

## Test Execution Summary

| Test Suite | Total Tests | Passed | Failed | Notes |
|------------|-------------|--------|--------|-------|
| Authentication | 3 | | | |
| Library Add | 6 | | | |
| Library List | 4 | | | |
| Library Remove | 5 | | | |
| Token Management | 2 | | | |
| Logout | 2 | | | |
| Complete Workflow | 1 | | | |
| Error Handling | 4 | | | |
| API Down | 1 | | | |
| Custom URL | 2 | | | |
| **TOTAL** | **30** | | | |

---

## Notes for Testers

1. **Database State**: Ensure the database is in a clean state before running tests.
2. **Keyring**: Some systems may require keyring setup (e.g., gnome-keyring on Linux).
3. **Ports**: Ensure port 8084 is not in use by other services.
4. **Time Zones**: Timestamps in output may vary based on system timezone.
5. **Response Times**: Network latency may affect command execution time.

---

## Automated Test Script

You can create a bash script to automate these tests:

```bash
#!/bin/bash

# automated_library_test.sh
# Run this script to test all library CLI commands

set -e  # Exit on error

API_URL="http://localhost:8084"
CLI_BIN="./mangahub"
TEST_USER="testuser_$(date +%s)"
TEST_PASS="Test123!"
TEST_EMAIL="${TEST_USER}@example.com"

echo "=== MangaHub CLI Library Test Suite ==="
echo ""

# Test 1: Register
echo "Test 1: Registering user..."
$CLI_BIN auth register -u "$TEST_USER" -p "$TEST_PASS" -e "$TEST_EMAIL"
echo "✅ Registration test passed"
echo ""

# Test 2: Login
echo "Test 2: Logging in..."
$CLI_BIN auth login -u "$TEST_USER" -p "$TEST_PASS"
echo "✅ Login test passed"
echo ""

# Test 3: List empty library
echo "Test 3: Listing empty library..."
$CLI_BIN library list
echo "✅ Empty library test passed"
echo ""

# Test 4: Add manga
echo "Test 4: Adding manga to library..."
$CLI_BIN library add 1
$CLI_BIN library add 2
echo "✅ Add manga test passed"
echo ""

# Test 5: List library with items
echo "Test 5: Listing library with items..."
$CLI_BIN library list
echo "✅ List library test passed"
echo ""

# Test 6: Remove manga
echo "Test 6: Removing manga from library..."
$CLI_BIN library remove 1
echo "✅ Remove manga test passed"
echo ""

# Test 7: Verify removal
echo "Test 7: Verifying removal..."
$CLI_BIN library list
echo "✅ Verification test passed"
echo ""

# Test 8: Logout
echo "Test 8: Logging out..."
$CLI_BIN auth logout
echo "✅ Logout test passed"
echo ""

echo "=== All Tests Completed Successfully ==="
```

Save this as `test/automated_library_test.sh` and run with:
```bash
chmod +x test/automated_library_test.sh
./test/automated_library_test.sh
```
