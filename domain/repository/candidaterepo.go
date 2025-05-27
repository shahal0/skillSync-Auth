package repository

import (
	"context"
	"io"
	model "skillsync-authservice/domain/models"
)

type CandidateRepository interface {
	Signup(input model.SignupRequest) (*model.AuthResponse, error)
	Login(input model.LoginRequest) (*model.LoginResponse, error)
	GoogleLogin(redirectURL string) (string, error)
    GoogleCallback(code string) (*model.LoginResponse, error)
	CreateCandidate(profile *model.Candidate)(int,error)
	UpdateCandidate(profile *model.UpdateCandidateInput,userid string) error
	GetCandidateByUserID(userID string) (*model.Candidate, error)
	GetCandidateByEmail(email string) (*model.Candidate, error)
	VerifyToken(token string) (string,string, error)
	AddSkills(skills model.Skills,userid string) error
	GetSkills(candidateID string) ([]string, error)
	AddEducation(edu model.Education,userid string) error
	VerifyEmail(email string,otp uint64) error
	ResendOtp(email string) error
	SendPasswordResetOtp(email string) error
	VerifyPasswordResetOtp(email string, otp uint64) error
	UpdatePassword(email string, newPassword string) error
	UpdatePasswordByID(id string, newPassword string) error
	AddResume(ctx context.Context, objectName string, file io.Reader) (string, error)
	AddResumePath(ctx context.Context,  userID string, filepath string) error
}
