# Docker Setup Commands

## Run these commands in order:

### 1. Clean up existing containers
```bash
docker stop redis-stack postgres-crawler 2>/dev/null
docker rm redis-stack postgres-crawler 2>/dev/null
sudo fuser -k 6379/tcp 5432/tcp 2>/dev/null
```

### 2. Start Redis
```bash
docker run -d --name redis-stack -p 6379:6379 redis/redis-stack:latest
```

### 3. Start PostgreSQL
```bash
docker run -d --name postgres-crawler -e POSTGRES_PASSWORD=crawler123 -e POSTGRES_DB=crawler -p 5432:5432 postgres:15-alpine
```

### 4. Wait and create schema
```bash
sleep 5
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
SELECT tablename FROM pg_tables WHERE schemaname='public';
EOF
```

### 5. Verify
```bash
docker ps
redis-cli ping
docker exec -it postgres-crawler psql -U postgres -d crawler -c "\dt"
```

### 6. Run crawler
```bash
go run main_v2.go
```

## Credentials
- PostgreSQL: user=postgres, password=crawler123, db=crawler
- Redis: localhost:6379 (no password)
