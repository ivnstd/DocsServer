package service

import (
	"context"
	"crypto/rand"
	"docs_server/internal/models"
	"docs_server/internal/repository"
	"encoding/hex"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo       repository.Users
	sessionTTL time.Duration
}

func NewUsersService(repo repository.Users) *UserService {
	return &UserService{
		repo:       repo,
		sessionTTL: time.Duration(24 * time.Hour),
	}
}

// Создание нового пользователя
func (s *UserService) CreateUser(ctx context.Context, login, password string) error {
	// Хеширование пароля
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		Login:        login,
		PasswordHash: hashedPassword,
	}
	return s.repo.CreateUser(ctx, user)
}

// Хэширование пароля
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Аутентификация пользователя, выдача токена
func (s *UserService) Auth(ctx context.Context, login, pswd string) (string, error) {
	session, err := s.CreateSession(ctx, login, pswd)
	if err != nil {
		return "", err
	}
	return session.Token, nil
}

// Создание новой сессии
func (s *UserService) CreateSession(ctx context.Context, login, pswd string) (*models.Session, error) {
	// Получение пользователя по логину
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, errors.New("invalid login or password")
	}

	// Проверка пароля по хэшу
	if !checkPasswordHash(pswd, user.PasswordHash) {
		return nil, errors.New("invalid login or password")
	}

	// Создание сессии в базе
	session := models.Session{
		UserID:    user.ID,
		Token:     generateToken(),
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}
	sessionID, err := s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}
	session.ID = *sessionID

	return &session, nil
}

// Сравнение пароля и хэша
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Генерация уникального токена сессии
func generateToken() string {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		panic("failed to generate token: " + err.Error())
	}

	return hex.EncodeToString(tokenBytes)
}

// Завершение авторизованной сессии
func (s *UserService) EndSession(ctx context.Context, token string) error {
	err := s.repo.DeleteSessionByToken(ctx, token)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) CheckAuth(ctx context.Context, token string) (*models.User, error) {
	session, err := s.repo.GetSessionByToken(ctx, token)
	if err != nil || session == nil {
		logrus.Errorf("CheckAuth: invalid or expired token: %v", err)
		logrus.Errorf("CheckAuth: invalid or expired token: %v", session)
		return nil, errors.New("invalid or expired token")
	}

	if !session.ExpiresAt.After(time.Now()) {
		logrus.Errorf("CheckAuth: expired token")
		s.repo.DeleteSessionByToken(ctx, token)
		return nil, errors.New("expired token")
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		logrus.Errorf("CheckAuth: user not found: %v", err)
		return nil, err
	}
	return user, err
}
