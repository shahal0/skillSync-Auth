package handlers

import (
	"log"
	"net/http"

	//"os/user"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/internal/usecase"
	"skillsync-authservice/pkg"

	"github.com/gin-gonic/gin"
)

type EmployerHandler struct {
	usecase   *usecase.EmployerUsecase
	jwthelper *pkg.JWTMaker
}

func NewEmployerHandler(router *gin.RouterGroup, uc *usecase.EmployerUsecase) {
	handler := &EmployerHandler{usecase: uc}

	employer := router.Group("/employer")
	{
		employer.POST("/signup", handler.Signup)
		employer.POST("/login", handler.Login)
		employer.POST("/verify-email", handler.VerifyEmail)
		employer.POST("/resend-otp", handler.ResendOtp)
		employer.POST("/forgot-password", handler.ForgotPassword) // Forgot Password route
		employer.POST("/reset-password", handler.ResetPassword)   // Reset Password route
	}
}

func (h *EmployerHandler) UpdateProfile(c *gin.Context) {
	// Get the Authorization header
	authHeader := c.Request.Header.Get("Authorization")
	log.Println("Authorization Header:", authHeader)

	// Extract the token
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		log.Println("Token Extraction Error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Extract user ID from the token
	userID, err := h.jwthelper.ExtractUserIDFromToken(token)
	if err != nil {
		log.Println("User ID Extraction Error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Parse the request body
	var profile model.UpdateEmployerInput
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input " + err.Error()})
		return
	}

	// Set the user ID in the profile
	profile.ID = userID

	// Call the usecase to update the profile
	err = h.usecase.UpdateProfile(c.Request.Context(), &profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employer profile updated successfully"})
}

func (h *EmployerHandler) GetProfile(c *gin.Context) {
	// Get the Authorization header
	authHeader := c.Request.Header.Get("Authorization")
	log.Println("Authorization Header:", authHeader)

	// Extract the token
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		log.Println("Token Extraction Error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	// Fetch the profile using the token
	profile, err := h.usecase.GetProfile(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}
func (h *EmployerHandler) Signup(c *gin.Context) {
	var req model.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Call the usecase to handle signup and send OTP
	res, err := h.usecase.Signup(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}
func (h *EmployerHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input " + err.Error()})
		return
	}

	res, err := h.usecase.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *EmployerHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		OTP   uint64 `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	err := h.usecase.VerifyEmail(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Verification failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *EmployerHandler) ResendOtp(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	err := h.usecase.ResendOtp(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend OTP: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP resent successfully"})
}

func (h *EmployerHandler) ForgotPassword(c *gin.Context) {
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

func (h *EmployerHandler) ResetPassword(c *gin.Context) {
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
