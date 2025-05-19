package model

type Candidate struct {
    ID                string        `gorm:"primaryKey;default:gen_random_uuid()"` // Auto-generate UUID
    Email             string        `json:"email"`
    Password          string        `json:"password"`
    Name              string        `json:"name"`
    Phone             int64         `json:"phone"` 
    Experience        int64         `json:"experience"` 
    Skills            []Skills      `json:"skills"`
    Resume            string        `json:"resume"`
    Education         []Education  `json:"education"`
    CurrentLocation   string        `json:"current_location"`
    PreferredLocation string        `json:"preferred_location"`
    Linkedin          string        `json:"linkedin"` 
    Github            string        `json:"github"`   
    ProfilePicture    string        `json:"profile_picture"` 
    IsVerified        bool          `json:"is_verified"`
}

type Skills struct {
	CandidateID string `json:"candidate_id"`
	Skill string `json:"skill"`
	Level string `json:"level"`
}

type Education struct {
	CandidateID string `json:"candidate_id"`
	University string `json:"university"`
	Location string `json:"location"`
	Major string `json:"major"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
	Grade string `json:"grade"`
}

type UpdateCandidateInput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone int64  `json:"phone"`
	Experience int64  `json:"experience"`
	Skills []Skills `gorm:"foreignKey:CandidateID" json:"skills"`
	Education []Education `gorm:"foreignKey:CandidateID" json:"education"`
	CurrentLocation string `json:"current_location"`
	Linkedin string `json:"linkedin"`
	Github string `json:"github"`
	ProfilePicture string `json:"profile_picture"`
	PreferredLocation string `json:"preferred_location"`
}

type Resume struct {
	CandidateID string `json:"candidate_id"`
	FilePath    string `json:"file_path"`
}

type CandidateResume struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	CandidateID string `gorm:"not null"`
	GCSPath     string `gorm:"not null"`
}

type CandidateSignupRequest struct {
    Email    string
    Password string
    Name     string
}

type CandidateLoginRequest struct {
    Email    string
    Password string
}

type EmployerSignupRequest struct {
    Email       string
    Password    string
    CompanyName string
}

type EmployerLoginRequest struct {
    Email    string
    Password string
}

type VerifyEmailRequest struct {
    Email string
    Otp   string
}

type ResendOtpRequest struct {
    Email string
}

type ForgotPasswordRequest struct {
    Email string
}

type ResetPasswordRequest struct {
    Email       string
    NewPassword string
    Otp         string
}

type ChangePasswordRequest struct {
    Email       string
    OldPassword string
    NewPassword string
}

type ProfileRequest struct {
    Email string
}

type ProfileUpdateRequest struct {
    Email string
    Name  string
    // Add more fields as needed, e.g.:
    Phone             int64
    Experience        int64
    CurrentLocation   string
    Linkedin          string
    Github            string
    ProfilePicture    string
    PreferredLocation string
}

type SkillsUpdateRequest struct {
    Email  string
    Skills []string
}

type EducationUpdateRequest struct {
    Email     string
    Education string
}

type UploadResumeRequest struct {
    Email  string
    Resume []byte
}

type GoogleLoginRequest struct {
    RedirectURL string
}

type GoogleCallbackRequest struct {
    Code string
}
