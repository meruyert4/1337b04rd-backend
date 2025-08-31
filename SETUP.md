# Quick Setup Guide

## Prerequisites

1. **Go 1.23+** installed
2. **PostgreSQL** running locally
3. **MinIO** running locally (optional, for image uploads)

## Quick Start

1. **Install dependencies**
   ```bash
   go mod tidy
   ```

2. **Setup PostgreSQL**
   ```bash
   # Create database
   createdb 1337b04rd
   
   # Or connect to PostgreSQL and run:
   # CREATE DATABASE 1337b04rd;
   ```

3. **Setup MinIO (optional)**
   ```bash
   # Download and run MinIO
   wget https://dl.min.io/server/minio/release/linux-amd64/minio
   chmod +x minio
   ./minio server /tmp/minio --console-address ":9001"
   
   # Or use Docker
   docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
   ```

4. **Run the application**
   ```bash
   # Option 1: Use the startup script
   ./start.sh
   
   # Option 2: Run directly
   go run cmd/main.go
   
   # Option 3: Build and run
   go build cmd/main.go
   ./main
   ```

## Default Configuration

- **Database**: `localhost:5432/1337b04rd`
- **MinIO**: `localhost:9000`
- **Server**: `localhost:8080`
- **API Base**: `http://localhost:8080/api`

## Test the API

```bash
# Health check
curl http://localhost:8080/health

# Create a post
curl -X POST http://localhost:8080/api/posts \
  -F "title=Test Post" \
  -F "content=This is a test post" \
  -F "author_id=user123" \
  -F "author_name=Test User"

# Get all posts
curl http://localhost:8080/api/posts
```

## Environment Variables

Copy `env.example` to `.env` and modify as needed:

```bash
cp env.example .env
# Edit .env with your configuration
```

## Troubleshooting

- **Database connection failed**: Ensure PostgreSQL is running and accessible
- **MinIO connection failed**: Image uploads won't work, but the app will run
- **Port already in use**: Change `SERVER_PORT` in environment variables
