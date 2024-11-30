package handler

import (
	"docs_server/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Аутентификация пользователя
func (h *Handler) Auth(c *gin.Context) {
	var loginRequest models.AuthInput
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		newResponse(c, http.StatusBadRequest, ErrBadRequest, nil, nil)
		return
	}

	token, err := h.services.Users.Auth(c.Request.Context(), loginRequest.Login, loginRequest.Password)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, &ErrorResponse{401, "Failed to identify User"}, nil, nil)
		return
	}

	newResponse(c, http.StatusOK, nil, gin.H{"token": token}, nil)
}

// Завершение авторизованной сессии
func (h *Handler) EndSession(c *gin.Context) {
	token := c.Param("token")

	if err := h.services.Users.EndSession(c.Request.Context(), token); err != nil {
		newResponse(c, http.StatusInternalServerError, ErrInternalServer, nil, nil)
		return
	}

	newResponse(c, http.StatusOK, nil, gin.H{token: true}, nil)
}
