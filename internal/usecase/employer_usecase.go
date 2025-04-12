package usecase

import (
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"context"
)

type EmployerUsecase struct {
	employerRepo repository.EmployerRepository
}

func NewEmployerUsecase(repo repository.EmployerRepository) *EmployerUsecase {
	return &EmployerUsecase{employerRepo: repo}
}

func (uc *EmployerUsecase) UpdateProfile(ctx context.Context, profile *model.UpdateEmployerInput) error {
	return uc.employerRepo.UpdateEmployer(profile)
}

func (uc *EmployerUsecase) GetProfile(ctx context.Context, userID string) (*model.Employer, error) {
	return uc.employerRepo.GetEmployerByUserID(userID)
}
func (uc *EmployerUsecase) Signup(input model.SignupRequest) (*model.AuthResponse, error) {
	return uc.employerRepo.Signup(input)
}
func (uc *EmployerUsecase) Login(input model.LoginRequest) (*model.LoginResponse, error) {
	return uc.employerRepo.Login(input)
}

