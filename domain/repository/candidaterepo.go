package repository

import model "skillsync-authservice/domain/models"

type CandidateRepository interface {
	Signup(input model.SignupRequest) (*model.AuthResponse, error)
	Login(input model.LoginRequest) (*model.LoginResponse, error)
	CreateCandidate(profile *model.Candidate)(int,error)
	UpdateCandidate(profile *model.UpdateCandidateInput,userid string) error
	GetCandidateByUserID(userID string) (*model.Candidate, error)
	GetCandidateByEmail(email string) (*model.Candidate, error)
	AddSkills(skills model.Skills) error
	AddEducation(edu model.Education,userid string) error
	VerifyEmail(email string,otp uint64) error
	ResendOtp(email string) error
	SendPasswordResetOtp(email string) error
	VerifyPasswordResetOtp(email string, otp uint64) error
	UpdatePassword(email string, newPassword string) error
	UpdatePasswordByID(id string, newPassword string) error
}
