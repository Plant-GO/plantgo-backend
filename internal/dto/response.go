package dto

import (
    "plantgo-backend/internal/modules/auth/infrastructure" 
)

type AuthResponse struct {
	Token string    `json:"token"`
	User  infrastructure.User `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
