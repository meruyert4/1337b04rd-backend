package app

import (
	"1337b04rd/config"
	"1337b04rd/internal/adapters/externalapi"
	"1337b04rd/internal/adapters/handler"
	"1337b04rd/internal/adapters/repository"
	"1337b04rd/internal/adapters/storage"
	"1337b04rd/internal/service"
	"1337b04rd/pkg/postgres"
	"database/sql"
)

type App struct {
	DB               *sql.DB
	Storage          *storage.MinioClient
	PostService      *service.PostService
	CommentService   *service.CommentService
	SessionService   *service.SessionService
	PostHandler      *handler.PostHandler
	CommentHandler   *handler.CommentHandler
	SessionHandler   *handler.SessionHandler
	CharacterHandler *handler.CharacterHandler
}

func NewApp(cfg *config.Config) (*App, error) {
	// Initialize database
	db := postgres.ConnectToDB(cfg.GetDBConnectionString())
	if err := postgres.InitDB(db); err != nil {
		return nil, err
	}

	// Initialize MinIO storage
	storageClient, err := storage.NewMinioClient(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		return nil, err
	}

	// Initialize Rick and Morty client
	rickAndMortyClient := externalapi.NewRickAndMortyClient()

	// Initialize repositories
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Initialize services
	postService := service.NewPostService(postRepo, commentRepo, storageClient)
	commentService := service.NewCommentService(commentRepo, storageClient)
	sessionService := service.NewSessionService(sessionRepo, storageClient)

	// Initialize handlers
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	characterHandler := handler.NewCharacterHandler(rickAndMortyClient)

	return &App{
		DB:               db,
		Storage:          storageClient,
		PostService:      postService,
		CommentService:   commentService,
		SessionService:   sessionService,
		PostHandler:      postHandler,
		CommentHandler:   commentHandler,
		SessionHandler:   sessionHandler,
		CharacterHandler: characterHandler,
	}, nil
}

func (a *App) Close() error {
	return a.DB.Close()
}
