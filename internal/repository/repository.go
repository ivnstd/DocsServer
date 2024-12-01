package repository

import (
	"context"
	"docs_server/internal/models"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)

	CreateSession(ctx context.Context, session models.Session) (*primitive.ObjectID, error)
	GetSessionByToken(ctx context.Context, token string) (*models.Session, error)
	DeleteSessionByToken(ctx context.Context, token string) error
	RemoveExpiredSessions(ctx context.Context)
}

type Documents interface {
	CreateDocument(ctx context.Context, doc *models.Document) error
	CreateFile(ctx context.Context, id primitive.ObjectID, file io.Reader) error
	GetDocumentsList(ctx context.Context, userID primitive.ObjectID, login, key, value string, limit int64) ([]models.Document, error)
	GetDocument(ctx context.Context, id, userID primitive.ObjectID) (*models.Document, error)
	GetFile(ctx context.Context, id primitive.ObjectID) ([]byte, error)
	DeleteDocument(ctx context.Context, id, userID primitive.ObjectID) (bool, error)
}

type Repository struct {
	Users
	Documents
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		Users:     NewUsersDB(db),
		Documents: NewDocumentsDB(db),
	}
}
