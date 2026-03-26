# Testing Documentation

Complete testing guide with curl commands and environment troubleshooting.

---

## Prerequisites Check

### 1. Verify Docker Containers

```bash
# Check if containers are running
docker ps

# Expected output: redis-stack and postgres containers
```

**If containers not running:**
```bash
# Start Redis
docker start redis-stack

# Start PostgreSQL
docker start postgres

# Or create new ones
docker run -d -p 6379:6379 --name redis-stack redis/redis-stack:latest
docker run -d -p 5432:5432 --name postgres -e POSTGRES_PASSWORD=crawler123 -e POSTGRES_DB=crawler postgres:latest
```

### 2. Verify Ports

```bash
# Check if ports are in use
netstat -tuln | grep 8080
netstat -tuln | grep 5432
netstat -tuln | grep 6379

# Or on macOS
lsof -i :8080
lsof -i :5432
lsof -i :6379
```

### 3. Test Database Connections

**PostgreSQL:**
```bash
PGPASSWORD=crawler123 psql -h localhost -p 5432 -U postgres -d crawler -c "SELECT 1;"
```

**Redis:**
```bash
redis-cli -h localhost -p 6379 ping
# Expected: PONG
```

---

## API Testing

### Test 1: Server Health Check

```bash
# Check if server is running
curl -I http://localhost:8080/api/stats

# Expected: HTTP/1.1 200 OK
```

**Bug Fix:** If connection refused:
```bash
# Check if server is running
ps aux | grep "go run main.go"

# Start server
go run main.go
```

---

### Test 2: Start Crawl

**Basic Test:**
```bash
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": ["http://example.com"]}'
```

**Expected Response:**
```json
{
  "status": "started",
  "message": "Crawling started successfully"
}
```

**Multiple URLs:**
```bash
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": ["http://example.com", "http://example.org"]}'
```

**Error Cases:**

1. **Empty URLs:**
```bash
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": []}'

# Expected: 400 Bad Request - "seed_urls cannot be empty"
```

2. **Invalid JSON:**
```bash
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{invalid}'

# Expected: 400 Bad Request - "Invalid request body"
```

3. **Wrong Method:**
```bash
curl -X GET http://localhost:8080/api/crawl/start

# Expected: 405 Method Not Allowed
```

---

### Test 3: Monitor Crawl Progress

```bash
# Get current stats
curl http://localhost:8080/api/stats

# Pretty print with jq
curl -s http://localhost:8080/api/stats | jq
```

**Expected Response:**
```json
{
  "status": "running",
  "crawled": 5,
  "duplicates": 2,
  "queue_size": 12
}
```

**Status Values:**
- `idle` - No crawl running
- `running` - Actively crawling
- `completed` - Crawl finished

**Continuous Monitoring:**
```bash
# Watch stats every 2 seconds
watch -n 2 'curl -s http://localhost:8080/api/stats | jq'

# Or with loop
while true; do
  curl -s http://localhost:8080/api/stats | jq
  sleep 2
done
```

---

### Test 4: Search Words

**Basic Search:**
```bash
curl "http://localhost:8080/api/search?q=example"
```

**URL Encoded Search:**
```bash
curl "http://localhost:8080/api/search?q=web%20crawler"
```

**Case Insensitive:**
```bash
curl "http://localhost:8080/api/search?q=EXAMPLE"
```

**Expected Response:**
```json
[
  {
    "word": "example",
    "domain": "example.com",
    "frequency": 25
  }
]
```

**Error Cases:**

1. **Missing Query:**
```bash
curl "http://localhost:8080/api/search"

# Expected: 400 Bad Request - "query parameter 'q' is required"
```

2. **No Results:**
```bash
curl "http://localhost:8080/api/search?q=nonexistentword123"

# Expected: [] (empty array)
```

---

### Test 5: Domain Statistics

```bash
curl http://localhost:8080/api/domains | jq
```

**Expected Response:**
```json
[
  {
    "domain": "example.com",
    "total_words": 5000,
    "unique_words": 1200
  }
]
```

---

### Test 6: Word Postings (Redis)

```bash
curl "http://localhost:8080/api/word/postings?word=example" | jq
```

**Expected Response:**
```json
[
  {
    "url_hash": "a1b2c3d4e5f6...",
    "frequency": 15
  }
]
```

**Error Case:**
```bash
curl "http://localhost:8080/api/word/postings"

# Expected: 400 Bad Request - "query parameter 'word' is required"
```

---

### Test 7: Top Words

**Default (50 words):**
```bash
curl "http://localhost:8080/api/words/top" | jq
```

**Custom Limit:**
```bash
curl "http://localhost:8080/api/words/top?limit=10" | jq
curl "http://localhost:8080/api/words/top?limit=100" | jq
```

**Expected Response:**
```json
[
  {
    "word": "the",
    "frequency": 15000
  },
  {
    "word": "example",
    "frequency": 8500
  }
]
```

---

### Test 8: Word by Domain

```bash
curl "http://localhost:8080/api/word/domains?word=example" | jq
```

**Expected Response:**
```json
[
  {
    "domain": "example.com",
    "frequency": 250
  },
  {
    "domain": "example.org",
    "frequency": 120
  }
]
```

---

## Complete Test Flow

```bash
#!/bin/bash

echo "=== Web Crawler API Test Suite ==="

# Test 1: Start crawl
echo -e "\n[1/8] Starting crawl..."
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": ["http://example.com"]}' | jq

# Test 2: Wait for crawling
echo -e "\n[2/8] Waiting 10 seconds for crawling..."
sleep 10

# Test 3: Check stats
echo -e "\n[3/8] Checking stats..."
curl -s http://localhost:8080/api/stats | jq

# Test 4: Search words
echo -e "\n[4/8] Searching for 'example'..."
curl -s "http://localhost:8080/api/search?q=example" | jq

# Test 5: Domain stats
echo -e "\n[5/8] Getting domain stats..."
curl -s http://localhost:8080/api/domains | jq

# Test 6: Top words
echo -e "\n[6/8] Getting top 10 words..."
curl -s "http://localhost:8080/api/words/top?limit=10" | jq

# Test 7: Word postings
echo -e "\n[7/8] Getting word postings for 'example'..."
curl -s "http://localhost:8080/api/word/postings?word=example" | jq

# Test 8: Word by domain
echo -e "\n[8/8] Getting word frequency by domain..."
curl -s "http://localhost:8080/api/word/domains?word=example" | jq

echo -e "\n=== Test Suite Complete ==="
```

Save as `test-api.sh` and run:
```bash
chmod +x test-api.sh
./test-api.sh
```

---

## Environment Troubleshooting

### Bug 1: Connection Refused (Port 8080)

**Symptom:**
```
curl: (7) Failed to connect to localhost port 8080: Connection refused
```

**Fix:**
```bash
# Check if server is running
ps aux | grep "go run main.go"

# Start server
go run main.go

# Or check if port is blocked
sudo lsof -i :8080
```

---

### Bug 2: PostgreSQL Connection Failed

**Symptom:**
```
PostgreSQL connection failed: dial tcp [::1]:5432: connect: connection refused
```

**Fix:**
```bash
# Check if PostgreSQL container is running
docker ps | grep postgres

# Start container
docker start postgres

# Or recreate
docker rm -f postgres
docker run -d -p 5432:5432 --name postgres \
  -e POSTGRES_PASSWORD=crawler123 \
  -e POSTGRES_DB=crawler \
  postgres:latest

# Test connection
PGPASSWORD=crawler123 psql -h localhost -p 5432 -U postgres -d crawler -c "SELECT 1;"
```

---

### Bug 3: Redis Connection Failed

**Symptom:**
```
Redis connection failed: dial tcp [::1]:6379: connect: connection refused
```

**Fix:**
```bash
# Check if Redis container is running
docker ps | grep redis

# Start container
docker start redis-stack

# Or recreate
docker rm -f redis-stack
docker run -d -p 6379:6379 --name redis-stack redis/redis-stack:latest

# Test connection
redis-cli -h localhost -p 6379 ping
```

---

### Bug 4: Bloom Filter Not Initialized

**Symptom:**
```
Bloom filter init failed: ERR unknown command 'BF.RESERVE'
```

**Fix:**
```bash
# You need Redis Stack, not regular Redis
docker rm -f redis-stack
docker run -d -p 6379:6379 --name redis-stack redis/redis-stack:latest

# Verify Bloom Filter module
redis-cli -h localhost -p 6379 MODULE LIST
```

---

### Bug 5: Empty Search Results

**Symptom:**
```json
[]
```

**Possible Causes:**
1. Crawl not completed yet
2. Word doesn't exist
3. Database empty

**Fix:**
```bash
# Check crawl status
curl -s http://localhost:8080/api/stats | jq

# Wait for completion
# status should be "completed" and queue_size should be 0

# Check if database has data
PGPASSWORD=crawler123 psql -h localhost -p 5432 -U postgres -d crawler \
  -c "SELECT COUNT(*) FROM word_index;"
```

---

### Bug 6: Timeout Errors

**Symptom:**
```
Failed to fetch http://example.com: context deadline exceeded
```

**Causes:**
- URL is down
- Network issues
- Firewall blocking

**Fix:**
```bash
# Test URL manually
curl -I http://example.com

# Use reliable test URLs
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": ["http://example.com", "http://example.org"]}'
```

---

### Bug 7: Port Already in Use

**Symptom:**
```
listen tcp :8080: bind: address already in use
```

**Fix:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use different port in main.go
# Change: server.Start("8080")
# To: server.Start("8081")
```

---

### Bug 8: CORS Errors (Frontend)

**Symptom:**
```
Access to fetch at 'http://localhost:8080/api/stats' from origin 'http://localhost:5173' 
has been blocked by CORS policy
```

**Fix:**
Server already has CORS enabled. If still seeing errors:

```bash
# Test CORS headers
curl -I http://localhost:8080/api/stats

# Should see:
# Access-Control-Allow-Origin: *
# Access-Control-Allow-Methods: GET, POST, OPTIONS
```

If missing, check `api/server.go` has `enableCORS` wrapper on all routes.

---

## Database Reset (Development)

**Linux/macOS:**
```bash
./reset-db.sh
```

**Windows:**
```powershell
.\reset-db.ps1
```

**Manual Reset:**
```bash
# PostgreSQL
PGPASSWORD=crawler123 psql -h localhost -p 5432 -U postgres -d crawler \
  -c "TRUNCATE TABLE word_index RESTART IDENTITY CASCADE;"

# Redis
redis-cli -h localhost -p 6379 FLUSHALL

# Reinitialize Bloom Filter
redis-cli -h localhost -p 6379 BF.RESERVE wiki_bf_2025 0.001 10000000
```

---

## Performance Testing

### Load Test: Multiple Concurrent Requests

```bash
# Install Apache Bench (if needed)
# Ubuntu: sudo apt-get install apache2-utils
# macOS: brew install ab

# Test stats endpoint
ab -n 1000 -c 10 http://localhost:8080/api/stats

# Test search endpoint
ab -n 100 -c 5 "http://localhost:8080/api/search?q=example"
```

### Stress Test: Large Crawl

```bash
# Start crawl with multiple seed URLs
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{
    "seed_urls": [
      "http://example.com",
      "http://example.org",
      "http://example.net"
    ]
  }'

# Monitor memory usage
watch -n 1 'ps aux | grep "go run main.go"'
```

---

## Automated Testing Script

Save as `run-tests.sh`:

```bash
#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

test_count=0
pass_count=0

run_test() {
  test_count=$((test_count + 1))
  echo -e "\n[Test $test_count] $1"
  
  if eval "$2"; then
    echo -e "${GREEN}✓ PASS${NC}"
    pass_count=$((pass_count + 1))
  else
    echo -e "${RED}✗ FAIL${NC}"
  fi
}

echo "=== API Test Suite ==="

# Test 1: Server is running
run_test "Server health check" \
  "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/stats | grep -q 200"

# Test 2: Start crawl
run_test "Start crawl endpoint" \
  "curl -s -X POST http://localhost:8080/api/crawl/start \
    -H 'Content-Type: application/json' \
    -d '{\"seed_urls\": [\"http://example.com\"]}' | grep -q 'started'"

# Test 3: Stats endpoint
run_test "Stats endpoint returns JSON" \
  "curl -s http://localhost:8080/api/stats | jq -e '.status' > /dev/null"

# Test 4: Search endpoint
run_test "Search endpoint" \
  "curl -s 'http://localhost:8080/api/search?q=test' | jq -e 'type == \"array\"' > /dev/null"

# Test 5: Domains endpoint
run_test "Domains endpoint" \
  "curl -s http://localhost:8080/api/domains | jq -e 'type == \"array\"' > /dev/null"

# Test 6: Top words endpoint
run_test "Top words endpoint" \
  "curl -s 'http://localhost:8080/api/words/top?limit=10' | jq -e 'type == \"array\"' > /dev/null"

# Test 7: Error handling - empty seed_urls
run_test "Error handling - empty URLs" \
  "curl -s -X POST http://localhost:8080/api/crawl/start \
    -H 'Content-Type: application/json' \
    -d '{\"seed_urls\": []}' | grep -q 'cannot be empty'"

# Test 8: Error handling - missing query param
run_test "Error handling - missing query" \
  "curl -s 'http://localhost:8080/api/search' | grep -q 'required'"

echo -e "\n=== Results ==="
echo "Passed: $pass_count/$test_count"

if [ $pass_count -eq $test_count ]; then
  echo -e "${GREEN}All tests passed!${NC}"
  exit 0
else
  echo -e "${RED}Some tests failed${NC}"
  exit 1
fi
```

Run:
```bash
chmod +x run-tests.sh
./run-tests.sh
```

---

## Quick Reference

### Essential Commands

```bash
# Start server
go run main.go

# Start crawl
curl -X POST http://localhost:8080/api/crawl/start \
  -H "Content-Type: application/json" \
  -d '{"seed_urls": ["http://example.com"]}'

# Check status
curl http://localhost:8080/api/stats | jq

# Search
curl "http://localhost:8080/api/search?q=example" | jq

# Reset database
./reset-db.sh
```

### Docker Commands

```bash
# Start containers
docker start redis-stack postgres

# Stop containers
docker stop redis-stack postgres

# View logs
docker logs redis-stack
docker logs postgres

# Remove containers
docker rm -f redis-stack postgres
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:latest
        env:
          POSTGRES_PASSWORD: crawler123
          POSTGRES_DB: crawler
        ports:
          - 5432:5432
      
      redis:
        image: redis/redis-stack:latest
        ports:
          - 6379:6379
    
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.24
      
      - name: Start server
        run: go run main.go &
        
      - name: Wait for server
        run: sleep 5
      
      - name: Run tests
        run: ./run-tests.sh
```

---

**Happy Testing!**
