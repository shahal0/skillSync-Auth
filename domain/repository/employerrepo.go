package repository

import model "skillsync-authservice/domain/models"

type EmployerRepository interface {
	Signup(input model.SignupRequest) (*model.AuthResponse, error)
	Login(input model.LoginRequest) (*model.LoginResponse, error)
	CreateEmployer(profile *model.Employer) (int,error)
	UpdateEmployer(profile *model.UpdateEmployerInput) error
	GetEmployerByUserID(userID string) (*model.Employer, error)
	GetEmployerByEmail(email string) (*model.Employer, error)
}
