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
	Domain     string `json:"domain"`
	TotalWords int    `json:"total_words"`
	UniqueWords int   `json:"unique_words"`
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
