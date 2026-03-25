package indexer

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type InvertedIndex struct {
	client *redis.Client
}

func NewInvertedIndex(client *redis.Client) *InvertedIndex {
	return &InvertedIndex{client: client}
}

func (idx *InvertedIndex) AddPosting(ctx context.Context, word, urlHash string, freq int) error {
	key := fmt.Sprintf("idx:%s", word)
	value := fmt.Sprintf("%s:%d", urlHash, freq)

	if err := idx.client.SAdd(ctx, key, value).Err(); err != nil {
		return fmt.Errorf("failed to add posting: %w", err)
	}

	return nil
}

func (idx *InvertedIndex) BatchAddPostings(ctx context.Context, postings map[string]map[string]int) error {
	if len(postings) == 0 {
		return nil
	}

	pipe := idx.client.Pipeline()
	count := 0

	for word, urlFreqs := range postings {
		key := fmt.Sprintf("idx:%s", word)
		for urlHash, freq := range urlFreqs {
			value := fmt.Sprintf("%s:%d", urlHash, freq)
			pipe.SAdd(ctx, key, value)
			count++
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	log.Printf("LOGGED HERE: Added %d postings to Redis inverted index", count)
	return nil
}

func (idx *InvertedIndex) GetPostings(ctx context.Context, word string) ([]string, error) {
	key := fmt.Sprintf("idx:%s", word)
	return idx.client.SMembers(ctx, key).Result()
}
