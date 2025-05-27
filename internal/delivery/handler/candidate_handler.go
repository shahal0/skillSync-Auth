package handlers

import (
	//"go/token"

	"log"
	"net/http"
	"os"
	logger "skillsync-authservice/Logger"
	"skillsync-authservice/config"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/internal/usecase"
	"skillsync-authservice/pkg"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
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
		candidate.GET("/auth/google/login", func(c *gin.Context) {
			GoogleLogin(c, "candidate")
		})
		candidate.GET("/auth/google/callback", GoogleCallback)
		candidate.POST("/upload/resume", handler.UploadResume)
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

	// Log the profile update request
	logger.Log.WithFields(logrus.Fields{
		"user_id": profile.ID,
		"action":  "update_profile",
	}).Info("Profile update request received")

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
	var skill model.Skills
	if err := c.ShouldBindJSON(&skill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Set the candidate ID
	skill.CandidateID = userID

	// Call the usecase to add the skill
	err = h.usecase.AddSkills(c.Request.Context(), skill, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		logger.HandleError(c, http.StatusBadRequest, nil, "State parameter is missing")
		return
	}

	// Retrieve the stored state token from the session
	session := sessions.Default(c)
	storedState := session.Get("state")
	if storedState == nil || state != storedState {
		logger.HandleError(c, http.StatusBadRequest, nil, "Invalid state token")
		return
	}

	// Get the authorization code from the query parameters
	code := c.Query("code")
	if code == "" {
		logger.HandleError(c, http.StatusBadRequest, nil, "Authorization code not provided")
		return
	}

	// Exchange the authorization code for an access token
	token, err := config.GoogleOAuthConfig.Exchange(c, code)
	if err != nil {
		logger.HandleError(c, http.StatusInternalServerError, err, "Failed to exchange token")
		return
	}

	logger.Log.WithField("access_token", token.AccessToken).Info("Access token retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Google login successful"})
}

func GoogleLogin(c *gin.Context, userType string) {
	state := pkg.GenerateStateToken()
	session := sessions.Default(c)
	session.Set("state", state) // Store the state token in the session
	session.Save()              // Save the session

	// Dynamically set the RedirectURL based on user type
	baseRedirectURL := os.Getenv("GOOGLE_REDIRECT_URI") // Use the base redirect URL from the environment
	var redirectURL string
	if userType == "employer" {
		redirectURL = baseRedirectURL + "/employer/auth/google/callback"
	} else {
		redirectURL = baseRedirectURL + "/candidate/auth/google/callback"
	}

	log.Println("Redirect URL:", redirectURL)

	// Update the RedirectURL in GoogleOAuthConfig
	config.GoogleOAuthConfig.RedirectURL = redirectURL

	// Generate the OAuth URL
	url := config.GoogleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	log.Println("Generated Google OAuth URL:", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *CandidateHandler) UploadResume(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	token, err := pkg.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	userID, err := h.usecase.ExtractUserIDFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	file, fileHeader, err := c.Request.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume file is required"})
		return
	}
	defer file.Close()

	_,errr := h.usecase.AddResume(c.Request.Context(), file, fileHeader, userID)
	if errr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload resume"+": " + errr.Error()})
		logger.Log.WithError(errr).Error("Failed to upload resume")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resume uploaded successfully"})
}

