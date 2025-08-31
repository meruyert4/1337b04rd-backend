# 1337b04rd Backend

A robust Go backend service for a social media platform with posts, comments, and user sessions.

## 🚀 Features

- **Posts Management**: Create, read, update, delete, archive, and unarchive posts
- **Comments System**: Full CRUD operations for comments with reply support
- **User Sessions**: Session management with expiration handling
- **Image Storage**: MinIO integration for image uploads and serving
- **CORS Support**: Configurable CORS for frontend integration
- **Database**: PostgreSQL with automatic table creation
- **Logging**: Structured logging with file output

## 🔧 Recent Fixes & Improvements

### 1. Postman Collection Consolidation
- ✅ Consolidated all API endpoints into a single, comprehensive Postman collection
- ✅ Fixed incorrect URLs and request formats
- ✅ Added proper request headers and body examples
- ✅ Organized endpoints by functionality (Posts, Comments, Sessions, Images)

### 2. CORS Configuration
- ✅ Enhanced CORS middleware with environment-based configuration
- ✅ Added support for configurable allowed origins via `ALLOWED_ORIGIN` environment variable
- ✅ Improved CORS headers including `Access-Control-Max-Age`
- ✅ Consistent CORS handling across all endpoints

### 3. Handler Fixes
- ✅ Fixed URL parameter extraction inconsistencies
- ✅ Updated handlers to use path parameters instead of query parameters where appropriate
- ✅ Added proper error handling and validation
- ✅ Fixed mux import issues in handlers

### 4. API Endpoint Corrections
- ✅ Fixed Rick and Morty API client URL construction
- ✅ Corrected image serving handlers
- ✅ Improved error responses and status codes

## 📋 API Endpoints

### Posts
- `POST /api/posts` - Create a new post
- `GET /api/posts` - Get all posts (with pagination and archive filtering)
- `GET /api/posts/{id}` - Get a specific post
- `PUT /api/posts/{id}` - Update a post
- `DELETE /api/posts/{id}` - Delete a post
- `POST /api/posts/{id}/archive` - Archive a post
- `POST /api/posts/{id}/unarchive` - Unarchive a post
- `GET /api/posts/author` - Get posts by author

### Comments
- `POST /api/comments` - Create a new comment
- `GET /api/comments/{id}` - Get a specific comment
- `PUT /api/comments/{id}` - Update a comment
- `DELETE /api/comments/{id}` - Delete a comment
- `GET /api/comments/post` - Get comments by post

### Sessions
- `POST /api/sessions` - Create a new session
- `GET /api/sessions/{id}` - Get a specific session
- `PUT /api/sessions/{id}` - Update a session
- `DELETE /api/sessions/{id}` - Delete a session
- `POST /api/sessions/cleanup` - Cleanup expired sessions

### Images
- `GET /images/posts/{filename}` - Serve post images
- `GET /images/comments/{filename}` - Serve comment images

### Health Check
- `GET /health` - Health check endpoint

## 🛠️ Setup & Configuration

### Prerequisites
- Go 1.23.0 or higher
- PostgreSQL
- MinIO (optional, for image storage)

### Environment Variables
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=1337b04rd
DB_SSLMODE=disable

# MinIO Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# CORS Configuration
ALLOWED_ORIGIN=http://localhost:3000

# Logging
LOG_LEVEL=info
```

### Quick Start
1. Copy `env.example` to `.env` and configure your environment
2. Ensure PostgreSQL is running
3. Run the startup script: `./start.sh`
4. Or build and run manually:
   ```bash
   go build -o main cmd/main.go
   ./main
   ```

## 📚 Postman Collection

Import the `1337b04rd.postman_collection.json` file into Postman to test all API endpoints. The collection includes:

- Pre-configured environment variables
- Example request bodies for all endpoints
- Proper headers and authentication setup
- Organized folder structure by functionality

## 🔍 Testing

### Manual Testing
Use the provided Postman collection to test all endpoints before connecting to the frontend.

### Health Check
```bash
curl http://localhost:8080/health
```

### Create a Post
```bash
curl -X POST http://localhost:8080/api/posts \
  -F "title=Test Post" \
  -F "content=This is a test post" \
  -F "author_id=user123" \
  -F "author_name=Test User"
```

## 🐛 Known Issues & Solutions

### CORS Issues
- Ensure `ALLOWED_ORIGIN` environment variable is set correctly
- Check that frontend origin matches the configured allowed origin

### Database Connection
- Verify PostgreSQL is running and accessible
- Check database credentials in environment variables

### Image Upload Issues
- Ensure MinIO is running (optional, but required for image functionality)
- Check MinIO endpoint and credentials

## 📝 Development Notes

- All handlers now use path parameters consistently
- CORS is properly configured for frontend integration
- Error handling has been improved across all endpoints
- Logging is configured to output to both console and file

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly using the Postman collection
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.
