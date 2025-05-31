package usecase

import (
	"context"
	"errors"
	"math/rand"
	"net/smtp"
	"strconv"
	"time"

	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"

	"gorm.io/gorm"
)

type EmployerUsecase struct {
	employerRepo repository.EmployerRepository
	jwtcliams    *pkg.JWTMaker
	db           *gorm.DB // Add the gorm.DB instance
}

func NewEmployerUsecase(repo repository.EmployerRepository, jwtMaker *pkg.JWTMaker, db *gorm.DB) *EmployerUsecase {
	return &EmployerUsecase{
		employerRepo: repo,
		jwtcliams:    jwtMaker,
		db:           db, // Initialize the gorm.DB instance
	}
}

func (uc *EmployerUsecase) UpdateProfile(ctx context.Context, profile *model.UpdateEmployerInput) error {
	return uc.employerRepo.UpdateEmployer(profile)
}

func (uc *EmployerUsecase) GetProfile(ctx context.Context, token string) (*model.Employer, error) {
	// Extract user ID from token
	userID, err := uc.jwtcliams.ExtractUserIDFromToken(token)
	if err != nil {
		return nil, err
	}

	// Fetch the employer profile using the extracted user ID
	return uc.employerRepo.GetEmployerByUserID(userID)
}

// GetProfileById fetches an employer profile by ID without requiring a token
func (uc *EmployerUsecase) GetProfileById(ctx context.Context, employerId string) (*model.Employer, error) {
	// Convert string ID to uint if necessary
	id, err := strconv.ParseUint(employerId, 10, 64)
	if err != nil {
		return nil, err
	}

	// Fetch the employer profile using the ID
	return uc.employerRepo.GetEmployerById(uint(id))
}

func (uc *EmployerUsecase) Signup(ctx context.Context, req model.SignupRequest) (*model.AuthResponse, error) {
	// Check if the email already exists
	existingEmployer, err := uc.employerRepo.GetEmployerByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingEmployer != nil {
		return nil, errors.New("email already exists")
	}

	// Hash the password
	hashedPassword, err := pkg.NewPasswordManager().HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	req.Password = hashedPassword

	// Parse phone number from context if available
	var phone int64
	if phoneStr, ok := req.Context["phone"]; ok && phoneStr != "" {
		phone, _ = strconv.ParseInt(phoneStr, 10, 64)
	}

	// Create the employer record with all available details
	employer := &model.Employer{
		Email:       req.Email,
		Password:    req.Password,
		CompanyName: req.Name,
		Phone:       phone,
	}

	// Add optional fields if they exist in the context
	if industry, ok := req.Context["industry"]; ok {
		employer.Industry = industry
	}
	if location, ok := req.Context["location"]; ok {
		employer.Location = location
	}
	if website, ok := req.Context["website"]; ok {
		employer.Website = website
	}

	id, err := uc.employerRepo.CreateEmployer(employer)
	if err != nil {
		return nil, err
	}

	// Send OTP to the employer's email
	err = uc.SendOtp(req.Email, 5*60) // 5 minutes expiry
	if err != nil {
		return nil, errors.New("failed to send OTP: " + err.Error())
	}

	return &model.AuthResponse{
		ID:      strconv.Itoa(id),
		Message: "Employer registered successfully. OTP sent to email.",
	}, nil
}

// SendOtp generates and sends an OTP to the given email
func (uc *EmployerUsecase) SendOtp(email string, expiry int64) error {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000 // Generate a 6-digit OTP

	// Store OTP in the database
	verification := model.VerificationTable{
		Email:              email,
		OTP:                uint64(otp),
		OTPExpiry:          uint64(time.Now().Unix() + expiry),
		VerificationStatus: false,
	}
	if err := uc.db.Where("email = ?", email).
		Assign(verification).
		FirstOrCreate(&verification).Error; err != nil {
		return errors.New("failed to store OTP in the database: " + err.Error())
	}

	// Send OTP via email
	auth := smtp.PlainAuth("", "petplate0@gmail.com", "fsjazfcjcllfnxqu", "smtp.gmail.com")
	message := []byte("Subject: Your OTP Code\n\nYour OTP is: " + strconv.Itoa(otp))
	err := smtp.SendMail("smtp.gmail.com:587", auth, "petplate0@gmail.com", []string{email}, message)
	if err != nil {
		return errors.New("failed to send OTP via email: " + err.Error())
	}

	return nil
}

func (uc *EmployerUsecase) Login(input model.LoginRequest) (*model.LoginResponse, error) {
	return uc.employerRepo.Login(input)
}

func (uc *EmployerUsecase) GoogleLogin(redirectURL string) (string, error) {
	return uc.employerRepo.GoogleLogin(redirectURL)
}

func (uc *EmployerUsecase) GoogleCallback(ctx context.Context, req model.GoogleCallbackRequest) (*model.LoginResponse, error) {
	// Call the repository's GoogleCallback method with just the code
	loginResp, err := uc.employerRepo.GoogleCallback(req.Code)
	if err != nil {
		return nil, err
	}

	// Convert LoginResponse to AuthResponse
	// Note: AuthResponse only has ID and Message fields, not Token
	return &model.LoginResponse{
		ID:      loginResp.ID,
		Message: loginResp.Message,
		Token:   loginResp.Token,
	}, nil
}
func (uc *EmployerUsecase) VerifyEmail(ctx context.Context, email string, otp uint64) error {
	return uc.employerRepo.VerifyEmail(email, otp)
}

func (uc *EmployerUsecase) ResendOtp(ctx context.Context, req model.ResendOtpRequest) (model.GenericResponse, error) {
	err := uc.employerRepo.ResendOtp(req.Email)
	if err != nil {
		return model.GenericResponse{}, err
	}
	return model.GenericResponse{
		Message: "OTP sent successfully",
		Success: true,
	}, nil
}

func (uc *EmployerUsecase) ForgotPassword(ctx context.Context, email string) error {
	// Generate and send OTP
	return uc.employerRepo.SendPasswordResetOtp(email)
}

func (uc *EmployerUsecase) ResetPassword(ctx context.Context, email string, otp uint64, newPassword string) error {
	// Verify OTP
	err := uc.employerRepo.VerifyPasswordResetOtp(email, otp)
	if err != nil {
		return err
	}

	// Hash the new password
	hashedPassword, err := pkg.NewPasswordManager().HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update the password in the database
	return uc.employerRepo.UpdatePassword(email, hashedPassword)
}

func (uc *EmployerUsecase) ExtractUserIDFromToken(token string) (string, error) {
	return uc.jwtcliams.ExtractUserIDFromToken(token)
}
func (uc *EmployerUsecase) ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string) error {
	// Fetch the employer by ID
	employer, err := uc.employerRepo.GetEmployerByUserID(userID)
	if err != nil {
		return errors.New("employer not found")
	}

	// Verify the current password
	passwordManager := pkg.NewPasswordManager()
	if err := passwordManager.CheckPassword(currentPassword, employer.Password); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := passwordManager.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// Update the password
	return uc.employerRepo.UpdatePasswordByID(userID, hashedPassword)
}

// GetEmployersWithPagination retrieves employers with pagination and filtering
func (uc *EmployerUsecase) GetEmployersWithPagination(ctx context.Context, page, limit int32, filters map[string]interface{}) ([]*model.Employer, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Call the repository method to get paginated employers
	return uc.employerRepo.GetEmployersWithPagination(page, limit, filters)
}
