# Docker Setup for Windows

## Prerequisites
- Docker Desktop for Windows installed
- WSL2 enabled (recommended)

## Setup Instructions

### 1. Start Docker Desktop
Make sure Docker Desktop is running before executing commands.

### 2. Setup Redis Stack

**PowerShell/CMD:**
```powershell
docker run -d -p 6379:6379 --name redis-stack redis/redis-stack:latest
```

**Verify Redis is running:**
```powershell
docker ps
```

### 3. Setup PostgreSQL

**PowerShell/CMD:**
```powershell
docker run -d -p 5432:5432 --name postgres -e POSTGRES_PASSWORD=crawler123 -e POSTGRES_DB=crawler postgres:latest
```

**Verify PostgreSQL is running:**
```powershell
docker ps
```

### 4. Initialize Database Schema

The schema will be automatically initialized when you run the application for the first time.

## Verify Setup

Check both containers are running:
```powershell
docker ps
```

You should see both `redis-stack` and `postgres` containers.

## Stop Containers

```powershell
docker stop redis-stack postgres
```

## Start Containers (after stopping)

```powershell
docker start redis-stack postgres
```

## Remove Containers (clean slate)

```powershell
docker rm -f redis-stack postgres
```

## Troubleshooting

### Port Already in Use
If ports 5432 or 6379 are already in use:

**Check what's using the port:**
```powershell
netstat -ano | findstr :5432
netstat -ano | findstr :6379
```

**Use different ports:**
```powershell
# Redis on port 6380
docker run -d -p 6380:6379 --name redis-stack redis/redis-stack:latest

# PostgreSQL on port 5433
docker run -d -p 5433:5432 --name postgres -e POSTGRES_PASSWORD=crawler123 -e POSTGRES_DB=crawler postgres:latest
```

Then update connection strings in `main.go` accordingly.

### Docker Desktop Not Starting
- Ensure WSL2 is installed and enabled
- Check Windows Hyper-V is enabled
- Restart Docker Desktop

## Run the Crawler

After containers are running:
```powershell
go run main.go
```

Server will start on `http://localhost:8080`
