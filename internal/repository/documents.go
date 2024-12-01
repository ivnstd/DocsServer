package repository

import (
	"bytes"
	"context"
	"docs_server/internal/models"
	"errors"
	"io"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocumentsDB struct {
	collection *mongo.Collection
}

func NewDocumentsDB(db *mongo.Database) *DocumentsDB {
	return &DocumentsDB{collection: db.Collection("documents")}
}

// Создание нового документа
func (r *DocumentsDB) CreateDocument(ctx context.Context, doc *models.Document) error {
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		logrus.Errorf("CreateDocument: Failed to execute query: %v", err)
		return err
	}

	logrus.Infof("CreateDocument: Created document with id: %v", doc.ID)
	return nil
}

// Создание нового файла
func (r *DocumentsDB) CreateFile(ctx context.Context, id primitive.ObjectID, file io.Reader) error {
	bucket, err := gridfs.NewBucket(r.collection.Database())
	if err != nil {
		logrus.Errorf("CreateFile: Failed to create GridFS bucket: %v", err)
		return err
	}

	err = bucket.UploadFromStreamWithID(id, id.Hex(), file)
	if err != nil {
		logrus.Errorf("CreateFile: Failed to execute query: %v", err)
		return err
	}

	logrus.Infof("CreateFile: Created file with id: %v", id)
	return nil
}

// Получение списка документов
func (r *DocumentsDB) GetDocumentsList(ctx context.Context, userID primitive.ObjectID, login, key, value string, limit int64) ([]models.Document, error) {
	filter := bson.M{}

	// Фильтр по логину или личным документам
	if login != "" {
		filter["$or"] = []bson.M{
			{"grant": login},
			{"public": true},
		}
	} else {
		filter["user_id"] = userID
	}

	// Фильтр по key и value
	if key != "" && value != "" {
		filter[key] = value
	}

	// Опции сортировки и лимита
	opts := options.Find().
		SetSort(bson.D{
			{Key: "name", Value: 1},
			{Key: "created", Value: 1},
		}).
		SetLimit(limit)

	// Запрос в базу данных
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logrus.Errorf("GetDocumentsList: Failed to execute query: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []models.Document
	if err := cursor.All(ctx, &documents); err != nil {
		logrus.Errorf("GetDocumentsList: Failed to decode documents: %v", err)
		return nil, err
	}

	logrus.Infof("GetDocumentsList: Found %d documents", len(documents))
	return documents, nil
}

func (r *DocumentsDB) GetDocument(ctx context.Context, id, userID primitive.ObjectID) (*models.Document, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}
	var document models.Document
	err := r.collection.FindOne(ctx, filter).Decode(&document)

	if err == mongo.ErrNoDocuments {
		logrus.Errorf("GetDocument: Document whith id: %v not found", id)
		return nil, errors.New("not found")
	} else if err != nil {
		logrus.Errorf("GetDocument: Failed to execute query: %v", err)
		return nil, err
	}

	logrus.Infof("GetDocument: Found document with id: %v", id)
	return &document, nil
}

// Получение одного документа
func (r *DocumentsDB) GetFile(ctx context.Context, id primitive.ObjectID) ([]byte, error) {
	bucket, err := gridfs.NewBucket(r.collection.Database())
	if err != nil {
		logrus.Errorf("GetFile: Failed to create GridFS bucket: %v", err)
		return nil, err
	}

	// Загрузка файла
	var buf bytes.Buffer
	_, err = bucket.DownloadToStream(id, &buf)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("not found")
	} else if err != nil {
		logrus.Errorf("GetFile: Failed to download file: %v", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

// Удаление документа
func (r *DocumentsDB) DeleteDocument(ctx context.Context, id, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		logrus.Errorf("DeleteDocument: Failed to execute query: %v", err)
		return false, err
	}

	if result.DeletedCount == 0 {
		logrus.Info("DeleteDocument: document not found")
		return false, nil
	}

	// Удаление файла, если есть
	fileID := id
	err = r.deleteFile(fileID)
	if err != nil {
		logrus.Errorf("DeleteDocument: Failed to delete document: %v", err)
		return false, errors.New("failed to delete document")
	}

	return true, nil
}

// Удаление файла
func (r *DocumentsDB) deleteFile(id primitive.ObjectID) error {
	bucket, err := gridfs.NewBucket(r.collection.Database())
	if err != nil {
		logrus.Errorf("DeleteFile: Failed to create GridFS bucket: %v", err)
		return err
	}

	if err := bucket.Delete(id); err != nil {
		logrus.Errorf("DeleteFile: Failed to delete file: %v", err)
		return errors.New("failed to delete file")
	}

	return nil
}
