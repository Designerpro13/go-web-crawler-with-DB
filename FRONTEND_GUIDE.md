# Frontend Integration Guide (ViteJS)

Complete guide for integrating the Web Crawler API with your ViteJS frontend.

## Quick Start

### 1. API Client Setup

Create an API client file: `src/api/crawler.js`

```javascript
const API_BASE = 'http://localhost:8080';

export const crawlerAPI = {
  // Start crawling
  startCrawl: async (urls) => {
    const response = await fetch(`${API_BASE}/api/crawl/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ seed_urls: urls })
    });
    return response.json();
  },

  // Get crawl statistics
  getStats: async () => {
    const response = await fetch(`${API_BASE}/api/stats`);
    return response.json();
  },

  // Search words
  searchWords: async (query) => {
    const response = await fetch(`${API_BASE}/api/search?q=${encodeURIComponent(query)}`);
    return response.json();
  },

  // Get domain statistics
  getDomainStats: async () => {
    const response = await fetch(`${API_BASE}/api/domains`);
    return response.json();
  },

  // Get word postings from Redis
  getWordPostings: async (word) => {
    const response = await fetch(`${API_BASE}/api/word/postings?word=${encodeURIComponent(word)}`);
    return response.json();
  },

  // Get top words
  getTopWords: async (limit = 50) => {
    const response = await fetch(`${API_BASE}/api/words/top?limit=${limit}`);
    return response.json();
  },

  // Get word frequency by domain
  getWordByDomain: async (word) => {
    const response = await fetch(`${API_BASE}/api/word/domains?word=${encodeURIComponent(word)}`);
    return response.json();
  }
};
```

---

## Usage Examples

### Example 1: Start Crawl Page

```vue
<script setup>
import { ref } from 'vue';
import { crawlerAPI } from '@/api/crawler';

const urls = ref(['https://example.com']);
const isLoading = ref(false);
const message = ref('');

const startCrawl = async () => {
  isLoading.value = true;
  try {
    const result = await crawlerAPI.startCrawl(urls.value);
    message.value = result.message;
  } catch (error) {
    message.value = 'Failed to start crawl';
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div>
    <h2>Start Web Crawl</h2>
    <input v-model="urls[0]" placeholder="Enter URL" />
    <button @click="startCrawl" :disabled="isLoading">
      {{ isLoading ? 'Starting...' : 'Start Crawl' }}
    </button>
    <p>{{ message }}</p>
  </div>
</template>
```

---

### Example 2: Monitor Crawl Progress

```vue
<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { crawlerAPI } from '@/api/crawler';

const stats = ref({
  status: 'idle',
  crawled: 0,
  duplicates: 0,
  queue_size: 0
});

const isComplete = ref(false);
let pollInterval = null;

const updateStats = async () => {
  const data = await crawlerAPI.getStats();
  stats.value = data;
  
  // Check if crawl is complete
  if (data.status === 'completed' || data.queue_size === 0) {
    isComplete.value = true;
    clearInterval(pollInterval);
  }
};

onMounted(() => {
  // Poll every 2 seconds
  pollInterval = setInterval(updateStats, 2000);
  updateStats();
});

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval);
});
</script>

<template>
  <div>
    <h2>Crawl Progress</h2>
    <div class="stats">
      <p>Status: <strong>{{ stats.status }}</strong></p>
      <p>Pages Crawled: {{ stats.crawled }}</p>
      <p>Duplicates Skipped: {{ stats.duplicates }}</p>
      <p>Queue Size: {{ stats.queue_size }}</p>
    </div>
    
    <button v-if="isComplete" @click="goToNextPage">
      Proceed to Results →
    </button>
    <p v-else>Crawling in progress...</p>
  </div>
</template>
```

---

### Example 3: Search Results Page

```vue
<script setup>
import { ref } from 'vue';
import { crawlerAPI } from '@/api/crawler';

const searchQuery = ref('');
const results = ref([]);
const isSearching = ref(false);

const search = async () => {
  if (!searchQuery.value) return;
  
  isSearching.value = true;
  try {
    results.value = await crawlerAPI.searchWords(searchQuery.value);
  } catch (error) {
    console.error('Search failed:', error);
  } finally {
    isSearching.value = false;
  }
};
</script>

<template>
  <div>
    <h2>Search Indexed Words</h2>
    <input 
      v-model="searchQuery" 
      @keyup.enter="search"
      placeholder="Search for words..."
    />
    <button @click="search" :disabled="isSearching">Search</button>
    
    <div v-if="results.length > 0" class="results">
      <h3>Results ({{ results.length }})</h3>
      <table>
        <thead>
          <tr>
            <th>Word</th>
            <th>Domain</th>
            <th>Frequency</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="result in results" :key="result.word + result.domain">
            <td>{{ result.word }}</td>
            <td>{{ result.domain }}</td>
            <td>{{ result.frequency }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
```

---

### Example 4: Top Words Dashboard

```vue
<script setup>
import { ref, onMounted } from 'vue';
import { crawlerAPI } from '@/api/crawler';

const topWords = ref([]);
const limit = ref(20);

const loadTopWords = async () => {
  topWords.value = await crawlerAPI.getTopWords(limit.value);
};

onMounted(loadTopWords);
</script>

<template>
  <div>
    <h2>Top Words</h2>
    <select v-model="limit" @change="loadTopWords">
      <option :value="10">Top 10</option>
      <option :value="20">Top 20</option>
      <option :value="50">Top 50</option>
      <option :value="100">Top 100</option>
    </select>
    
    <ul class="word-list">
      <li v-for="(word, index) in topWords" :key="word.word">
        <span class="rank">{{ index + 1 }}</span>
        <span class="word">{{ word.word }}</span>
        <span class="frequency">{{ word.frequency }}</span>
      </li>
    </ul>
  </div>
</template>
```

---

### Example 5: Domain Statistics

```vue
<script setup>
import { ref, onMounted } from 'vue';
import { crawlerAPI } from '@/api/crawler';

const domains = ref([]);

onMounted(async () => {
  domains.value = await crawlerAPI.getDomainStats();
});
</script>

<template>
  <div>
    <h2>Domain Statistics</h2>
    <div class="domain-cards">
      <div v-for="domain in domains" :key="domain.domain" class="card">
        <h3>{{ domain.domain }}</h3>
        <p>Total Words: {{ domain.total_words.toLocaleString() }}</p>
        <p>Unique Words: {{ domain.unique_words.toLocaleString() }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.domain-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 1rem;
}

.card {
  border: 1px solid #ddd;
  padding: 1rem;
  border-radius: 8px;
}
</style>
```

---

## React Examples

### React Hook for Crawl Status

```javascript
import { useState, useEffect } from 'react';
import { crawlerAPI } from './api/crawler';

export const useCrawlStatus = () => {
  const [stats, setStats] = useState({
    status: 'idle',
    crawled: 0,
    duplicates: 0,
    queue_size: 0
  });
  const [isComplete, setIsComplete] = useState(false);

  useEffect(() => {
    const interval = setInterval(async () => {
      const data = await crawlerAPI.getStats();
      setStats(data);
      
      if (data.status === 'completed' || data.queue_size === 0) {
        setIsComplete(true);
        clearInterval(interval);
      }
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  return { stats, isComplete };
};
```

### React Component

```jsx
import { useCrawlStatus } from './hooks/useCrawlStatus';

function CrawlMonitor() {
  const { stats, isComplete } = useCrawlStatus();

  return (
    <div>
      <h2>Crawl Progress</h2>
      <p>Status: <strong>{stats.status}</strong></p>
      <p>Pages Crawled: {stats.crawled}</p>
      <p>Queue Size: {stats.queue_size}</p>
      
      {isComplete && (
        <button onClick={() => navigate('/results')}>
          View Results →
        </button>
      )}
    </div>
  );
}
```

---

## Best Practices

### 1. Error Handling

```javascript
const safeAPICall = async (apiFunction, fallback = null) => {
  try {
    return await apiFunction();
  } catch (error) {
    console.error('API Error:', error);
    return fallback;
  }
};

// Usage
const stats = await safeAPICall(() => crawlerAPI.getStats(), {
  status: 'error',
  crawled: 0,
  duplicates: 0,
  queue_size: 0
});
```

### 2. Debounced Search

```javascript
import { debounce } from 'lodash';

const debouncedSearch = debounce(async (query) => {
  if (query.length < 3) return;
  const results = await crawlerAPI.searchWords(query);
  setResults(results);
}, 500);
```

### 3. Loading States

```javascript
const [isLoading, setIsLoading] = useState(false);

const fetchData = async () => {
  setIsLoading(true);
  try {
    const data = await crawlerAPI.getTopWords();
    setData(data);
  } finally {
    setIsLoading(false);
  }
};
```

---

## Workflow Example

**Multi-Step Crawl Application:**

1. **Step 1: Input URLs** → User enters seed URLs
2. **Step 2: Monitor Progress** → Poll `/api/stats` every 2s
3. **Step 3: Wait for Completion** → Check `status === 'completed'`
4. **Step 4: Show Results** → Display search, top words, domains

```javascript
// Router setup (Vue Router example)
const routes = [
  { path: '/', component: StartCrawl },
  { path: '/progress', component: CrawlProgress },
  { path: '/results', component: SearchResults },
  { path: '/analytics', component: Analytics }
];

// Navigation guard
router.beforeEach(async (to, from, next) => {
  if (to.path === '/results') {
    const stats = await crawlerAPI.getStats();
    if (stats.status !== 'completed') {
      next('/progress'); // Redirect if not complete
    } else {
      next();
    }
  } else {
    next();
  }
});
```

---

## Troubleshooting

### CORS Issues
If you see CORS errors, ensure the backend is running on `localhost:8080`.

### Polling Not Working
Check browser console for errors. Ensure intervals are cleared on component unmount.

### Empty Results
Wait for crawl to complete (`status === 'completed'`) before querying data.

---

## API Response Types (TypeScript)

```typescript
interface CrawlStats {
  status: 'idle' | 'running' | 'completed';
  crawled: number;
  duplicates: number;
  queue_size: number;
}

interface SearchResult {
  word: string;
  domain: string;
  frequency: number;
}

interface DomainStats {
  domain: string;
  total_words: number;
  unique_words: number;
}

interface TopWord {
  word: string;
  frequency: number;
}

interface URLPosting {
  url_hash: string;
  frequency: number;
}

interface WordByDomain {
  domain: string;
  frequency: number;
}
```

---

## Complete Example App Structure

```
src/
├── api/
│   └── crawler.js          # API client
├── components/
│   ├── StartCrawl.vue      # Step 1: Input URLs
│   ├── CrawlProgress.vue   # Step 2: Monitor
│   ├── SearchResults.vue   # Step 3: Search
│   └── Analytics.vue       # Step 4: Stats
├── hooks/
│   └── useCrawlStatus.js   # Reusable hook
└── router/
    └── index.js            # Route guards
```

Happy coding! 
