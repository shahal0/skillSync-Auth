package handlers

import (
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CandidateHandler struct {
	usecase *usecase.CandidateUsecase
}

func NewCandidateHandler(router *gin.RouterGroup, uc *usecase.CandidateUsecase) {
	handler := &CandidateHandler{usecase: uc}

	candidate := router.Group("/candidate")
	{
		candidate.PUT("/profile/update", handler.UpdateProfile)
		candidate.GET("/profile", handler.GetProfile)
		candidate.PUT("/Skills/update", handler.UpdateSkills)
		candidate.PUT("/Education/update", handler.UpdateEducation)
		candidate.POST("/signup", handler.Signup)
		candidate.POST("/login", handler.Login)
	}
}

func (h *CandidateHandler) UpdateProfile(c *gin.Context) {
	var profile model.UpdateCandidateInput
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.usecase.UpdateCandidateProfile(c.Request.Context(),&profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidate profile updated successfully"})
}

func (h *CandidateHandler) GetProfile(c *gin.Context) {
	userID := c.Param("user_id")

	profile, err := h.usecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
func(h *CandidateHandler) UpdateSkills(c *gin.Context) {
	var skills model.Skills
	if err := c.ShouldBindJSON(&skills); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.usecase.AddSkills(c.Request.Context(), skills)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidate skills updated successfully"})
}
func (h *CandidateHandler)UpdateEducation(c *gin.Context) {
	var education model.Education
	if err := c.ShouldBindJSON(&education); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.usecase.AddEducation(c.Request.Context(), education)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update education"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidate education updated successfully"})
}
func (h *CandidateHandler) Signup(c *gin.Context) {
	var req model.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	res, err := h.usecase.Signup(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}
func (h *CandidateHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}