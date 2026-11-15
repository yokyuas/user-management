package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"user-management/internal/domain"
	"user-management/internal/repository"
	"user-management/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{svc: s}
}

func (h *UserHandler) Get(c *gin.Context) {
	user, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, user)
}
func (h *UserHandler) List(c *gin.Context) {
	user, err := h.svc.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, user)
}
func (h *UserHandler) Update(c *gin.Context) {
	var in domain.UpdateUserInput
	if c.ShouldBindJSON(&in) != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	user, err := h.svc.Update(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(200, user)
}
func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func writeError(c *gin.Context, err error) {
	switch err {
	case service.ErrInvalidInput:
		c.JSON(400, gin.H{"error": err.Error()})
	case service.ErrEmailExists:
		c.JSON(409, gin.H{"error": err.Error()})
	case service.ErrInvalidCredentials:
		c.JSON(401, gin.H{"error": "invalid credentials"})
	case repository.ErrNotFound:
		c.JSON(404, gin.H{"error": err.Error()})
	default:
		c.JSON(500, gin.H{"error": "internal server error"})
	}
}
