package usecase

import (
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"context"
)

type CandidateUsecase struct {
	candidateRepo repository.CandidateRepository
}

func NewCandidateUsecase(repo repository.CandidateRepository) *CandidateUsecase {
	return &CandidateUsecase{candidateRepo: repo}
}

func (uc *CandidateUsecase) GetProfile(ctx context.Context, userID string) (*model.Candidate, error) {
	return uc.candidateRepo.GetCandidateByUserID(userID)
}
func (uc *CandidateUsecase) UpdateCandidateProfile(ctx context.Context, input *model.UpdateCandidateInput) error {
	return uc.candidateRepo.UpdateCandidate(input)
}
func (uc *CandidateUsecase) AddSkills(ctx context.Context, skills model.Skills) error {
	return uc.candidateRepo.AddSkills(skills)
}
func (uc *CandidateUsecase) AddEducation(ctx context.Context, edu model.Education) error {
	return uc.candidateRepo.AddEducation(edu)
}
func (uc *CandidateUsecase) Signup(input model.SignupRequest) (*model.AuthResponse, error) {
	return uc.candidateRepo.Signup(input)
}
func (uc *CandidateUsecase) Login(input model.LoginRequest) (*model.LoginResponse, error) {
	return uc.candidateRepo.Login(input)
}