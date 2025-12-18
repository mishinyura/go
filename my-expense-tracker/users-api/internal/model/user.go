package model

import "time"

// Модель пользователя доменного слоя
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     *string   `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Запросы/ответы для хендлеров (DTO)
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required,min=3" example:"alice"`
	Email    *string `json:"email" example:"alice@example.com"`
}

type UpdateUserRequest struct {
	Username *string `json:"username" example:"alice2"`
	Email    *string `json:"email" example:"alice2@example.com"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     *string   `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
