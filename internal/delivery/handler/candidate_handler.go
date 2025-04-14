package handlers

import (
	//"go/token"
	"log"
	"net/http"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/internal/usecase"
	"skillsync-authservice/pkg"

	"github.com/gin-gonic/gin"
)

type CandidateHandler struct {
	usecase   *usecase.CandidateUsecase
	jwthelper *pkg.JWTMaker
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
		candidate.POST("/verify-email", handler.VerifyEmail)
		candidate.POST("/resend-otp", handler.ResendOtp)
		candidate.POST("/forgot-password", handler.ForgotPassword)
		candidate.PUT("/reset-password", handler.ResetPassword)
		candidate.PATCH("/change-password", handler.ChangePassword)
	}
}

func (h *CandidateHandler) UpdateProfile(c *gin.Context) {
	// Get the Authorization header
	authHeader := c.Request.Header.Get("Authorization")
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Parse the request body
	var profile model.UpdateCandidateInput
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the usecase to update the candidate profile
	err = h.usecase.UpdateCandidateProfile(c.Request.Context(), &profile, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidate profile updated successfully"})
}

func (h *CandidateHandler) GetProfile(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	log.Println("Authorization Header:", authHeader)

	// Extract the token
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		log.Println("Token Extraction Error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	profile, err := h.usecase.GetProfile(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
func (h *CandidateHandler) UpdateSkills(c *gin.Context) {
	// Extract the token from the Authorization header
	authHeader := c.Request.Header.Get("Authorization")
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Extract user ID from the token
	userID, err := h.usecase.ExtractUserIDFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Parse the request body
	var skills model.Skills
	if err := c.ShouldBindJSON(&skills); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Set the CandidateID in the skills object
	skills.CandidateID = userID

	// Call the usecase to add the skills
	err = h.usecase.AddSkills(c.Request.Context(), skills)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidate skills updated successfully"})
}
func (h *CandidateHandler) UpdateEducation(c *gin.Context) {
	// Extract the token from the Authorization header
	authHeader := c.Request.Header.Get("Authorization")
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Parse the request body
	var education model.Education
	if err := c.ShouldBindJSON(&education); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Call the usecase to add the education
	err = h.usecase.AddEducation(c.Request.Context(), education, token)
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

func (h *CandidateHandler) VerifyEmail(c *gin.Context) {
	// Parse the request body
	var req struct {
		Email string `json:"email" binding:"required"`
		OTP   uint64 `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the usecase to verify the email
	err := h.usecase.VerifyEmail(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Verification failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *CandidateHandler) ResendOtp(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the usecase to resend the OTP
	err := h.usecase.ResendOtp(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend OTP: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP resent successfully"})
}

func (h *CandidateHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the usecase to handle forgot password
	err := h.usecase.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process forgot password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset OTP sent to email"})
}

func (h *CandidateHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required"`
		OTP         uint64 `json:"otp" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the usecase to reset the password
	err := h.usecase.ResetPassword(c.Request.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to reset password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *CandidateHandler) ChangePassword(c *gin.Context) {
	// Extract the token from the Authorization header
	authHeader := c.Request.Header.Get("Authorization")
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Extract user ID from the token
	userID, err := h.usecase.ExtractUserIDFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Parse the request body
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the usecase to change the password
	err = h.usecase.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to change password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
