# MangaHub Development Commands & File Structure Guide

I've created a comprehensive development guide that explains every file in your project and all the commands you'll need. Here's a summary of the key sections:

## Quick Reference: Most Important Commands

### **Daily Development**

```bash
# Start everything (most common command)
make dev

# Start only API server  
make dev-api

# View logs while developing
make logs-api
```

### **File Structure Explanation**

**Configuration Files**:

- `Dockerfile` → Instructions to build your Go app container
- `docker-compose.yml` → Orchestrates all services (API, TCP, UDP, gRPC, Redis)
- `.air.*.toml` → Live reload configs for each service
- `.env` → Environment variables (ports, secrets, database paths)

**Source Code**:

- `cmd/*/main.go` → Entry points for each server
- `internal/` → Your private business logic (auth, manga, user management)
- `pkg/models/` → Shared data structures
- `proto/` → gRPC protocol definitions

### **What Happens When You Run Commands**

**`make dev`**:

1. Builds Docker images with Go + SQLite + Air
2. Starts 5 containers: API (8080), TCP (8081), UDP (8082), gRPC (8083), Redis (6379)
3. Mounts your code directory so changes appear instantly
4. Each service watches for file changes and auto-restarts

**When You Edit Code**:

1. Save any Go file
2. Air detects the change
3. Rebuilds only the affected service
4. Restarts that service
5. Your changes are live immediately

**`make logs-api`**:

- Shows real-time output from API server
- Helps debug issues and see request logs

### **Cross-Platform Benefits**

This setup ensures that whether you're on Windows, your teammate on Mac, or deploying on Linux:

- Same Go version (1.23)
- Same SQLite version
- Same dependencies
- Same file paths
- Same network ports
- Same environment variables

The Docker containers create a standardized Linux environment regardless of your host OS, eliminating "works on my machine" problems completely.

### **Development Workflow Example**

```bash
# Morning routine
git pull origin main
make dev                    # Start all services
# Edit internal/manga/handlers.go
# Air automatically rebuilds API server
# Test at http://localhost:8080/api/manga
make logs-api              # Check for errors
```

The guide covers everything from basic startup commands to advanced debugging, database management, and production deployment. Each command is explained with what it does and why you need it.
