package postgres

import (
	"errors"
	"math/rand"
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
	if input.Resume != "" {
		candidate.Resume = input.Resume
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

func (c *candidatePG) AddSkills(skills model.Skills) error {
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
		ID:    cand.ID,
		Role:  "candidate",
		Token: token,
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
