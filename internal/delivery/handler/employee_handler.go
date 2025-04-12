package handlers

import (
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EmployerHandler struct {
	usecase *usecase.EmployerUsecase
}
func NewEmployerHandler(router *gin.RouterGroup, uc *usecase.EmployerUsecase) {
	handler := &EmployerHandler{usecase: uc}

	employer := router.Group("/employer")
	{
		employer.PUT("/profile/update", handler.UpdateProfile)
		employer.GET("/profile", handler.GetProfile)
		employer.POST("/signup", handler.Signup)
		employer.POST("/login", handler.Login)
	}
}

func (h *EmployerHandler) UpdateProfile(c *gin.Context) {
	var profile model.UpdateEmployerInput
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.usecase.UpdateProfile(c.Request.Context(), &profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employer profile updated successfully"})
}

func (h *EmployerHandler) GetProfile(c *gin.Context) {
	userID := c.Param("user_id")

	profile, err := h.usecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
func (h *EmployerHandler) Signup(c *gin.Context) {
	var req model.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"+err.Error()})
		return
	}

	res, err := h.usecase.Signup(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}
func (h *EmployerHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input "+err.Error()})
		return
	}

	res, err := h.usecase.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}