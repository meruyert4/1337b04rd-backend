#!/bin/bash

# 1337b04rd Backend Startup Script

echo "Starting 1337b04rd Backend..."

# Set default environment variables if not already set
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-1337b04rd}
export DB_SSLMODE=${DB_SSLMODE:-disable}

export MINIO_ENDPOINT=${MINIO_ENDPOINT:-localhost:9000}
export MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}
export MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}
export MINIO_USE_SSL=${MINIO_USE_SSL:-false}

export SERVER_PORT=${SERVER_PORT:-8080}
export SERVER_HOST=${SERVER_HOST:-localhost}

export LOG_LEVEL=${LOG_LEVEL:-info}

echo "Configuration:"
echo "  Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "  MinIO: $MINIO_ENDPOINT"
echo "  Server: $SERVER_HOST:$SERVER_PORT"
echo "  Log Level: $LOG_LEVEL"
echo ""

# Check if PostgreSQL is running
echo "Checking PostgreSQL connection..."
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
    echo "ERROR: PostgreSQL is not accessible at $DB_HOST:$DB_PORT"
    echo "Please ensure PostgreSQL is running and accessible"
    exit 1
fi
echo "✓ PostgreSQL is accessible"

# Check if MinIO is running
echo "Checking MinIO connection..."
if ! curl -s "http://$MINIO_ENDPOINT/minio/health/live" > /dev/null 2>&1; then
    echo "WARNING: MinIO is not accessible at $MINIO_ENDPOINT"
    echo "Image uploads will not work, but the application will start"
    echo "Please ensure MinIO is running for full functionality"
else
    echo "✓ MinIO is accessible"
fi

echo ""
echo "Starting application..."
echo "Press Ctrl+C to stop"
echo ""

# Run the application
go run cmd/main.go
