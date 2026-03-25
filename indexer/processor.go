package indexer

import (
	"context"
	"log"

	"github/yeshu2004/go-epics/db"
)

type Processor struct {
	invertedIdx *InvertedIndex
	postgresDB  *db.PostgresDB
	batchSize   int
	buffer      []db.WordPosting
}

func NewProcessor(invertedIdx *InvertedIndex, postgresDB *db.PostgresDB, batchSize int) *Processor {
	return &Processor{
		invertedIdx: invertedIdx,
		postgresDB:  postgresDB,
		batchSize:   batchSize,
		buffer:      make([]db.WordPosting, 0, batchSize),
	}
}

func (p *Processor) ProcessWordFrequencies(ctx context.Context, freqMap map[string]int, urlHash, domain string) error {
	redisPostings := make(map[string]map[string]int)

	for word, freq := range freqMap {
		if redisPostings[word] == nil {
			redisPostings[word] = make(map[string]int)
		}
		redisPostings[word][urlHash] = freq

		p.buffer = append(p.buffer, db.WordPosting{
			Word:      word,
			URLHash:   urlHash,
			Domain:    domain,
			Frequency: freq,
		})
	}

	if err := p.invertedIdx.BatchAddPostings(ctx, redisPostings); err != nil {
		log.Printf("LOGGED HERE: Redis indexing error: %v", err)
	}

	if len(p.buffer) >= p.batchSize {
		return p.Flush(ctx)
	}

	return nil
}

func (p *Processor) Flush(ctx context.Context) error {
	if len(p.buffer) == 0 {
		return nil
	}

	if err := p.postgresDB.BatchUpsertWords(ctx, p.buffer); err != nil {
		log.Printf("LOGGED HERE: PostgreSQL batch upsert error: %v", err)
		return err
	}

	log.Printf("LOGGED HERE: Flushed %d word postings to PostgreSQL", len(p.buffer))
	p.buffer = p.buffer[:0]
	return nil
}
