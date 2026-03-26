# Optimized Web Crawler v2

PostgreSQL + Redis Inverted Index implementation for efficient web crawling and word indexing.

## Architecture

- **PostgreSQL**: Persistent storage for word frequencies by domain with automatic UPSERT
- **Redis Bloom Filter**: Duplicate URL detection
- **Redis Inverted Index**: Fast in-memory word-to-URL mappings
- **Batch Processing**: Efficient bulk operations (100 records per batch)

## Setup

### Quick Setup with Docker
```bash
# Run automated setup
chmod +x docker-setup.sh
./docker-setup.sh
```

### Manual Docker Setup

#### 1. Redis Stack
```bash
docker run -d --name redis-stack -p 6379:6379 redis/redis-stack:latest
redis-cli ping  # Should return PONG
```

#### 2. PostgreSQL
```bash
# Start PostgreSQL container
docker run -d \
  --name postgres-crawler \
  -e POSTGRES_PASSWORD=crawler123 \
  -e POSTGRES_DB=crawler \
  -p 5432:5432 \
  postgres:15-alpine

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

### 3. Install Go Dependencies
```bash
go get github.com/lib/pq
go get github.com/redis/go-redis/v9
go get golang.org/x/net/html
```

### 4. PostgreSQL Connection (Pre-configured)
Default in `main_v2.go`:
```go
postgresConnStr := "host=localhost port=5432 user=postgres password=crawler123 dbname=crawler sslmode=disable"
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
