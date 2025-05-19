package repository

import model "skillsync-authservice/domain/models"

type EmployerRepository interface {
	Signup(input model.SignupRequest) (*model.AuthResponse, error)
	Login(input model.LoginRequest) (*model.LoginResponse, error)
	GoogleLogin(redirectURL string) (string, error)
    GoogleCallback(code string) (*model.LoginResponse, error)
	CreateEmployer(profile *model.Employer) (int,error)
	UpdateEmployer(profile *model.UpdateEmployerInput) error
	GetEmployerByUserID(userID string) (*model.Employer, error)
	GetEmployerByEmail(email string) (*model.Employer, error)
	VerifyEmail(email string, otp uint64) error
	ResendOtp(email string) error
	SendPasswordResetOtp(email string) error
	VerifyPasswordResetOtp(email string, otp uint64) error
	UpdatePassword(email string, newPassword string) error
	UpdatePasswordByID(id string, newPassword string) error

}
