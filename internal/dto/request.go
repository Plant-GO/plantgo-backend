package dto



type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GuestLoginRequest struct {
	AndroidID string `json:"android_id" binding:"required"`
	Username  string `json:"username" binding:"required"`
}

// Level DTOs
type CreateLevelRequest struct {
	Riddle    string `json:"riddle" binding:"required,min=10,max=500"`
	PlantName string `json:"plant_name" binding:"required,min=2,max=255"`
	Reward    int    `json:"reward" binding:"required,min=1"`
}

type UpdateLevelRequest struct {
	Riddle    string `json:"riddle" binding:"omitempty,min=10,max=500"`
	PlantName string `json:"plant_name" binding:"omitempty,min=2,max=255"`
	Reward    int    `json:"reward" binding:"omitempty,min=1"`
}
