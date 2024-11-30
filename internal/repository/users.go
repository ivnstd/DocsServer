package repository

import (
	"context"
	"docs_server/internal/models"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UsersDB struct {
	userCollection    *mongo.Collection
	sessionCollection *mongo.Collection
}

func NewUsersDB(db *mongo.Database) *UsersDB {
	return &UsersDB{
		userCollection:    db.Collection("users"),
		sessionCollection: db.Collection("sessions"),
	}
}

// Создание нового пользователя
func (r *UsersDB) CreateUser(ctx context.Context, user models.User) error {
	// Проверка на занятость логина
	filter := bson.M{"login": user.Login}
	exists, err := r.userCollection.CountDocuments(ctx, filter)
	if err != nil {
		logrus.Errorf("CreateUser: failed to check existing user: %v", err)
		return err
	}
	if exists > 0 {
		err = errors.New("user with login already exists")
		logrus.Errorf("CreateUser: %v", err)
		return err
	}

	_, err = r.userCollection.InsertOne(ctx, user)
	if err != nil {
		logrus.Errorf("CreateUser: failed create user: %v", err)
		return err
	}

	return nil
}

// Получение пользователя по логину
func (r *UsersDB) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.userCollection.FindOne(ctx, bson.M{"login": login}).Decode(&user)
	if err != nil {
		logrus.Errorf("GetUserByLogin: failed find user by login: %v", err)
		return nil, err
	}
	return &user, nil
}

// Получение пользователя по ID
func (r *UsersDB) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		logrus.Errorf("GetUserByID: failed find user by ID: %v", err)
		return nil, err
	}
	return &user, nil
}

// Создание новой сессии
func (r *UsersDB) CreateSession(ctx context.Context, session models.Session) (*primitive.ObjectID, error) {
	result, err := r.sessionCollection.InsertOne(ctx, session)
	if err != nil {
		logrus.Errorf("CreateSession: failed create session: %v", err)
		return nil, err
	}

	// Извлечение ID сессии после добавления в базу
	sessionID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		err = errors.New("failed to convert inserted ID to ObjectID")
		logrus.Errorf("CreateSession: %v", err)
		return nil, err
	}

	return &sessionID, nil
}

// Получение сессии по токену
func (r *UsersDB) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	filter := bson.M{"token": token}

	err := r.sessionCollection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		logrus.Errorf("GetSessionByToken: failed find session by token: %v", err)
		return nil, err
	}

	return &session, nil
}

// Удаление сессии по токену
func (r *UsersDB) DeleteSessionByToken(ctx context.Context, token string) error {
	_, err := r.sessionCollection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// Очистка истекших сессий
func (r *UsersDB) RemoveExpiredSessions(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		_, err := r.sessionCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": time.Now()}})
		if err != nil {
			logrus.Errorf("Error deleting expired sessions: %v", err)
		} else {
			logrus.Info("Expired sessions removed")
		}
	}
}
