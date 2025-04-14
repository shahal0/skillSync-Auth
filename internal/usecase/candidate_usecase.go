package usecase

import (
	"context"
	"errors"
	"fmt"

	//"os/user"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"

	"gorm.io/gorm" // Import GORM package
	//"github.com/golang-jwt/jwt/v4"
)

type CandidateUsecase struct {
	candidateRepo repository.CandidateRepository
	db            *gorm.DB // Assuming you're using GORM for database operations
	jwtcliams     *pkg.JWTMaker
}

func NewCandidateUsecase(repo repository.CandidateRepository, jwtmaker *pkg.JWTMaker, db *gorm.DB) *CandidateUsecase {
	return &CandidateUsecase{candidateRepo: repo, jwtcliams: jwtmaker, db: db}
}

func (uc *CandidateUsecase) GetProfile(ctx context.Context, token string) (*model.Candidate, error) {
	userID, err := uc.jwtcliams.ExtractUserIDFromToken(token)
	if err != nil {
		return nil, err
	}
	return uc.candidateRepo.GetCandidateByUserID(userID)
}
func (uc *CandidateUsecase) UpdateCandidateProfile(ctx context.Context, input *model.UpdateCandidateInput, token string) error {
	// Extract user ID from token
	userID, err := uc.jwtcliams.ExtractUserIDFromToken(token)
	if err != nil {
		return err
	}

	// Call the repository to update the candidate profile
	return uc.candidateRepo.UpdateCandidate(input, userID)
}
func (uc *CandidateUsecase) AddSkills(ctx context.Context, skills model.Skills) error {
	return uc.candidateRepo.AddSkills(skills)
}
func (uc *CandidateUsecase) AddEducation(ctx context.Context, edu model.Education, token string) error {
	// Extract user ID from token
	userID, err := uc.jwtcliams.ExtractUserIDFromToken(token)
	if err != nil {
		return err
	}

	// Call the repository to add the education
	return uc.candidateRepo.AddEducation(edu, userID)
}
func (uc *CandidateUsecase) Signup(input model.SignupRequest) (*model.AuthResponse, error) {
	// Check if the email already exists
	existingCandidate, _ := uc.candidateRepo.GetCandidateByEmail(input.Email)
	if existingCandidate != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash the password
	hashedPassword, err := pkg.NewPasswordManager().HashPassword(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	input.Password = hashedPassword

	// Save the candidate
	return uc.candidateRepo.Signup(input)
}
func (uc *CandidateUsecase) Login(input model.LoginRequest) (*model.LoginResponse, error) {
	return uc.candidateRepo.Login(input)
}
func (uc *CandidateUsecase) ExtractUserIDFromToken(token string) (string, error) {
	return uc.jwtcliams.ExtractUserIDFromToken(token)
}
func (uc *CandidateUsecase) VerifyEmail(ctx context.Context, email string, otp uint64) error {
	// Delegate the verification logic to the repository
	return uc.candidateRepo.VerifyEmail(email, otp)
}
func (uc *CandidateUsecase) ResendOtp(ctx context.Context, email string) error {
	// Delegate the resend OTP logic to the repository
	return uc.candidateRepo.ResendOtp(email)
}
func (uc *CandidateUsecase) ForgotPassword(ctx context.Context, email string) error {
	// Generate and send OTP
	return uc.candidateRepo.SendPasswordResetOtp(email)
}

func (uc *CandidateUsecase) ResetPassword(ctx context.Context, email string, otp uint64, newPassword string) error {
	// Verify OTP
	err := uc.candidateRepo.VerifyPasswordResetOtp(email, otp)
	if err != nil {
		return err
	}

	// Hash the new password
	hashedPassword, err := pkg.NewPasswordManager().HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update the password in the database
	return uc.candidateRepo.UpdatePassword(email, hashedPassword)
}

func (uc *CandidateUsecase) ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string) error {
	// Fetch the candidate by user ID
	candidate, err := uc.candidateRepo.GetCandidateByUserID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify the current password
	passwordManager := pkg.NewPasswordManager()
	if err := passwordManager.CheckPassword(currentPassword, candidate.Password); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := passwordManager.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// Update the password in the database
	return uc.candidateRepo.UpdatePasswordByID(userID, hashedPassword)
}
