package model

type SignupRequest struct {
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Password string            `json:"password"`
	Role     string            `json:"role"`              // "candidate" or "employer"
	Context  map[string]string `json:"context,omitempty"` // Additional fields like phone, industry, location, website
}

type AuthResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "candidate" or "employer"
}

type LoginResponse struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type VerificationTable struct {
	Email              string `validate:"required,email" gorm:"type:varchar(255);unique_index"`
	Role               string
	OTP                uint64
	OTPExpiry          uint64
	VerificationStatus bool
}
