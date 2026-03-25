# Optimized Web Crawler v2

PostgreSQL + Redis Inverted Index implementation for efficient web crawling and word indexing.

## Architecture

- **PostgreSQL**: Persistent storage for word frequencies by domain with automatic UPSERT
- **Redis Bloom Filter**: Duplicate URL detection
- **Redis Inverted Index**: Fast in-memory word-to-URL mappings
- **Batch Processing**: Efficient bulk operations (100 records per batch)

## Setup

### 1. PostgreSQL Setup
```bash
# Install PostgreSQL
sudo apt-get install postgresql postgresql-contrib

# Start PostgreSQL
sudo service postgresql start

# Create database and schema
psql -U postgres -f setup.sql
```

### 2. Redis Stack Setup
```bash
# Pull Redis Stack
docker pull redis/redis-stack:latest

# Run Redis Stack
docker run -d -p 6379:6379 --name redis-stack redis/redis-stack:latest

# Verify
docker ps
```

### 3. Install Dependencies
```bash
go get github.com/lib/pq
go get github.com/redis/go-redis/v9
go get golang.org/x/net/html
```

### 4. Configure PostgreSQL Connection
Edit `main_v2.go` line 186:
```go
postgresConnStr := "host=localhost port=5432 user=postgres password=YOUR_PASSWORD dbname=crawler sslmode=disable"
```

## Run

```bash
go run main_v2.go
```

## Data Flow

```
Seed URLs → Queue → Workers → Fetch HTML → Extract Text
                                              ↓
                                         Word Frequency
                                              ↓
                                    ┌─────────┴─────────┐
                                    ↓                   ↓
                          Redis Inverted Index    Batch Buffer
                          (word→[url:freq])            ↓
                                              PostgreSQL UPSERT
                                              (word, domain, freq)
```

## Query Examples

```sql
-- Top 20 most frequent words
SELECT word, SUM(frequency) as total_freq 
FROM word_index 
GROUP BY word 
ORDER BY total_freq DESC 
LIMIT 20;

-- Top domains by unique words
SELECT domain, COUNT(DISTINCT word) as unique_words 
FROM word_index 
GROUP BY domain 
ORDER BY unique_words DESC 
LIMIT 10;

-- Words appearing across multiple domains
SELECT word, COUNT(DISTINCT domain) as domain_count 
FROM word_index 
GROUP BY word 
HAVING COUNT(DISTINCT domain) > 1 
ORDER BY domain_count DESC 
LIMIT 20;
```

## Logging

All critical operations are marked with `LOGGED HERE:` for easy tracking:
- Connection status
- URL fetching
- Duplicate detection
- Word processing
- Batch operations
- Database writes

Search logs: `grep "LOGGED HERE" output.log`

## Features

- ✅ Automatic UPSERT (no duplicate handling needed)
- ✅ Batch processing for performance
- ✅ Domain-level aggregation
- ✅ Redis inverted index for fast lookups
- ✅ Graceful shutdown with data flush
- ✅ Comprehensive logging
- ✅ Connection pooling
- ✅ BFS crawling strategy
