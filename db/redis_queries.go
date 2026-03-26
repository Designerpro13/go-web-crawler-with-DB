package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type RedisQueries struct {
	client *redis.Client
}

type URLPosting struct {
	URLHash   string `json:"url_hash"`
	Frequency int    `json:"frequency"`
}

func NewRedisQueries(client *redis.Client) *RedisQueries {
	return &RedisQueries{client: client}
}

func (r *RedisQueries) GetWordPostings(ctx context.Context, word string) ([]URLPosting, error) {
	key := fmt.Sprintf("idx:%s", strings.ToLower(word))
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get postings: %w", err)
	}

	var postings []URLPosting
	for _, member := range members {
		parts := strings.Split(member, ":")
		if len(parts) != 2 {
			continue
		}
		freq, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		postings = append(postings, URLPosting{
			URLHash:   parts[0],
			Frequency: freq,
		})
	}

	return postings, nil
}
