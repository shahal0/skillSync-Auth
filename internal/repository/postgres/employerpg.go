package postgres

import (
	//"go/token"

	"math/rand"
	"os"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	"strconv"
	"time"

	//"github.com/golang-jwt/jwt"
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
		return 0, err
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
func (r *employerPG) GetEmployerByEmail(email string) (*model.Employer, error) {
	query := `SELECT id, company_name, email, password, company_name FROM employers WHERE email = $1`

	var emp model.Employer
	err := r.db.Raw(query, email).Scan(&emp).Error
	if err != nil {
		return nil, errors.New("employer not found")
	}

	return &emp, nil
}

func (r *employerPG) Login(request model.LoginRequest) (*model.LoginResponse, error) {
	var emp model.Employer
	if err := r.db.Where("email = ? AND password = ?", request.Email, request.Password).First(&emp).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}
	id := strconv.Itoa(emp.ID)
	token, err := r.jwtMaker.GenerateToken(id, "employer")
	if err != nil {
		return nil, err
	}
	response := &model.LoginResponse{
		ID:    id,
		Role:  "employer",
		Token: token,
	}
	return response, nil
}
func (e *employerPG) Signup(request model.SignupRequest) (*model.AuthResponse, error) {
	profile := &model.Employer{
		CompanyName: request.Name,
		Email:       request.Email,
		Password:    request.Password,
	}
	if err := e.db.Create(profile).Error; err != nil {
		return nil, err
	}
	response := &model.AuthResponse{
		ID:      strconv.Itoa(profile.ID),
		Message: "Employer created successfully",
	}
	return response, nil
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
