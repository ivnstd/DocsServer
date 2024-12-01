package handler

import (
	"docs_server/internal/configs"
	"docs_server/internal/models"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// Валидность логина
func isValidLogin(login string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9]{8,}$`)
	return re.MatchString(login)
}

// Валидность пароля
func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[^\w\s]`).MatchString(password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}

// Регистрация пользователя
func (h *Handler) Register(c *gin.Context) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, ErrBadRequest, nil, nil)
		return
	}

	// Проверка токена администратора
	if input.Token != configs.Config.AdminToken {
		newResponse(c, http.StatusForbidden, &ErrorResponse{403, "Invalid admin token"}, nil, nil)
		return
	}

	// Валидация логина и пароля
	if !isValidLogin(input.Login) {
		newResponse(c, http.StatusBadRequest, &ErrorResponse{400, "Invalid login"}, nil, nil)
		return
	}
	if !isValidPassword(input.Password) {
		newResponse(c, http.StatusBadRequest, &ErrorResponse{400, "Invalid password"}, nil, nil)
		return
	}

	// Создание пользователя
	if err := h.services.Users.CreateUser(c.Request.Context(), input.Login, input.Password); err != nil {
		if strings.Contains(err.Error(), "login already exists") {
			newResponse(c, http.StatusConflict, &ErrorResponse{409, "User with this login already exists"}, nil, nil)
		} else {
			newResponse(c, http.StatusInternalServerError, ErrInternalServer, nil, nil)
		}
		return
	}

	newResponse(c, http.StatusOK, nil, map[string]string{"login": input.Login}, nil)
}
