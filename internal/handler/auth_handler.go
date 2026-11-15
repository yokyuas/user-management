package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"user-management/internal/domain"
	"user-management/internal/service"
)

type AuthHandler struct {
	svc *service.UserService
}

func NewAuthHandler(s *service.UserService) *AuthHandler {
	return &AuthHandler{svc: s}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input domain.CreateUserInput
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	user, err := h.svc.Register(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	token, err := h.svc.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, gin.H{"token": token, "token_type": "Bearer"})
}
