# Web Crawler API

Event-triggered web crawler with REST API, built in Go. Designed for crawling web pages with BFS approach and exposing results via HTTP endpoints.

## Features

- REST API for event-triggered crawling
- Multi-threaded crawling (8 workers)
- Bloom Filter for duplicate URL detection
- PostgreSQL for word indexing
- Redis for inverted index & bloom filter
- Full-text search API
- Real-time statistics
- CORS enabled for frontend integration

## Prerequisites

1. **PostgreSQL** (port 5432)
2. **Redis Stack** (port 6379)

### Quick Setup with Docker:

**Linux/macOS:**
```bash
# Redis Stack
docker run -d -p 6379:6379 --name redis-stack redis/redis-stack:latest

# PostgreSQL
docker run -d -p 5432:5432 --name postgres \
  -e POSTGRES_PASSWORD=crawler123 \
  -e POSTGRES_DB=crawler \
  postgres:latest
```

**Windows (PowerShell/CMD):**
```powershell
# Redis Stack
docker run -d -p 6379:6379 --name redis-stack redis/redis-stack:latest

# PostgreSQL
docker run -d -p 5432:5432 --name postgres -e POSTGRES_PASSWORD=crawler123 -e POSTGRES_DB=crawler postgres:latest
```

> See [DOCKER_WINDOWS.md](DOCKER_WINDOWS.md) for detailed Windows setup instructions

## Run

```bash
go run main.go
```

Server starts on `http://localhost:8080`

## API Endpoints

### Start Crawling
```bash
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": ["https://example.com"]}'
```

### Search Words
```bash
curl "http://localhost:8080/api/search?q=example"
```

### Get Statistics
```bash
curl http://localhost:8080/api/stats
```

### Get Domain Stats
```bash
curl http://localhost:8080/api/domains
```

## Frontend Integration (ViteJS)

```javascript
const startCrawl = async (urls) => {
  const res = await fetch('http://localhost:8080/api/crawl/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ seed_urls: urls })
  });
  return res.json();
};
```

See [API_DOCS.md](API_DOCS.md) for complete API documentation.
