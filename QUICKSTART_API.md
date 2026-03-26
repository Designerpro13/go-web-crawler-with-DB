# Quick Start - API Server

## Prerequisites
1. PostgreSQL running on `localhost:5432`
2. Redis Stack running on `localhost:6379`

## Run the Server

```bash
go run main.go
```

Server will start on `http://localhost:8080`

## Test the API

### Start a crawl:
```bash
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": ["http://finetranscendentsublimeeclipse.neverssl.com/online/"]}'
```

### Get stats:
```bash
curl http://localhost:8080/api/stats
```

### Search words:
```bash
curl "http://localhost:8080/api/search?q=example"
```

### Get domain stats:
```bash
curl http://localhost:8080/api/domains
```

## Integration with ViteJS

Your ViteJS frontend can now call these endpoints. Example:

```javascript
// In your Vue/React component
const handleStartCrawl = async () => {
  const urls = ['https://example.com'];
  const response = await fetch('http://localhost:8080/api/crawl/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ seed_urls: urls })
  });
  const data = await response.json();
  console.log(data);
};
```
