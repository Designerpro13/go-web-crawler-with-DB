package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	db "github/yeshu2004/go-epics/db"
	"github/yeshu2004/go-epics/indexer"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/html"
)

func initialUrlSeed() []string {
	return []string{
		"https://en.wikipedia.org/wiki/Hindus",
		"https://www.indiatoday.in/",
	}
}

type CrawlerClient struct {
	postgresDB *db.PostgresDB
	redisDB    *redis.Client
	processor  *indexer.Processor
}

var (
	duplicateCount atomic.Int64
	crawledCount   atomic.Int64
	queue          = make(chan string, 10000)
	wg             sync.WaitGroup
	client         = &http.Client{Timeout: 30 * time.Second}
	expected       = 10000000
	fp_rate        = 0.001
)

const (
	workers    = 8
	politeness = 800 * time.Millisecond
	bfKey      = "wiki_bf_2025"
	batchSize  = 100
)

func (c *CrawlerClient) worker(ctx context.Context, rdb *redis.Client) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("LOGGED HERE: Worker exiting due to context cancellation")
			return
		case url, ok := <-queue:
			if !ok {
				log.Println("LOGGED HERE: Worker exiting - queue closed")
				return
			}

			time.Sleep(politeness)

			body, err := fetchBody(url)
			if err != nil {
				log.Printf("LOGGED HERE: Failed to fetch %s: %v", url, err)
				continue
			}

			urlHash := hashURL(url)
			domain := db.ExtractDomain(url)

			log.Printf("LOGGED HERE: Processing URL=%s, Hash=%s, Domain=%s", url, urlHash[:16], domain)

			text := extractText(body)
			freqMap := buildFreqMap(text)

			log.Printf("LOGGED HERE: Extracted %d unique words from %s", len(freqMap), url)

			if err := c.processor.ProcessWordFrequencies(ctx, freqMap, urlHash, domain); err != nil {
				log.Printf("LOGGED HERE: Failed to process word frequencies: %v", err)
			}

			links := extractLinks(body, url)

			newLinksCount := 0
			for _, link := range links {
				hashed := hashURL(link)
				added, err := rdb.BFAdd(ctx, bfKey, hashed).Result()
				if err != nil {
					log.Printf("LOGGED HERE: BFAdd error for %s: %v", link, err)
					continue
				}

				if !added {
					total := duplicateCount.Add(1)
					if total <= 10 || total%1000 == 0 {
						log.Printf("LOGGED HERE: Duplicate skipped (%d total): %s", total, link)
					}
					continue
				}

				newLinksCount++
				select {
				case <-ctx.Done():
					return
				case queue <- link:
				default:
					log.Println("LOGGED HERE: Queue full, dropping link")
				}
			}

			crawled := crawledCount.Add(1)
			log.Printf("LOGGED HERE: Crawled %d pages | Extracted %d new links from %s", crawled, newLinksCount, url)
		}
	}
}

func buildFreqMap(text string) map[string]int {
	freqMap := make(map[string]int)
	text = strings.ToLower(text)

	reg := regexp.MustCompile(`[^\p{L}\p{N}]+`)
	words := reg.Split(text, -1)

	for _, word := range words {
		if len(word) < 3 {
			continue
		}

		hasValidLetter := false
		for _, r := range word {
			if unicode.IsLetter(r) {
				hasValidLetter = true
				break
			}
		}

		if hasValidLetter {
			freqMap[word]++
		}
	}

	return freqMap
}

func extractText(body []byte) string {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		log.Printf("LOGGED HERE: HTML parse error: %v", err)
		return ""
	}

	var textBuilder strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			textBuilder.WriteString(n.Data)
			textBuilder.WriteString(" ")
		}
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return textBuilder.String()
}

func fetchBody(u string) ([]byte, error) {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "MyCollageProjectCrawler (https://github.com/yourname/my-crawler; yourname@example.com)")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", res.StatusCode)
	}

	log.Printf("LOGGED HERE: Successfully fetched %s (Status: %d)", u, res.StatusCode)
	return io.ReadAll(res.Body)
}

func extractLinks(body []byte, baseURLStr string) []string {
	base, _ := url.Parse(baseURLStr)
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	var links []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if full := resolveURL(a.Val, base); full != "" {
						links = append(links, full)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return links
}

func seenBefore(ctx context.Context, rdb *redis.Client, url string) bool {
	hash := hashURL(url)
	exists, err := rdb.BFExists(ctx, bfKey, hash).Result()
	if err != nil {
		if ctx.Err() == context.Canceled {
			return true
		}
		log.Printf("LOGGED HERE: BFExists error: %v", err)
		return true
	}
	return exists
}

func markSeen(ctx context.Context, rdb *redis.Client, url string) {
	hash := hashURL(url)
	if err := rdb.BFAdd(ctx, bfKey, hash).Err(); err != nil {
		log.Printf("LOGGED HERE: BFAdd failed for %s: %v", url, err)
	}
}

func hashURL(u string) string {
	h := sha256.Sum256([]byte(u))
	return hex.EncodeToString(h[:])
}

func resolveURL(href string, base *url.URL) string {
	u, err := url.Parse(href)
	if err != nil || (u.Host != "" && u.Host != base.Host) {
		return ""
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	resolved.Fragment = ""
	resolved.RawQuery = ""
	resolved.Path = strings.ToLower(resolved.Path)
	if resolved.Path != "" && !strings.HasSuffix(resolved.Path, "/") {
		resolved.Path = strings.TrimSuffix(resolved.Path, "/")
	}
	return resolved.String()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("LOGGED HERE: Starting optimized web crawler with PostgreSQL + Redis")

	// PostgreSQL connection
	postgresConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=crawler sslmode=disable"
	postgresDB, err := db.PostgresInit(ctx, postgresConnStr)
	if err != nil {
		log.Fatal("LOGGED HERE: PostgreSQL connection failed:", err)
	}
	defer postgresDB.Close()

	if err := postgresDB.InitSchema(ctx); err != nil {
		log.Fatal("LOGGED HERE: Schema initialization failed:", err)
	}

	// Redis connection
	rdb, err := db.RedisInit(ctx)
	if err != nil {
		log.Fatal("LOGGED HERE: Redis connection failed:", err)
	}
	defer rdb.Close()

	if err := db.InitializeBloomFilter(ctx, rdb, bfKey, fp_rate, int64(expected)); err != nil {
		log.Fatal("LOGGED HERE: Bloom filter init failed:", err)
	}

	// Initialize inverted index and processor
	invertedIdx := indexer.NewInvertedIndex(rdb)
	processor := indexer.NewProcessor(invertedIdx, postgresDB, batchSize)

	cli := &CrawlerClient{
		postgresDB: postgresDB,
		redisDB:    rdb,
		processor:  processor,
	}

	// Graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("LOGGED HERE: Shutting down gracefully...")
		cancel()
		
		log.Println("LOGGED HERE: Flushing remaining data to PostgreSQL...")
		if err := processor.Flush(ctx); err != nil {
			log.Printf("LOGGED HERE: Final flush error: %v", err)
		}
		
		close(queue)
	}()

	// Seed URLs
	for _, seed := range initialUrlSeed() {
		if !seenBefore(ctx, rdb, seed) {
			markSeen(ctx, rdb, seed)
			queue <- seed
			log.Printf("LOGGED HERE: Seeded URL: %s", seed)
		}
	}

	// Start workers
	log.Printf("LOGGED HERE: Starting %d workers", workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go cli.worker(ctx, rdb)
	}

	wg.Wait()
	
	log.Println("LOGGED HERE: Final flush before exit...")
	if err := processor.Flush(context.Background()); err != nil {
		log.Printf("LOGGED HERE: Final flush error: %v", err)
	}
	
	log.Printf("LOGGED HERE: Crawl completed! Total pages crawled: %d, Duplicates skipped: %d", 
		crawledCount.Load(), duplicateCount.Load())
}
