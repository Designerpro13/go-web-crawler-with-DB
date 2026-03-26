package db

import (
	"context"
	"fmt"
)

type SearchResult struct {
	Word      string `json:"word"`
	Domain    string `json:"domain"`
	Frequency int    `json:"frequency"`
}

type DomainStats struct {
	Domain      string `json:"domain"`
	TotalWords  int    `json:"total_words"`
	UniqueWords int    `json:"unique_words"`
}

type TopWord struct {
	Word      string `json:"word"`
	Frequency int    `json:"frequency"`
}

type WordByDomain struct {
	Domain    string `json:"domain"`
	Frequency int    `json:"frequency"`
}

func (p *PostgresDB) SearchWords(ctx context.Context, query string) ([]SearchResult, error) {
	rows, err := p.DB.QueryContext(ctx, `
		SELECT word, domain, frequency 
		FROM word_index 
		WHERE word ILIKE $1 
		ORDER BY frequency DESC 
		LIMIT 100
	`, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Word, &r.Domain, &r.Frequency); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

func (p *PostgresDB) GetDomainStats(ctx context.Context) ([]DomainStats, error) {
	rows, err := p.DB.QueryContext(ctx, `
		SELECT domain, SUM(frequency) as total_words, COUNT(*) as unique_words
		FROM word_index 
		GROUP BY domain 
		ORDER BY total_words DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("domain stats query failed: %w", err)
	}
	defer rows.Close()

	var stats []DomainStats
	for rows.Next() {
		var s DomainStats
		if err := rows.Scan(&s.Domain, &s.TotalWords, &s.UniqueWords); err != nil {
			continue
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (p *PostgresDB) GetTopWords(ctx context.Context, limit int) ([]TopWord, error) {
	rows, err := p.DB.QueryContext(ctx, `
		SELECT word, SUM(frequency) as total_frequency
		FROM word_index
		GROUP BY word
		ORDER BY total_frequency DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("top words query failed: %w", err)
	}
	defer rows.Close()

	var results []TopWord
	for rows.Next() {
		var w TopWord
		if err := rows.Scan(&w.Word, &w.Frequency); err != nil {
			continue
		}
		results = append(results, w)
	}
	return results, nil
}

func (p *PostgresDB) GetWordByDomain(ctx context.Context, word string) ([]WordByDomain, error) {
	rows, err := p.DB.QueryContext(ctx, `
		SELECT domain, frequency
		FROM word_index
		WHERE word = $1
		ORDER BY frequency DESC
	`, word)
	if err != nil {
		return nil, fmt.Errorf("word by domain query failed: %w", err)
	}
	defer rows.Close()

	var results []WordByDomain
	for rows.Next() {
		var w WordByDomain
		if err := rows.Scan(&w.Domain, &w.Frequency); err != nil {
			continue
		}
		results = append(results, w)
	}
	return results, nil
}
