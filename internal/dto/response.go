package dto

import (
    "plantgo-backend/internal/modules/auth/infrastructure" 
	"time"
)

type AuthResponse struct {
	Token string    `json:"token"`
	User  infrastructure.User `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}


type LevelResponse struct {
	ID        uint      `json:"id"`
	Riddle    string    `json:"riddle"`
	PlantName string    `json:"plant_name"`
	Reward    int       `json:"reward"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LevelListResponse struct {
	Levels []LevelResponse `json:"levels"`
	Total  int64           `json:"total"`
}

// Game Data DTOs
type GameLevelData struct {
	ID          uint `json:"id"`
	Reward      int  `json:"reward"`
	IsCompleted bool `json:"is_completed"`
	IsUnlocked  bool `json:"is_unlocked"`
}

type GameDataResponse struct {
	UserReward struct {
		ID           uint `json:"id"`
		UserID       uint `json:"user_id"`
		TotalRewards int  `json:"total_rewards"`
		LevelReached int  `json:"level_reached"`
	} `json:"user_reward"`
	Levels          []GameLevelData `json:"levels"`
	CompletedLevels int             `json:"completed_levels"`
	TotalLevels     int             `json:"total_levels"`
}

type LevelDetailsResponse struct {
	ID        uint   `json:"id"`
	Riddle    string `json:"riddle"`
	PlantName string `json:"plant_name"`
	Reward    int    `json:"reward"`
}

// User Progress DTOs
type SubmitAnswerRequest struct {
	LevelID uint   `json:"level_id" binding:"required"`
	Answer  string `json:"answer" binding:"required,min=1,max=255"`
}

type SubmitAnswerResponse struct {
	IsCorrect      bool   `json:"is_correct"`
	Message        string `json:"message"`
	RewardGained   int    `json:"reward_gained"`
	LevelCompleted bool   `json:"level_completed"`
	TotalRewards   int    `json:"total_rewards"`
	CorrectAnswer  string `json:"correct_answer,omitempty"`
}

type UserRewardResponse struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	TotalRewards int       `json:"total_rewards"`
	LevelReached int       `json:"level_reached"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Error DTOs
type LevelErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// Success DTOs
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

