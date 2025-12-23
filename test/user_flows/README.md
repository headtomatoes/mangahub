# MangaHub CLI User Workflow Scripts

This directory contains automated user workflow scripts that demonstrate different usage patterns of the MangaHub CLI.

## Overview

These PowerShell scripts automate common user workflows, making it easy to:
- Test CLI functionality end-to-end
- Demonstrate typical usage patterns
- Onboard new users with examples
- Verify system behavior across different scenarios

## Available Workflows

### 1. New User Registration Flow (`new_user_flow.ps1`)

**Purpose:** Complete onboarding for a brand new user

**Steps:**
1. Register new user account
2. Login with credentials
3. Browse manga catalog
4. Search for specific manga
5. View manga details
6. Add manga to personal library
7. Rate manga
8. Post comments
9. Track reading progress

**Usage:**
```powershell
.\new_user_flow.ps1
```

**Expected Outcome:** New user account created with library, ratings, comments, and progress tracking set up.

---

### 2. Admin Workflow (`admin_workflow.ps1`)

**Purpose:** Demonstrate admin-only operations

**Steps:**
1. Login as admin
2. Create new manga
3. Update manga details
4. Create new genre
5. List all genres
6. View manga by genre
7. List all manga
8. Delete test manga

**Usage:**
```powershell
.\admin_workflow.ps1
```

**Prerequisites:** Admin credentials (username: admin2, password: SecurePass123!)

**Expected Outcome:** Admin successfully performs CRUD operations on manga and genres.

---

### 3. Reader Workflow (`reader_workflow.ps1`)

**Purpose:** Typical activities of an active manga reader

**Steps:**
1. Login
2. Browse new releases
3. Search for favorite manga
4. View manga details
5. Add manga to library
6. View personal library
7. Rate multiple manga
8. Check average ratings
9. View other readers' ratings
10. Post comments
11. Read comments from community
12. Update reading progress

**Usage:**
```powershell
.\reader_workflow.ps1
```

**Expected Outcome:** Reader interacts with catalog, rates, comments, and tracks progress for multiple manga.

---

### 4. TCP Sync Workflow (`tcp_sync_workflow.ps1`)

**Purpose:** Demonstrate persistent TCP connection and real-time sync

**Steps:**
1. Login
2. Check TCP status (before connection)
3. Connect to TCP sync server
4. Verify connection status
5. Sync progress for multiple manga
6. Check sync status
7. Simulate CLI restart (daemon persists)
8. Verify connection still active
9. Sync more progress
10. Check uptime increased
11. Start monitor
12. Check final status
13. Disconnect gracefully
14. Verify disconnection

**Usage:**
```powershell
.\tcp_sync_workflow.ps1
```

**Expected Outcome:** TCP daemon runs persistently in background, state survives CLI restarts, uptime tracking works correctly.

---

### 5. Multi-Protocol Workflow (`multi_protocol_workflow.ps1`)

**Purpose:** Demonstrate all communication protocols (HTTP, gRPC, TCP, UDP)

**Steps:**
1. Login via HTTP API
2. List manga via HTTP
3. Get manga via gRPC
4. Search manga via gRPC
5. Test UDP connection
6. Listen for UDP notifications
7. Connect to TCP server
8. Update progress via HTTP
9. Sync progress via TCP
10. Update progress via gRPC
11. Check TCP status
12. Disconnect from TCP

**Usage:**
```powershell
.\multi_protocol_workflow.ps1
```

**Expected Outcome:** All four protocols (HTTP API, gRPC, TCP, UDP) work correctly and can be used interchangeably.

---

### 6. Master Runner (`run_all_workflows.ps1`)

**Purpose:** Execute all workflows in sequence

**Usage:**
```powershell
.\run_all_workflows.ps1
```

**What it does:**
- Runs all 5 workflow scripts in order
- Captures success/failure for each
- Displays execution summary
- Waits between workflows to avoid conflicts

**Expected Outcome:** All workflows complete successfully, comprehensive system test passed.

---

## Prerequisites

### Required Services
- **API Server**: Running on `localhost:8084`
- **gRPC Server**: Running on `localhost:8083`
- **TCP Server**: Running on `localhost:8081`
- **UDP Server**: Running on `localhost:8082`
- **PostgreSQL Database**: Running in Docker (`mangahub_db` container)

### User Accounts
- **Admin account**: username=`admin2`, password=`SecurePass123!`
- **New users**: Created automatically by `new_user_flow.ps1`

### CLI Binary
- Path: `c:\Project\Go\mangahub\cmd\cli\mangahub.exe`
- Must be compiled and up-to-date

## Running the Scripts

### Individual Workflow
```powershell
# Navigate to the directory
cd c:\Project\Go\mangahub\test\user_flows

# Run specific workflow
.\tcp_sync_workflow.ps1
```

### All Workflows
```powershell
cd c:\Project\Go\mangahub\test\user_flows
.\run_all_workflows.ps1
```

### From Another Directory
```powershell
powershell.exe -ExecutionPolicy Bypass -File "c:\Project\Go\mangahub\test\user_flows\reader_workflow.ps1"
```

## Troubleshooting

### Common Issues

**1. Execution Policy Error**
```
Solution: Run PowerShell as Administrator and execute:
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**2. Service Not Running**
```
Error: Connection refused / timeout
Solution: Start the required service (API, gRPC, TCP, or UDP server)
```

**3. Authentication Failed**
```
Error: 401 Unauthorized
Solution: Verify admin2 credentials or recreate user
```

**4. CLI Binary Not Found**
```
Error: Cannot find path
Solution: Update $CliPath in scripts or rebuild CLI:
cd c:\Project\Go\mangahub\cmd\cli
go build -o mangahub.exe
```

**5. TCP Daemon Already Running**
```
Error: Connection already established
Solution: Disconnect first:
.\mangahub sync disconnect
```

## Output and Logs

- Scripts output to console with color-coded status
- Successful operations: **Green**
- Warnings: **Yellow**
- Errors: **Red**
- Section headers: **Cyan/Magenta**

## Customization

### Modify User Credentials
Edit the script variables at the top:
```powershell
$username = "your_username"
$password = "your_password"
```

### Change CLI Path
Update the `$CliPath` variable:
```powershell
$CliPath = "your\custom\path\mangahub.exe"
```

### Adjust Timing
Modify `Start-Sleep` values for faster/slower execution:
```powershell
Start-Sleep -Seconds 3  # Change duration
```

## Testing and Validation

These scripts are designed for:
- **Integration testing**: Verify all CLI commands work end-to-end
- **Regression testing**: Ensure new changes don't break existing functionality
- **Demo purposes**: Show potential users how to use the CLI
- **Onboarding**: Help new developers understand CLI workflow
- **CI/CD**: Can be integrated into automated test pipelines

## Related Documentation

- [CLI Command Reference](../docs/cli_docs/CLI command.txt)
- [TCP Daemon Architecture](../docs/cli_docs/TCP-DAEMON-ARCHITECTURE.md)
- [Comprehensive Test Results](../docs/cli_docs/CLI-COMPREHENSIVE-TEST-RESULTS.md)
- [CLI Testing Guide](../docs/CLI-TESTING-GUIDE.md)

## Maintenance

### Updating Scripts
When CLI commands change:
1. Update affected workflow scripts
2. Test individually
3. Run `run_all_workflows.ps1` to verify
4. Update this README if new workflows are added

### Adding New Workflows
1. Create new `.ps1` file in this directory
2. Follow existing script structure
3. Add to `run_all_workflows.ps1`
4. Document in this README

## Support

For issues or questions:
- Check [CLI-TESTING-GUIDE.md](../docs/CLI-TESTING-GUIDE.md)
- Review [TCP-DAEMON-ARCHITECTURE.md](../docs/cli_docs/TCP-DAEMON-ARCHITECTURE.md)
- Examine [CLI-COMPREHENSIVE-TEST-RESULTS.md](../docs/cli_docs/CLI-COMPREHENSIVE-TEST-RESULTS.md)

---

**Last Updated:** 2025-12-23
**Version:** 1.0
**Maintainer:** MangaHub Development Team
