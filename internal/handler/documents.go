package handler

import (
	"docs_server/internal/models"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Загрузка нового документа
func (h *Handler) CreateDocument(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		newResponse(c, http.StatusBadRequest, ErrBadRequest, nil, nil)
		return
	}

	// Получение метаданных
	meta := c.PostForm("meta")
	var metadata models.DocumentMeta
	if err := json.Unmarshal([]byte(meta), &metadata); err != nil {
		newResponse(c, http.StatusBadRequest, &ErrorResponse{400, "Invalid meta"}, nil, nil)
		return
	}

	// Проверка токена
	user, err := h.services.Users.CheckAuth(c.Request.Context(), metadata.Token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, ErrUnauthorized, nil, nil)
		return
	}

	// Получение JSON (опционально)
	jsonData := c.PostForm("json")
	var decodedJSON map[string]interface{}
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &decodedJSON); err != nil {
			newResponse(c, http.StatusBadRequest, &ErrorResponse{400, "Invalid json"}, nil, nil)
			return
		}
	}

	// Получение файла
	var fileHeader *multipart.FileHeader
	if files := form.File["file"]; len(files) > 0 {
		fileHeader = files[0]
	}

	// Создание документа
	document, err := h.services.Documents.CreateDocument(c.Request.Context(), user.ID, metadata, decodedJSON, fileHeader)
	if err != nil {
		if strings.Contains(err.Error(), "open file") {
			newResponse(c, http.StatusBadRequest, &ErrorResponse{400, "Failed open file"}, nil, nil)
		} else {
			newResponse(c, http.StatusInternalServerError, ErrInternalServer, nil, nil)
		}
		return
	}

	newResponse(c, http.StatusCreated, nil, nil, gin.H{
		"json": decodedJSON,
		"file": document.Name,
	})
}

// Получение списка документов
func (h *Handler) GetDocumentsList(c *gin.Context) {
	token := c.Query("token")
	login := c.Query("login")
	key := c.Query("key")
	value := c.Query("value")
	limitParam := c.Query("limit")

	// Проверка токена
	user, err := h.services.Users.CheckAuth(c.Request.Context(), token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, ErrUnauthorized, nil, nil)
		return
	}

	limit, err := strconv.ParseInt(limitParam, 10, 64)
	if err != nil {
		limit = 100
	}

	documents, err := h.services.Documents.GetDocumentsList(c.Request.Context(), user.ID, login, key, value, limit)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, ErrInternalServer, nil, nil)
		return
	}

	newResponse(c, http.StatusOK, nil, nil, gin.H{"docs": documents})
}

// Получение одного документа
func (h *Handler) GetDocument(c *gin.Context) {
	idParam := c.Param("id")
	token := c.Query("token")

	// Проверка токена
	user, err := h.services.Users.CheckAuth(c.Request.Context(), token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, ErrUnauthorized, nil, nil)
		return
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		newResponse(c, http.StatusBadRequest, &ErrorResponse{400, "Invalid document ID"}, nil, nil)
	}

	document, fileData, err := h.services.Documents.GetDocument(c.Request.Context(), id, user.ID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			newResponse(c, http.StatusNotFound, ErrNotFound, nil, nil)
		case strings.Contains(err.Error(), "forbidden"):
			newResponse(c, http.StatusForbidden, ErrForbidden, nil, nil)
		default:
			newResponse(c, http.StatusInternalServerError, ErrInternalServer, nil, nil)
		}
		return
	}

	// Выдача файла с нужным mime, если есть
	if fileData != nil {
		if c.Request.Method == http.MethodHead {
			c.Status(http.StatusOK)
		} else {
			c.Data(http.StatusOK, document.Mime, fileData)
		}
		return
	}

	// Выдача JSON, если файла нет
	newResponse(c, http.StatusOK, nil, nil, document.JSON)
}

// Удаление документа
func (h *Handler) DeleteDocument(c *gin.Context) {
	token := c.Query("token")
	idParam := c.Param("id")

	// Проверка токена
	user, err := h.services.Users.CheckAuth(c.Request.Context(), token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, ErrUnauthorized, nil, nil)
		return
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		newResponse(c, http.StatusBadRequest, &ErrorResponse{400, "Invalid document ID"}, nil, nil)
	}

	// Вызов сервиса для удаления документа
	success, err := h.services.Documents.DeleteDocument(c.Request.Context(), id, user.ID)
	if err != nil {
		newResponse(c, http.StatusBadRequest, ErrBadRequest, nil, nil)
		return
	}

	// Возврат ответа
	newResponse(c, http.StatusOK, nil, gin.H{token: success}, nil)
}
