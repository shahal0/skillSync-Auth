package postgres

import (
	"context"
	"errors"
	//m
	"io"
	"log"
	"math/rand"
	"os"
	"strings"

	"golang.org/x/oauth2"

	//logger "skillsync-authservice/Logger"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type candidatePG struct {
	db       *gorm.DB
	jwtMaker *pkg.JWTMaker
	pkg      *pkg.GcsClient
}

func NewCandidateRepository(db *gorm.DB, jwtMaker *pkg.JWTMaker) repository.CandidateRepository {
	return &candidatePG{db: db, jwtMaker: jwtMaker}
}

func (c *candidatePG) CreateCandidate(profile *model.Candidate) (int, error) {
	if err := c.db.Create(profile).Error; err != nil {
		return 0, err
	}
	res, _ := strconv.Atoi(profile.ID)
	return res, nil
}
func (c *candidatePG) VerifyToken(token string) (string, string, error) {
	claims, err := c.jwtMaker.VerifyToken(token)
	if err != nil {
		return "", "", err
	}
	return claims.UserID, claims.Role, nil
}
func (c *candidatePG) UpdateCandidate(input *model.UpdateCandidateInput, userID string) error {
	// Fetch the existing candidate by userID
	var candidate model.Candidate
	if err := c.db.Where("id = ?", userID).First(&candidate).Error; err != nil {
		return err // Return error if candidate is not found
	}

	// Update the candidate fields with the input values
	if input.Email != "" {
		candidate.Email = input.Email
	}
	if input.Name != "" {
		candidate.Name = input.Name
	}
	if input.Phone != 0 {
		candidate.Phone = input.Phone
	}
	if input.Experience != 0 {
		candidate.Experience = input.Experience
	}
	if input.CurrentLocation != "" {
		candidate.CurrentLocation = input.CurrentLocation
	}
	if input.PreferredLocation != "" {
		candidate.PreferredLocation = input.PreferredLocation
	}

	// Save the updated candidate back to the database
	return c.db.Save(&candidate).Error
}

func (c *candidatePG) GetCandidateByUserID(userID string) (*model.Candidate, error) {
	var cand model.Candidate
	skills := []model.Skills{}
	education := []model.Education{}
	err := c.db.Where("candidate_id = ?", userID).Find(&skills).Error
	if err != nil {
		return nil, err
	}
	err = c.db.Where("candidate_id = ?", userID).Find(&education).Error
	if err != nil {
		return nil, err
	}
	cand.Skills = skills
	cand.Education = education
	if err := c.db.Where("id = ?", userID).First(&cand).Error; err != nil {
		return nil, err
	}
	return &cand, nil
}

func (c *candidatePG) AddEducation(education model.Education, userID string) error {
	education.CandidateID = userID // Set the candidate ID

	return c.db.Create(&education).Error
}

func (c *candidatePG) AddSkills(skills model.Skills, candidateID string) error {
	// Normalize the skill
	skills.CandidateID = candidateID
	skills.Skill = strings.TrimSpace(skills.Skill)
	if skills.Skill == "" {
		return errors.New("skill cannot be empty")
	}

	log.Printf("DEBUG: Adding skill '%s' for candidate %s", skills.Skill, candidateID)

	// Check if the skill already exists for the candidate (case insensitive)
	var existingSkill model.Skills
	err := c.db.Where("candidate_id = ? AND LOWER(skill) = LOWER(?)", candidateID, skills.Skill).First(&existingSkill).Error
	if err == nil {
		// Skill already exists, update it
		log.Printf("DEBUG: Updating existing skill '%s' for candidate %s", skills.Skill, candidateID)
		return c.db.Model(&existingSkill).Updates(skills).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("ERROR: Failed to check for existing skill: %v", err)
		return err
	}

	// Skill does not exist, create a new one
	log.Printf("DEBUG: Creating new skill '%s' for candidate %s", skills.Skill, candidateID)
	return c.db.Create(&skills).Error
}

func (c *candidatePG) GetCandidateByEmail(email string) (*model.Candidate, error) {
	var emp model.Candidate
	if err := c.db.Where("email = ?", email).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (c *candidatePG) Login(request model.LoginRequest) (*model.LoginResponse, error) {
	var cand model.Candidate

	// Fetch the candidate by email
	if err := c.db.Where("email = ?", request.Email).First(&cand).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify the password
	passwordManager := pkg.NewPasswordManager()
	if err := passwordManager.CheckPassword(request.Password, cand.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate a token
	token, err := c.jwtMaker.GenerateToken(cand.ID, "candidate")
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Create the response
	response := &model.LoginResponse{
		ID:      cand.ID,
		Role:    "candidate",
		Token:   token,
		Message: "Login successful",
	}
	return response, nil
}

func (c *candidatePG) Signup(request model.SignupRequest) (*model.AuthResponse, error) {
	candidate := model.Candidate{
		Email:    request.Email,
		Name:     request.Name,
		Password: request.Password,
	}
	VerificationTable := model.VerificationTable{
		Email:              request.Email,
		Role:               request.Role,
		VerificationStatus: false,
	}
	if err := c.db.Create(&VerificationTable).Error; err != nil {
		return nil, err
	}
	err := pkg.SendOtp(c.db, request.Email, 5)
	if err != nil {
		return nil, err
	}
	if err := c.db.Create(&candidate).Error; err != nil {
		return nil, err
	}

	response := &model.AuthResponse{
		ID:      candidate.ID,
		Message: "Candidate created successfully",
	}
	return response, nil
}
func (c *candidatePG) UpdateCandidateProfile(profile *model.Candidate) error {
	if err := c.db.Save(profile).Where("id = ?", profile.ID).Error; err != nil {
		return err
	}
	return nil
}

func (c *candidatePG) VerifyEmail(email string, otp uint64) error {
	// Fetch the verification record from the database
	var verification model.VerificationTable
	err := c.db.Where("email = ?", email).First(&verification).Error
	if err != nil {
		return errors.New("email not found or not registered")
	}

	// Check if the OTP matches and is not expired
	now := time.Now().Unix()
	if verification.OTP != otp {
		return errors.New("invalid OTP")
	}
	if uint64(now) > verification.OTPExpiry {
		return errors.New("OTP has expired")
	}

	// Update the verification status
	err = c.db.Model(&model.VerificationTable{}).
		Where("email = ?", email).
		Update("verification_status", true).Error
	if err != nil {
		return errors.New("failed to update verification status")
	}
	errr := c.db.Model(&model.Candidate{}).Where("email = ?", email).Update("is_verified", true).Error
	if errr != nil {
		return errors.New("failed to update candidate verification status")
	}

	return nil
}

func (c *candidatePG) ResendOtp(email string) error {
	// Fetch the verification record from the database
	var verification model.VerificationTable
	err := c.db.Where("email = ?", email).First(&verification).Error
	if err != nil {
		return errors.New("email not found or not registered")
	}

	// Generate a new OTP
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	newOtp := r.Intn(900000) + 100000     // Generates a 6-digit OTP
	newExpiry := time.Now().Unix() + 5*60 // 5 minutes from now

	// Update the OTP and expiry in the database
	err = c.db.Model(&model.VerificationTable{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"otp":        newOtp,
			"otp_expiry": newExpiry,
		}).Error
	if err != nil {
		return errors.New("failed to update OTP in the database")
	}

	// Send the new OTP via email
	err = pkg.SendOtp(c.db, email, uint64(newOtp))
	if err != nil {
		return errors.New("failed to send OTP: " + err.Error())
	}

	return nil
}

func (c *candidatePG) SendPasswordResetOtp(email string) error {
	// Generate a new OTP
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000  // Generate a 6-digit OTP
	expiry := time.Now().Unix() + 5*60 // OTP valid for 5 minutes

	// Update or create the OTP record in the database
	err := c.db.Model(&model.VerificationTable{}).
		Where("email = ?", email).
		Assign(map[string]interface{}{
			"otp":                 otp,
			"otp_expiry":          expiry,
			"verification_status": false,
		}).
		FirstOrCreate(&model.VerificationTable{}).Error
	if err != nil {
		return errors.New("failed to store OTP in the database: " + err.Error())
	}

	// Send the OTP via email
	err = pkg.SendOtp(c.db, email, uint64(otp))
	if err != nil {
		return errors.New("failed to send OTP: " + err.Error())
	}

	return nil
}

func (c *candidatePG) VerifyPasswordResetOtp(email string, otp uint64) error {
	// Fetch the verification record
	var verification model.VerificationTable
	err := c.db.Where("email = ?", email).First(&verification).Error
	if err != nil {
		return errors.New("email not found or not registered")
	}

	// Check if the OTP matches and is not expired
	now := time.Now().Unix()
	if verification.OTP != otp {
		return errors.New("invalid OTP")
	}
	if uint64(now) > verification.OTPExpiry {
		return errors.New("OTP has expired")
	}

	return nil
}

func (c *candidatePG) UpdatePassword(email string, hashedPassword string) error {
	// Update the password in the database
	return c.db.Model(&model.Candidate{}).
		Where("email = ?", email).
		Update("password", hashedPassword).Error
}

func (c *candidatePG) UpdatePasswordByID(userID string, hashedPassword string) error {
	// Update the password in the database
	return c.db.Model(&model.Candidate{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func (c *candidatePG) AddResume(ctx context.Context, candidateID string, resumeReader io.Reader) (string, error) {
	// Generate a unique file path for the resume
	filePath := "resumes/" + candidateID + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ".pdf"

	// Save the resume file to storage
	_, err := c.pkg.UploadResume(ctx, resumeReader, filePath)
	if err != nil {
		return "", err
	}
	// Save to CandidateResume table
	candidateResume := model.CandidateResume{
		CandidateID: candidateID,
		GCSPath:     filePath,
	}
	if err := c.db.Create(&candidateResume).Error; err != nil {
		return "", err
	}

	// Also update Candidate's Resume field
	if err := c.db.Model(&model.Candidate{}).Where("id = ?", candidateID).Update("resume", filePath).Error; err != nil {
		return "", err
	}

	return filePath, nil
}

func (c *candidatePG) AddResumePath(ctx context.Context, candidateID string, filePath string) error {
	// Save to CandidateResume table
	candidateResume := model.CandidateResume{
		CandidateID: candidateID,
		GCSPath:     filePath,
	}
	if err := c.db.Create(&candidateResume).Error; err != nil {
		return err
	}

	// Update Candidate's Resume field
	return c.db.Model(&model.Candidate{}).Where("id = ?", candidateID).Update("resume", filePath).Error
}

func (c *candidatePG) GoogleLogin(redirectURL string) (string, error) {
	conf := pkg.GetGoogleOAuthConfig(redirectURL)
	return conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline), nil
}

func (c *candidatePG) GoogleCallback(code string) (*model.LoginResponse, error) {
	conf := pkg.GetGoogleOAuthConfig(os.Getenv("GOOGLE_REDIRECT_URI"))
	userinfo, err := pkg.GetGoogleUserInfo(conf, code)
	if err != nil {
		return nil, err
	}

	// Find or create candidate in your DB
	var candidate model.Candidate
	if err := c.db.Where("email = ?", userinfo.Email).First(&candidate).Error; err != nil {
		// Not found, create new candidate
		candidate = model.Candidate{
			Email: userinfo.Email,
			Name:  userinfo.Name,
			// You may want to set IsVerified = true, etc.
		}
		if err := c.db.Create(&candidate).Error; err != nil {
			return nil, errors.New("failed to create candidate: " + err.Error())
		}
	}
	tokenStr, err := c.jwtMaker.GenerateToken(candidate.ID, "candidate")
	if err != nil {
		return nil, errors.New("failed to generate JWT: " + err.Error())
	}

	return &model.LoginResponse{
		ID:      candidate.ID,
		Role:    "candidate",
		Token:   tokenStr,
		Message: "Google login successful",
	}, nil
}

func (c *candidatePG) GetSkills(candidateID string) ([]string, error) {
	var skills []model.Skills

	// Query the database for skills associated with the candidate
	if err := c.db.Where("candidate_id = ?", candidateID).Find(&skills).Error; err != nil {
		log.Printf("ERROR: Failed to retrieve skills for candidate %s: %v", candidateID, err)
		return nil, err
	}

	log.Printf("DEBUG: Retrieved %d skills for candidate %s", len(skills), candidateID)

	// Extract the skill names
	skillNames := make([]string, 0, len(skills))
	for _, skill := range skills {
		if skill.Skill != "" {
			skillNames = append(skillNames, skill.Skill)
		}
	}
	if len(skillNames) == 0 {
		log.Printf("WARNING: No valid skills found for candidate %s", candidateID)
	}

	return skillNames, nil
}

// GetCandidatesWithPagination retrieves candidates with pagination and filtering
func (c *candidatePG) GetCandidatesWithPagination(page, limit int32, filters map[string]interface{}) ([]*model.Candidate, int64, error) {
	var candidates []*model.Candidate
	var totalCount int64

	// Create a query builder
	query := c.db.Model(&model.Candidate{})

	// Apply filters if any
	if filters != nil {
		for key, value := range filters {
			switch key {
			case "skills":
				if skills, ok := value.([]string); ok && len(skills) > 0 {
					// Join with skills table and filter by skill names
					query = query.Joins("JOIN skills ON skills.candidate_id = candidates.id")
					query = query.Where("skills.skill IN (?)", skills)
					query = query.Group("candidates.id")
				}
			case "experience":
				if exp, ok := value.(int); ok {
					query = query.Where("experience >= ?", exp)
				}
			case "location":
				if loc, ok := value.(string); ok && loc != "" {
					query = query.Where("current_location LIKE ? OR preferred_location LIKE ?", "%"+loc+"%", "%"+loc+"%")
				}
			default:
				// For other fields, apply direct equality filter
				query = query.Where(key+" = ?", value)
			}
		}
	}

	// Count total matching records (before pagination)
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(int(offset)).Limit(int(limit)).Find(&candidates).Error; err != nil {
		return nil, 0, err
	}

	// For each candidate, fetch their skills and education
	for _, candidate := range candidates {
		// Fetch skills
		var skills []model.Skills
		if err := c.db.Where("candidate_id = ?", candidate.ID).Find(&skills).Error; err != nil {
			log.Printf("Warning: Failed to fetch skills for candidate %s: %v", candidate.ID, err)
		}
		candidate.Skills = skills

		// Fetch education
		var education []model.Education
		if err := c.db.Where("candidate_id = ?", candidate.ID).Find(&education).Error; err != nil {
			log.Printf("Warning: Failed to fetch education for candidate %s: %v", candidate.ID, err)
		}
		candidate.Education = education
	}

	return candidates, totalCount, nil
}
