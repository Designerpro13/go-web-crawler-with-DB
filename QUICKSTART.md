## Quick Start Guide - Optimized Crawler

### Prerequisites
1. PostgreSQL running on localhost:5432
2. Redis Stack running on localhost:6379

### Step 1: Setup PostgreSQL
```bash
# Login to PostgreSQL
psql -U postgres

# Run setup script
\i setup.sql

# Verify
\c crawler
\dt
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

### Step 3: Configure Database Connection
Edit main_v2.go line 186 with your PostgreSQL password:
```go
postgresConnStr := "host=localhost port=5432 user=postgres password=YOUR_PASSWORD dbname=crawler sslmode=disable"
```

### Step 4: Run
```bash
go run main_v2.go
```

### Step 5: Monitor Logs
```bash
# Watch all logged operations
go run main_v2.go 2>&1 | grep "LOGGED HERE"
```

### Step 6: Query Results
```sql
-- Connect to database
psql -U postgres -d crawler

-- View top words
SELECT word, SUM(frequency) as total FROM word_index GROUP BY word ORDER BY total DESC LIMIT 10;

-- View by domain
SELECT domain, COUNT(*) as word_count FROM word_index GROUP BY domain;
```

### Architecture Benefits
- ✅ No BadgerDB overhead
- ✅ Automatic UPSERT in PostgreSQL
- ✅ Fast Redis inverted index
- ✅ Batch processing (100 records)
- ✅ Domain-level aggregation
- ✅ Comprehensive logging
