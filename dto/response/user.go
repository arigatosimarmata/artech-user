package response

import "time"

// UserResponse represents the user response payload
type UserResponse struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	FullName       string    `json:"full_name"`
	Phone          string    `json:"phone,omitempty"`
	ProfilePicture *string   `json:"profile_picture,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
