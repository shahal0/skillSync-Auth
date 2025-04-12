package repository

import model "skillsync-authservice/domain/models"

type CandidateRepository interface {
	Signup(input model.SignupRequest) (*model.AuthResponse, error)
	Login(input model.LoginRequest) (*model.LoginResponse, error)
	CreateCandidate(profile *model.Candidate)(int,error) 
	UpdateCandidate(profile *model.UpdateCandidateInput) error
	GetCandidateByUserID(userID string) (*model.Candidate, error)
	GetCandidateByEmail(email string) (*model.Employer, error)
	AddSkills(skills model.Skills) error
	AddEducation(edu model.Education) error
}
