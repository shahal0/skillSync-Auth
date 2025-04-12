package domain



type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "candidate" or "employer"
}

type AuthResponse struct {
	ID string `json:"id"`
	Message string `json:"message"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "candidate" or "employer"
}
type LoginResponse struct {
	ID       string    `json:"id"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}
