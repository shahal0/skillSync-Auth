package usecase

import model "skillsync-authservice/domain/models"

type ProfileUsecase interface {
	UpdateEmployerProfile(userID string, updatedData *model.Employer) error
	UpdateCandidateProfile(userID string, updatedData *model.Candidate) error
	GetEmployerProfile(userID string) (*model.Employer, error)
	GetCandidateProfile(userID string) (*model.Candidate, error)
    AddSkills(skills model.Skills) error
	AddEducation(edu model.Education) error
}
