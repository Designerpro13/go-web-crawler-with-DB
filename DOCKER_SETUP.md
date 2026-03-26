# Docker Setup Guide

## One-Command Setup

```bash
chmod +x docker-setup.sh && ./docker-setup.sh
```

## Manual Setup

### 1. Stop Existing Containers
```bash
docker stop redis-stack postgres-crawler 2>/dev/null
docker rm redis-stack postgres-crawler 2>/dev/null
```

### 2. Start Redis Stack
```bash
docker run -d \
  --name redis-stack \
  -p 6379:6379 \
  redis/redis-stack:latest

# Test
redis-cli ping
```

### 3. Start PostgreSQL
```bash
docker run -d \
  --name postgres-crawler \
  -e POSTGRES_PASSWORD=crawler123 \
  -e POSTGRES_DB=crawler \
  -p 5432:5432 \
  postgres:15-alpine
```

### 4. Create Schema
```bash
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

## Connection Details

**PostgreSQL:**
- Host: `localhost`
- Port: `5432`
- User: `postgres`
- Password: `crawler123`
- Database: `crawler`

**Redis:**
- Host: `localhost`
- Port: `6379`

## Verify Setup

```bash
# Check containers
docker ps

# Test Redis
redis-cli ping

# Test PostgreSQL
docker exec -it postgres-crawler psql -U postgres -d crawler -c "\dt"
```

## Troubleshooting

**Port already in use:**
```bash
# Find process using port
sudo lsof -i :6379
sudo lsof -i :5432

# Kill process
sudo fuser -k 6379/tcp
sudo fuser -k 5432/tcp
```

**View container logs:**
```bash
docker logs redis-stack
docker logs postgres-crawler
```

**Stop containers:**
```bash
docker stop redis-stack postgres-crawler
```

**Remove containers:**
```bash
docker rm redis-stack postgres-crawler
```

## Access PostgreSQL Shell

```bash
docker exec -it postgres-crawler psql -U postgres -d crawler
```

Inside psql:
```sql
-- List tables
\dt

-- View data
SELECT * FROM word_index LIMIT 10;

-- Exit
\q
```
