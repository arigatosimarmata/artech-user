package request

// RefreshTokenRequest represents the refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
