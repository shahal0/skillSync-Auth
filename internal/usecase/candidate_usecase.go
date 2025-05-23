package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	//"os/user"
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	//"github.com/golang-jwt/jwt/v4"
)

type CandidateUsecase struct {
	repo     repository.CandidateRepository
	jwtMaker *pkg.JWTMaker
	gcs      *pkg.GcsClient
}

func NewCandidateUsecase(repo repository.CandidateRepository, jwtMaker *pkg.JWTMaker, gcs *pkg.GcsClient) *CandidateUsecase {
	return &CandidateUsecase{
		repo:     repo,
		jwtMaker: jwtMaker,
		gcs:      gcs,
	}
}

func (uc *CandidateUsecase) GetProfile(ctx context.Context, token string) (*model.Candidate, error) {
	userID, err := uc.jwtMaker.ExtractUserIDFromToken(token)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetCandidateByUserID(userID)
}
func (uc *CandidateUsecase) UpdateCandidateProfile(ctx context.Context, input *model.UpdateCandidateInput, token string) error {
	// Extract user ID from token
	userID, err := uc.jwtMaker.ExtractUserIDFromToken(token)
	if err != nil {
		return err
	}

	// Call the repository to update the candidate profile
	return uc.repo.UpdateCandidate(input, userID)
}
func (uc *CandidateUsecase) AddSkills(ctx context.Context, skills model.Skills, token string) error {
	// Extract user ID from token
	userID, err := uc.jwtMaker.ExtractUserIDFromToken(token)
	if err != nil {
		return err
	}
	return uc.repo.AddSkills(skills, userID)
}
func (uc *CandidateUsecase) AddEducation(ctx context.Context, edu model.Education, token string) error {
	// Extract user ID from token
	userID, err := uc.jwtMaker.ExtractUserIDFromToken(token)
	if err != nil {
		return err
	}

	// Call the repository to add the education
	return uc.repo.AddEducation(edu, userID)
}
func (uc *CandidateUsecase) Signup(input model.SignupRequest) (*model.AuthResponse, error) {
	// Check if the email already exists
	existingCandidate, _ := uc.repo.GetCandidateByEmail(input.Email)
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
	return uc.repo.Signup(input)
}
func (uc *CandidateUsecase) Login(input model.LoginRequest) (*model.LoginResponse, error) {
	return uc.repo.Login(input)
}
func (uc *CandidateUsecase) ExtractUserIDFromToken(token string) (string, error) {
	return "", nil
}

// VerifyToken verifies a JWT token and returns the claims
func (uc *CandidateUsecase) VerifyToken(token string) (*pkg.Claims, error) {
	return uc.jwtMaker.VerifyToken(token)
}

func (uc *CandidateUsecase) VerifyEmail(ctx context.Context, email string, otp uint64) error {
	// Delegate the verification logic to the repository
	return uc.repo.VerifyEmail(email, otp)
}
func (uc *CandidateUsecase) ResendOtp(ctx context.Context, email string) error {
	// Delegate the resend OTP logic to the repository
	return uc.repo.ResendOtp(email)
}
func (uc *CandidateUsecase) ForgotPassword(ctx context.Context, email string) error {
	// Generate and send OTP
	return uc.repo.SendPasswordResetOtp(email)
}

func (uc *CandidateUsecase) ResetPassword(ctx context.Context, email string, otp uint64, newPassword string) error {
	// Verify OTP
	err := uc.repo.VerifyPasswordResetOtp(email, otp)
	if err != nil {
		return err
	}

	// Hash the new password
	hashedPassword, err := pkg.NewPasswordManager().HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update the password in the database
	return uc.repo.UpdatePassword(email, hashedPassword)
}

func (uc *CandidateUsecase) ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string) error {
	// Fetch the candidate by user ID
	candidate, err := uc.repo.GetCandidateByUserID(userID)
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
	return uc.repo.UpdatePasswordByID(userID, hashedPassword)
}

func (uc *CandidateUsecase) AddResume(ctx context.Context, file io.Reader, fileHeader *multipart.FileHeader, userID string) (string, error) {
	// Generate a unique object name for the file in GCS
	objectName := "resumes/" + userID + "/" + fileHeader.Filename

	// Upload the file to GCS
	filePath, err := uc.gcs.UploadResume(ctx, file, objectName)
	if err != nil {
		return "", err
	}

	// Save the file path in the database
	err = uc.repo.AddResumePath(ctx, userID, filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
func (uc *CandidateUsecase) AddResumeFromBytes(ctx context.Context, email string, resumeBytes []byte) (model.GenericResponse, error) {
	// Create a reader from the resume bytes
	file := bytes.NewReader(resumeBytes)

	// Create a multipart file header with appropriate metadata
	fileHeader := &multipart.FileHeader{
		Filename: fmt.Sprintf("%s_resume.pdf", email),
		Size:     int64(len(resumeBytes)),
	}

	// Call the original method
	filePath, err := uc.AddResume(ctx, file, fileHeader, email)
	if err != nil {
		return model.GenericResponse{
			Success: false,
			Message: "Failed to upload resume: " + err.Error(),
		}, err
	}

	return model.GenericResponse{
		Success: true,
		Message: "Resume uploaded successfully: " + filePath,
	}, nil
}

func (u *CandidateUsecase) GoogleLogin(redirectURL string) (string, error) {
	return u.repo.GoogleLogin(redirectURL)
}

func (u *CandidateUsecase) GoogleCallback(code string) (*model.LoginResponse, error) {
	return u.repo.GoogleCallback(code)
}
