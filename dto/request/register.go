package request

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=3"`
	Phone    string `json:"phone" validate:"omitempty,min=10"`
}
