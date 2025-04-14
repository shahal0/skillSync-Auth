package usecase

import (
	"context"
	"errors"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	"strconv"

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

func (uc *EmployerUsecase) Signup(ctx context.Context, req model.SignupRequest) (*model.AuthResponse, error) {
	// Create the employer record
	employer := &model.Employer{
		Email:       req.Email,
		Password:    req.Password,
		CompanyName: req.Name,
	}
	id, err := uc.employerRepo.CreateEmployer(employer)
	if err != nil {
		return nil, err
	}

	// Generate and send OTP
	err = pkg.SendOtp(uc.db, req.Email, 5)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		ID:      strconv.Itoa(id),
		Message: "Employer registered successfully",
	}, nil
}

func (uc *EmployerUsecase) Login(input model.LoginRequest) (*model.LoginResponse, error) {
	return uc.employerRepo.Login(input)
}

func (uc *EmployerUsecase) VerifyEmail(ctx context.Context, email string, otp uint64) error {
	return uc.employerRepo.VerifyEmail(email, otp)
}

func (uc *EmployerUsecase) ResendOtp(ctx context.Context, email string) error {
	return uc.employerRepo.ResendOtp(email)
}

func (uc *EmployerUsecase) ForgotPassword(ctx context.Context, email string) error {
	// Generate and send OTP
	return uc.employerRepo.SendPasswordResetOtp(email)
}

func (uc *EmployerUsecase) ResetPassword(ctx context.Context, email string, otp uint64, newPassword string) error {
	// Verify OTP
	// Check if the email exists in the database
	existingEmployer, err := uc.employerRepo.GetEmployerByEmail(email)
	if err != nil {
		return errors.New("email not found or not registered")
	}
	if existingEmployer == nil {
		return errors.New("email not found or not registered")
	}
	errr:= uc.employerRepo.VerifyPasswordResetOtp(email, otp)
	if errr != nil {
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
