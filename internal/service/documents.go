package service

import (
	"context"
	"docs_server/internal/models"
	"docs_server/internal/repository"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocumentService struct {
	repo  repository.Documents
	cache *cache.Cache
}

func NewDocumentService(repo repository.Documents) *DocumentService {
	return &DocumentService{
		repo:  repo,
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// Загрузка нового документа
func (s *DocumentService) CreateDocument(ctx context.Context, userID primitive.ObjectID, meta models.DocumentMeta, jsonData map[string]interface{}, fileHeader *multipart.FileHeader) (*models.Document, error) {
	id := primitive.NewObjectID()

	// Сохранение файла, если есть
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			logrus.Errorf("CreateDocument: failed to open file: %v", err)
			return nil, errors.New("failed open file")
		}
		defer file.Close()

		err = s.repo.CreateFile(ctx, id, file)
		if err != nil {
			logrus.Errorf("CreateDocument: failed to save file with ID %s: %v", id.Hex(), err)
			return nil, errors.New("failed save file")
		} else {
			// Кэширование
			cacheFileKey := fmt.Sprintf("file_%v_%v", id.Hex(), userID.Hex())
			s.cache.Set(cacheFileKey, file, cache.DefaultExpiration)
		}
	}

	// Создание документа
	document := &models.Document{
		ID:      id,
		UserID:  userID,
		Name:    meta.Name,
		Mime:    meta.Mime,
		File:    meta.File,
		Public:  meta.Public,
		Created: time.Now(),
		Grant:   meta.Grant,
		JSON:    jsonData,
	}

	// Сохранение документа
	if err := s.repo.CreateDocument(ctx, document); err != nil {
		return nil, err
	}

	// Кэширование
	cacheDocKey := fmt.Sprintf("document_%v_%v", id.Hex(), userID.Hex())
	s.cache.Set(cacheDocKey, document, cache.DefaultExpiration)

	// Удаление всех кэшированных списков
	s.clearCacheKeys()

	return document, nil
}

// Получение списка документов
func (s *DocumentService) GetDocumentsList(ctx context.Context, userID primitive.ObjectID, login, key, value string, limit int64) ([]models.Document, error) {
	if limit <= 0 {
		limit = 100
	}

	// Проверка наличия списка в кэше
	cacheKey := fmt.Sprintf("list_%s_%s_%s_%d_%v", login, key, value, limit, userID.Hex())
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		logrus.Infof("GetDocumentsList: retrieved from cache: %v", cacheKey)
		return cachedData.([]models.Document), nil
	}

	// Получение списка документов
	documents, err := s.repo.GetDocumentsList(ctx, userID, login, key, value, limit)
	if err != nil {
		return nil, err
	}

	// Кэширование и обновление индекса ключей списков
	s.cache.Set(cacheKey, documents, cache.DefaultExpiration)
	s.updateCacheKeys(cacheKey)

	return documents, nil
}

// Получение одного документа
func (s *DocumentService) GetDocument(ctx context.Context, id, userID primitive.ObjectID) (*models.Document, []byte, error) {
	cacheKey := fmt.Sprintf("document_%v_%v", id.Hex(), userID.Hex())
	cacheFileKey := fmt.Sprintf("file_%v_%v", id.Hex(), userID.Hex())

	// Проверка на наличие документа в кэше
	cachedData, found := s.cache.Get(cacheKey)
	if found {
		cachedDoc, ok := cachedData.(models.Document)
		if ok {
			if cachedDoc.File {
				// Проверка на наличие файла в кэше
				cachedFileData, found := s.cache.Get(cacheFileKey)
				if found {
					logrus.Infof("GetDocument (with file): retrieved from cache: %v", cacheKey)
					return &cachedDoc, cachedFileData.([]byte), nil
				}
			} else {
				logrus.Infof("GetDocument: retrieved from cache: %v", cacheKey)
				return &cachedDoc, nil, nil
			}
		}
	}

	// Получение документа
	document, err := s.repo.GetDocument(ctx, id, userID)
	if err != nil {
		return nil, nil, err
	}
	// Кэширование
	s.cache.Set(cacheKey, *document, cache.DefaultExpiration)

	// Получение файла, если есть
	if document.File {
		fileData, err := s.repo.GetFile(ctx, document.ID)
		if err != nil {
			return nil, nil, err
		}
		// Кэширование
		s.cache.Set(cacheFileKey, fileData, cache.DefaultExpiration)
		return document, fileData, nil
	}

	return document, nil, nil
}

// Удаление документа
func (s *DocumentService) DeleteDocument(ctx context.Context, id, userID primitive.ObjectID) (bool, error) {
	success, err := s.repo.DeleteDocument(ctx, id, userID)
	if err != nil {
		return false, err
	}

	// Удаляем кэшированный документ
	cacheDocKey := fmt.Sprintf("document_%v_%v", id.Hex(), userID.Hex())
	s.cache.Delete(cacheDocKey)
	s.clearCacheKeys()

	return success, nil
}

// Очистка кэшированных списков
func (s *DocumentService) updateCacheKeys(newKey string) {
	keysKey := "cache_list_keys"
	cachedKeys, found := s.cache.Get(keysKey)
	if !found {
		cachedKeys = make(map[string]struct{})
	}
	keys := cachedKeys.(map[string]struct{})
	keys[newKey] = struct{}{}
	s.cache.Set(keysKey, keys, cache.NoExpiration)
}

// Обновление индекса кэшированных ключей
func (s *DocumentService) clearCacheKeys() {
	keysKey := "cache_list_keys"
	cachedKeys, found := s.cache.Get(keysKey)
	if found {
		keys := cachedKeys.(map[string]struct{})
		for key := range keys {
			s.cache.Delete(key)
		}
		s.cache.Delete(keysKey)
	}
}
