-- PostgreSQL Setup Script for Web Crawler
-- Run this inside the PostgreSQL container

-- Create word_index table
CREATE TABLE IF NOT EXISTS word_index (
    id SERIAL PRIMARY KEY,
    word VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    frequency INTEGER DEFAULT 1,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(word, domain)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_word ON word_index(word);
CREATE INDEX IF NOT EXISTS idx_domain ON word_index(domain);
CREATE INDEX IF NOT EXISTS idx_frequency ON word_index(frequency DESC);

-- Sample queries for analysis
-- Top words by frequency
-- SELECT word, SUM(frequency) as total_freq FROM word_index GROUP BY word ORDER BY total_freq DESC LIMIT 20;

-- Top domains by word count
-- SELECT domain, COUNT(DISTINCT word) as unique_words FROM word_index GROUP BY domain ORDER BY unique_words DESC LIMIT 10;

-- Words across multiple domains
-- SELECT word, COUNT(DISTINCT domain) as domain_count FROM word_index GROUP BY word HAVING COUNT(DISTINCT domain) > 1 ORDER BY domain_count DESC LIMIT 20;
