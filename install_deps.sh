#!/bin/bash

# Install PostgreSQL driver
echo "Installing dependencies..."

# Add to go.mod
cat >> go.mod << 'EOF'

require github.com/lib/pq v1.10.9
EOF

echo "Dependencies added to go.mod"
echo "Run: go mod tidy"
