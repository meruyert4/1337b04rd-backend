package postgres

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func ConnectToDB(connectionString string) *sql.DB {
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connectionString)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to PostgreSQL database")
				break
			}
		}
		log.Printf("Waiting for database to be ready... Try: %d", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}

func InitDB(db *sql.DB) error {
	// Create tables if they don't exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			author_id VARCHAR(255) NOT NULL,
			author_name VARCHAR(255) NOT NULL,
			image_url TEXT,
			is_archive BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			author_id VARCHAR(255) NOT NULL,
			author_name VARCHAR(255) NOT NULL,
			image_url TEXT,
			reply_to_comment_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			gender VARCHAR(50),
			age VARCHAR(50),
			image TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_is_archive ON posts(is_archive)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	// Run migration to update existing sessions table if needed
	migrationQueries := []string{
		`ALTER TABLE sessions ADD COLUMN IF NOT EXISTS gender VARCHAR(50)`,
		`ALTER TABLE sessions ADD COLUMN IF NOT EXISTS age VARCHAR(50)`,
		`UPDATE sessions SET gender = 'Unknown' WHERE gender IS NULL`,
		`UPDATE sessions SET age = 'Unknown' WHERE age IS NULL`,
	}

	for _, query := range migrationQueries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Migration query failed: %v", err)
			// Continue with other migrations even if one fails
		}
	}

	log.Println("Database tables initialized successfully")
	return nil
}
