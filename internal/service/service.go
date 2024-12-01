package service

import (
	"context"
	"docs_server/internal/models"
	"docs_server/internal/repository"
	"mime/multipart"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Users interface {
	CreateUser(ctx context.Context, login, password string) error
	Auth(ctx context.Context, login, pswd string) (string, error)
	CreateSession(ctx context.Context, login, password string) (*models.Session, error)
	CheckAuth(ctx context.Context, token string) (*models.User, error)
	EndSession(ctx context.Context, token string) error
}

type Documents interface {
	CreateDocument(ctx context.Context, userID primitive.ObjectID, meta models.DocumentMeta, jsonData map[string]interface{}, fileHeader *multipart.FileHeader) (*models.Document, error)
	GetDocumentsList(ctx context.Context, userID primitive.ObjectID, login, key, value string, limit int64) ([]models.Document, error)
	GetDocument(ctx context.Context, id, userID primitive.ObjectID) (*models.Document, []byte, error)
	DeleteDocument(ctx context.Context, id, userID primitive.ObjectID) (bool, error)
}

type Service struct {
	Users
	Documents
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Users:     NewUsersService(repos.Users),
		Documents: NewDocumentService(repos.Documents),
	}
}
