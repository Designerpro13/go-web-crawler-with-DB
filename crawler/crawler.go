package crawler

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

type Crawler struct {
	PostgresDB *db.PostgresDB
	RedisDB    *redis.Client
	Processor  *indexer.Processor

	duplicateCount atomic.Int64
	crawledCount   atomic.Int64
	queue          chan string
	wg             sync.WaitGroup
	client         *http.Client
	ctx            context.Context
	cancel         context.CancelFunc
}

const (
	workers    = 8
	politeness = 800 * time.Millisecond
	bfKey      = "wiki_bf_2025"
	queueSize  = 10000
)

func New(postgresDB *db.PostgresDB, redisDB *redis.Client, processor *indexer.Processor) *Crawler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Crawler{
		PostgresDB: postgresDB,
		RedisDB:    redisDB,
		Processor:  processor,
		queue:      make(chan string, queueSize),
		client:     &http.Client{Timeout: 30 * time.Second},
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (c *Crawler) Start(seedURLs []string) error {
	for _, seed := range seedURLs {
		if !c.seenBefore(seed) {
			c.markSeen(seed)
			c.queue <- seed
			log.Printf("Seeded URL: %s", seed)
		}
	}

	for i := 0; i < workers; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	return nil
}

func (c *Crawler) Stop() {
	log.Println("Stopping crawler...")
	c.cancel()
	close(c.queue)
	c.wg.Wait()
	c.Processor.Flush(context.Background())
	log.Println("Crawler stopped")
}

func (c *Crawler) GetStats() (crawled, duplicates int64) {
	return c.crawledCount.Load(), c.duplicateCount.Load()
}

func (c *Crawler) worker() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case url, ok := <-c.queue:
			if !ok {
				return
			}

			time.Sleep(politeness)

			body, err := c.fetchBody(url)
			if err != nil {
				log.Printf("Failed to fetch %s: %v", url, err)
				continue
			}

			urlHash := hashURL(url)
			domain := db.ExtractDomain(url)

			text := extractText(body)
			freqMap := buildFreqMap(text)

			if err := c.Processor.ProcessWordFrequencies(c.ctx, freqMap, urlHash, domain); err != nil {
				log.Printf("Failed to process word frequencies: %v", err)
			}

			links := extractLinks(body, url)

			newLinksCount := 0
			for _, link := range links {
				hashed := hashURL(link)
				added, err := c.RedisDB.BFAdd(c.ctx, bfKey, hashed).Result()
				if err != nil {
					continue
				}

				if !added {
					c.duplicateCount.Add(1)
					continue
				}

				newLinksCount++
				select {
				case <-c.ctx.Done():
					return
				case c.queue <- link:
				default:
				}
			}

			crawled := c.crawledCount.Add(1)
			log.Printf("Crawled %d pages | Extracted %d new links from %s", crawled, newLinksCount, url)
		}
	}
}

func (c *Crawler) fetchBody(u string) ([]byte, error) {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "MyCollageProjectCrawler")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", res.StatusCode)
	}

	return io.ReadAll(res.Body)
}

func (c *Crawler) seenBefore(url string) bool {
	hash := hashURL(url)
	exists, err := c.RedisDB.BFExists(c.ctx, bfKey, hash).Result()
	if err != nil {
		return true
	}
	return exists
}

func (c *Crawler) markSeen(url string) {
	hash := hashURL(url)
	c.RedisDB.BFAdd(c.ctx, bfKey, hash)
}

func hashURL(u string) string {
	h := sha256.Sum256([]byte(u))
	return hex.EncodeToString(h[:])
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
