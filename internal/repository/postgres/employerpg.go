package postgres

import (
	//"go/token"
	"log"
	"math/rand"
	"os"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	"strconv"
	"time"
	//"context"

	//"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	//"context"
	"errors"
)

type employerPG struct {
	db       *gorm.DB
	jwtMaker *pkg.JWTMaker
}

func NewEmployerRepository(db *gorm.DB) repository.EmployerRepository {
	return &employerPG{db: db, jwtMaker: pkg.NewJWTMaker(os.Getenv("JWT_SECRET"))}
}

func (e *employerPG) CreateEmployer(profile *model.Employer) (int, error) {
	if err := e.db.Create(profile).Error; err != nil {
		return 0, errors.New("failed to create employer record: " + err.Error())
	}
	return profile.ID, nil
}

func (e *employerPG) UpdateEmployer(input *model.UpdateEmployerInput) error {
	var profile model.Employer
	if err := e.db.First(&profile, "id = ?", input.ID).Error; err != nil {
		return err
	}
	profile.CompanyName = input.CompanyName
	profile.Email = input.Email
	profile.Industry = input.Industry
	profile.Phone = input.Phone
	profile.Location = input.Location
	profile.Website = input.Website
	return e.db.Save(&profile).Error
}

func (e *employerPG) GetEmployerByUserID(userID string) (*model.Employer, error) {
	var emp model.Employer
	if err := e.db.Where("id= ?", userID).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (e *employerPG) GetEmployerById(id uint) (*model.Employer, error) {
	var emp model.Employer
	if err := e.db.Where("id = ?", id).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (e *employerPG) GetEmployerByEmail(email string) (*model.Employer, error) {
	var employer model.Employer
	if email == "" {
		return nil, errors.New("email is empty")
	}
	err := e.db.Where("email = ?", email).First(&employer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // Return nil if no record is found
	}
	if err != nil {
		return nil, err // Return the error if it's not a "record not found" error
	}
	return &employer, nil
}

func (r *employerPG) Login(request model.LoginRequest) (*model.LoginResponse, error) {
	var emp model.Employer

	// Check if the email exists in the database
	if err := r.db.Where("email = ?", request.Email).First(&emp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("email not found")
		}
		return nil, err
	}

	// Verify the password
	passwordManager := pkg.NewPasswordManager()
	if err := passwordManager.CheckPassword(request.Password, emp.Password); err != nil {
		log.Print("passwords :", emp.Password, request.Password)
		return nil, errors.New("invalid email or password")
	}

	// Generate a token
	id := strconv.Itoa(emp.ID)
	token, err := r.jwtMaker.GenerateToken(id, "employer")
	if err != nil {
		return nil, err
	}

	// Create the response
	response := &model.LoginResponse{
		ID:      id,
		Role:    "employer",
		Token:   token,
		Message: "Login successful",
	}
	return response, nil
}

func (e *employerPG) Signup(request model.SignupRequest) (*model.AuthResponse, error) {
	// Extract additional fields from the request context if available
	phone, _ := strconv.ParseInt(request.Context["phone"], 10, 64)

	profile := &model.Employer{
		CompanyName: request.Name,
		Email:       request.Email,
		Password:    request.Password,
		Phone:       phone,
		Industry:    request.Context["industry"],
		Location:    request.Context["location"],
		Website:     request.Context["website"],
	}

	// Create the employer record in the database
	if err := e.db.Create(profile).Error; err != nil {
		return nil, errors.New("failed to create employer record: " + err.Error())
	}

	// Generate a token
	id := strconv.Itoa(profile.ID)
	//, err := e.jwtMaker.GenerateToken(id, "employer")
	// if err != nil {
	// 	return nil, errors.New("failed to generate token: " + err.Error())
	// }

	// Return the AuthResponse
	return &model.AuthResponse{
		ID:      id,
		Message: "Employer registered successfully. OTP sent to email.",
	}, nil
}

func (r *employerPG) SendOtp(email string) error {
	// Generate a new OTP
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000  // Generate a 6-digit OTP
	expiry := time.Now().Unix() + 5*60 // OTP valid for 5 minutes

	// Update or create the OTP record in the database
	err := r.db.Model(&model.VerificationTable{}).
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
	err = pkg.SendOtp(r.db, email, uint64(otp))
	if err != nil {
		return errors.New("failed to send OTP: " + err.Error())
	}

	return nil
}

func (r *employerPG) VerifyEmail(email string, otp uint64) error {
	// Fetch the verification record from the database
	var verification model.VerificationTable
	err := r.db.Where("email = ?", email).First(&verification).Error
	if err != nil {
		return errors.New("email not found or not registered")
	}

	// Check if the OTP matches
	if verification.OTP != otp {
		return errors.New("invalid OTP")
	}

	// Check if the OTP has expired
	now := time.Now().Unix()
	if uint64(now) > verification.OTPExpiry {
		return errors.New("OTP has expired")
	}

	// Update the verification status
	err = r.db.Model(&model.VerificationTable{}).
		Where("email = ?", email).
		Update("verification_status", true).Error
	if err != nil {
		return errors.New("failed to update verification status")
	}

	// Update the employer's verification status
	err = r.db.Model(&model.Employer{}).
		Where("email = ?", email).
		Update("is_verified", true).Error
	if err != nil {
		return errors.New("failed to update employer verification status")
	}

	return nil
}

func (r *employerPG) ResendOtp(email string) error {
	return r.SendOtp(email)
}

func (r *employerPG) SendPasswordResetOtp(email string) error {
	// Generate a new OTP
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000  // Generate a 6-digit OTP
	expiry := time.Now().Unix() + 5*60 // OTP valid for 5 minutes

	// Update or create the OTP record in the database
	err := r.db.Model(&model.VerificationTable{}).
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
	err = pkg.SendOtp(r.db, email, uint64(otp))
	if err != nil {
		return errors.New("failed to send OTP: " + err.Error())
	}

	return nil
}

func (r *employerPG) VerifyPasswordResetOtp(email string, otp uint64) error {
	// Fetch the verification record
	var verification model.VerificationTable
	err := r.db.Where("email = ?", email).First(&verification).Error
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

func (r *employerPG) UpdatePassword(email string, hashedPassword string) error {
	// Update the password in the database
	return r.db.Model(&model.Employer{}).
		Where("email = ?", email).
		Update("password", hashedPassword).Error
}

func (r *employerPG) UpdatePasswordByID(userID string, hashedPassword string) error {
	// Update the password in the database
	return r.db.Model(&model.Employer{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func (e *employerPG) GoogleLogin(redirectURL string) (string, error) {
	conf := pkg.GetGoogleOAuthConfig(redirectURL)
	return conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline), nil
}

func (e *employerPG) GoogleCallback(code string) (*model.LoginResponse, error) {
	conf := pkg.GetGoogleOAuthConfig(os.Getenv("GOOGLE_REDIRECT_URI"))
	userinfo, err := pkg.GetGoogleUserInfo(conf, code)
	if err != nil {
		return nil, err
	}

	// Find or create employer in your DB
	var employer model.Employer
	if err := e.db.Where("email = ?", userinfo.Email).First(&employer).Error; err != nil {
		// Not found, create new employer
		employer = model.Employer{
			Email:       userinfo.Email,
			CompanyName: userinfo.Name,
			// You may want to set IsVerified = true, etc.
		}
		if err := e.db.Create(&employer).Error; err != nil {
			return nil, errors.New("failed to create employer: " + err.Error())
		}
	}
	tokenStr, err := e.jwtMaker.GenerateToken(strconv.Itoa(employer.ID), "employer")
	if err != nil {
		return nil, errors.New("failed to generate JWT: " + err.Error())
	}

	return &model.LoginResponse{
		ID:      strconv.Itoa(employer.ID),
		Role:    "employer",
		Token:   tokenStr,
		Message: "Google login successful",
	}, nil
}

// GetEmployersWithPagination retrieves employers with pagination and filtering support
func (e *employerPG) GetEmployersWithPagination(page, limit int32, filters map[string]interface{}) ([]*model.Employer, int64, error) {
	// Ensure valid pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Start building the query
	query := e.db.Model(&model.Employer{})

	// Apply filters if provided
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		keywordSearch := "%" + keyword + "%"
		query = query.Where("company_name ILIKE ? OR email ILIKE ?", keywordSearch, keywordSearch)
	}

	if industry, ok := filters["industry"].(string); ok && industry != "" {
		query = query.Where("industry ILIKE ?", "%"+industry+"%")
	}

	if location, ok := filters["location"].(string); ok && location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	// Count total records that match the filters
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, errors.New("error counting employers: " + err.Error())
	}

	// Apply pagination
	offset := (page - 1) * limit
	
	// Execute the query with pagination
	var employers []*model.Employer
	if err := query.Offset(int(offset)).Limit(int(limit)).Find(&employers).Error; err != nil {
		return nil, 0, errors.New("error fetching employers: " + err.Error())
	}

	return employers, totalCount, nil
}
