## Quick Start Guide - Optimized Crawler

### Prerequisites
1. Docker installed
2. Go installed

### Step 1: Setup Docker Containers (Redis + PostgreSQL)
```bash
# Make script executable
chmod +x docker-setup.sh

# Run setup script
./docker-setup.sh
```

This will:
- Clean up existing containers
- Free up ports 6379 and 5432
- Start Redis Stack on port 6379
- Start PostgreSQL on port 5432 (user: postgres, password: crawler123)
- Create database 'crawler'
- Create word_index table with indexes

**Manual Setup (if script fails):**
```bash
# Stop existing
docker stop redis-stack postgres-crawler 2>/dev/null
docker rm redis-stack postgres-crawler 2>/dev/null
sudo fuser -k 6379/tcp 6379/tcp 2>/dev/null

# Redis
docker run -d --name redis-stack -p 6379:6379 redis/redis-stack:latest

# PostgreSQL
docker run -d --name postgres-crawler -e POSTGRES_PASSWORD=crawler123 -e POSTGRES_DB=crawler -p 5432:5432 postgres:15-alpine

# Wait for PostgreSQL
sleep 5

# Create schema
docker exec -i postgres-crawler psql -U postgres -d crawler << 'EOF'
CREATE TABLE IF NOT EXISTS word_index (
    id SERIAL PRIMARY KEY,
    word VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    frequency INTEGER DEFAULT 1,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(word, domain)
);
CREATE INDEX IF NOT EXISTS idx_word ON word_index(word);
CREATE INDEX IF NOT EXISTS idx_domain ON word_index(domain);
CREATE INDEX IF NOT EXISTS idx_frequency ON word_index(frequency DESC);
\dt
EOF
```

### Step 2: Update go.mod
Add this line to go.mod:
```
require github.com/lib/pq v1.10.9
```

Then run:
```bash
go mod tidy
```

### Step 3: Verify Containers
```bash
# Check running containers
docker ps

# Test Redis
redis-cli ping  # Should return PONG

# Test PostgreSQL
docker exec -it postgres-crawler psql -U postgres -d crawler -c "\dt"
```

### Step 4: Database Connection (Already Configured)
Default credentials in main_v2.go:
```go
postgresConnStr := "host=localhost port=5432 user=postgres password=crawler123 dbname=crawler sslmode=disable"
```

### Step 5: Run
```bash
go run main_v2.go
```

### Step 6: Monitor Logs
```bash
# Watch all logged operations
go run main_v2.go 2>&1 | grep "LOGGED HERE"
```

### Step 7: Query Results
```bash
# Connect to database
docker exec -it postgres-crawler psql -U postgres -d crawler
```

```sql
-- View top words
SELECT word, SUM(frequency) as total FROM word_index GROUP BY word ORDER BY total DESC LIMIT 10;

-- View by domain
SELECT domain, COUNT(*) as word_count FROM word_index GROUP BY domain;

-- Exit
\q
```

### Architecture Benefits
- ✅ No BadgerDB overhead
- ✅ Automatic UPSERT in PostgreSQL
- ✅ Fast Redis inverted index
- ✅ Batch processing (100 records)
- ✅ Domain-level aggregation
- ✅ Comprehensive logging
