#!/bin/bash

echo "Setting up Docker containers for Web Crawler..."

# Stop and remove existing containers if any
echo "Cleaning up existing containers..."
docker stop redis-stack postgres-crawler 2>/dev/null
docker rm redis-stack postgres-crawler 2>/dev/null

# Kill processes on ports if needed
echo "Freeing up ports..."
sudo fuser -k 6379/tcp 2>/dev/null
sudo fuser -k 5432/tcp 2>/dev/null
sleep 2

# Start Redis Stack
echo "Starting Redis Stack..."
docker run -d \
  --name redis-stack \
  -p 6379:6379 \
  redis/redis-stack:latest

# Start PostgreSQL
echo "Starting PostgreSQL..."
docker run -d \
  --name postgres-crawler \
  -e POSTGRES_PASSWORD=crawler123 \
  -e POSTGRES_DB=crawler \
  -p 5432:5432 \
  postgres:15-alpine

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Create schema
echo "Creating database schema..."
docker exec -i postgres-crawler psql -U postgres -d crawler << 'EOF'
CREATE TABLE IF NOT EXISTS word_index (
    id SERIAL PRIMARY KEY,
    word VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    frequency INTEGER DEFAULT 1,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(word, domain)
);

CREATE INDEX IF NOT EXISTS idx_word ON word_index(word);
CREATE INDEX IF NOT EXISTS idx_domain ON word_index(domain);
CREATE INDEX IF NOT EXISTS idx_frequency ON word_index(frequency DESC);

\dt
EOF

# Verify containers
echo ""
echo "Verifying containers..."
docker ps | grep -E "redis-stack|postgres-crawler"

echo ""
echo "✅ Setup complete!"
echo ""
echo "Connection details:"
echo "  Redis:      localhost:6379"
echo "  PostgreSQL: localhost:5432"
echo "  DB Name:    crawler"
echo "  User:       postgres"
echo "  Password:   crawler123"
echo ""
echo "Test connections:"
echo "  redis-cli ping"
echo "  docker exec -it postgres-crawler psql -U postgres -d crawler"
