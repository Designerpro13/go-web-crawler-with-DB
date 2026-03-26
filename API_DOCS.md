# API Documentation

## Base URL
```
http://localhost:8080
```

## Endpoints

### 1. Start Crawling
**POST** `/api/crawl/start`

Start a new crawl with seed URLs.

**Request Body:**
```json
{
  "seed_urls": [
    "https://example.com",
    "https://another-site.com"
  ]
}
```

**Response:**
```json
{
  "status": "started",
  "message": "Crawling started successfully"
}
```

### 2. Search Words
**GET** `/api/search?q=<query>`

Search for words in the indexed database.

**Query Parameters:**
- `q` (required): Search query string

**Response:**
```json
[
  {
    "word": "example",
    "domain": "example.com",
    "frequency": 42
  }
]
```

### 3. Get Crawl Statistics
**GET** `/api/stats`

Get current crawling statistics.

**Response:**
```json
{
  "crawled": 1234,
  "duplicates": 567
}
```

### 4. Get Domain Statistics
**GET** `/api/domains`

Get aggregated statistics by domain.

**Response:**
```json
[
  {
    "domain": "example.com",
    "total_words": 50000,
    "unique_words": 1500
  }
]
```

## CORS
All endpoints support CORS with `Access-Control-Allow-Origin: *`

## Example Usage (JavaScript/ViteJS)

```javascript
// Start crawling
const startCrawl = async (urls) => {
  const response = await fetch('http://localhost:8080/api/crawl/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ seed_urls: urls })
  });
  return response.json();
};

// Search words
const searchWords = async (query) => {
  const response = await fetch(`http://localhost:8080/api/search?q=${query}`);
  return response.json();
};

// Get stats
const getStats = async () => {
  const response = await fetch('http://localhost:8080/api/stats');
  return response.json();
};

// Get domain stats
const getDomainStats = async () => {
  const response = await fetch('http://localhost:8080/api/domains');
  return response.json();
};
```

### 5. Get Word Postings (Redis)
**GET** `/api/word/postings?word=<word>`

Get all URL hashes where a word appears with frequencies from Redis inverted index.

**Query Parameters:**
- `word` (required): The word to lookup

**Response:**
```json
[
  {
    "url_hash": "a1b2c3d4e5f6...",
    "frequency": 15
  },
  {
    "url_hash": "f6e5d4c3b2a1...",
    "frequency": 8
  }
]
```

### 6. Get Top Words
**GET** `/api/words/top?limit=<number>`

Get the most frequent words across all domains.

**Query Parameters:**
- `limit` (optional): Number of results (default: 50)

**Response:**
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

### 7. Get Word Frequency by Domain
**GET** `/api/word/domains?word=<word>`

Get frequency of a specific word across different domains.

**Query Parameters:**
- `word` (required): The word to lookup

**Response:**
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

## New JavaScript Examples

```javascript
// Get word postings from Redis
const getWordPostings = async (word) => {
  const response = await fetch(`http://localhost:8080/api/word/postings?word=${word}`);
  return response.json();
};

// Get top words
const getTopWords = async (limit = 50) => {
  const response = await fetch(`http://localhost:8080/api/words/top?limit=${limit}`);
  return response.json();
};

// Get word frequency by domain
const getWordByDomain = async (word) => {
  const response = await fetch(`http://localhost:8080/api/word/domains?word=${word}`);
  return response.json();
};
```
