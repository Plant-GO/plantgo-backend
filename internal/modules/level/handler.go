package level

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"plantgo-backend/internal/dto"
	"plantgo-backend/internal/modules/level/infrastructure"
)

type PlantService struct {
	plantRepo *infrastructure.PlantRepository
}

func NewPlantService(db *gorm.DB) *PlantService {
	return &PlantService{
		plantRepo: infrastructure.NewPlantRepository(db),
	}
}

// Game Loading Handler

// GetGameDataHandler godoc
// @Summary      Get game data for user
// @Description  Returns user's game progress, completed levels, level reached, and all levels with their status
// @Tags         Game
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} dto.GameDataResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /game/data [get]
func (s *PlantService) GetGameDataHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	id, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
		return
	}

	gameData, err := s.plantRepo.GetGameData(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to fetch game data",
		})
		return
	}

	c.JSON(http.StatusOK, gameData)
}

// GetLevelDetailsHandler godoc
// @Summary      Get level details
// @Description  Returns the riddle and plant name for a specific level when user clicks on it
// @Tags         Game
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Level ID"
// @Success      200 {object} dto.LevelDetailsResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /game/level/{id} [get]
func (s *PlantService) GetLevelDetailsHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	levelIDStr := c.Param("id")
	levelID, err := strconv.ParseUint(levelIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid level ID",
		})
		return
	}

	// Check if user can access this level
	userIDUint, _ := strconv.ParseUint(userID.(string), 10, 32)
	userReward, err := s.plantRepo.GetOrCreateUserReward(uint(userIDUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get user progress",
		})
		return
	}

	if int(levelID) > userReward.LevelReached {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Level not unlocked yet",
		})
		return
	}

	levelDetails, err := s.plantRepo.GetLevelDetails(uint(levelID))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Level not found",
		})
		return
	}

	c.JSON(http.StatusOK, levelDetails)
}

// SubmitAnswerHandler godoc
// @Summary      Submit answer for a level
// @Description  Submits user's answer for a level and returns result with reward if correct
// @Tags         Game
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body dto.SubmitAnswerRequest true "Answer submission"
// @Success      200 {object} dto.SubmitAnswerResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /game/submit-answer [post]
func (s *PlantService) SubmitAnswerHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	var req dto.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request payload",
		})
		return
	}

	userIDUint, _ := strconv.ParseUint(userID.(string), 10, 32)

	// Check if level is already completed
	if s.plantRepo.IsLevelCompleted(uint(userIDUint), req.LevelID) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Level already completed",
		})
		return
	}

	// Get level details
	level, err := s.plantRepo.GetLevelByID(req.LevelID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Level not found",
		})
		return
	}

	// Check if user can access this level
	userReward, err := s.plantRepo.GetOrCreateUserReward(uint(userIDUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get user progress",
		})
		return
	}

	if int(req.LevelID) > userReward.LevelReached {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Level not unlocked yet",
		})
		return
	}

	// Check answer (case-insensitive)
	userAnswer := strings.ToLower(strings.TrimSpace(req.Answer))
	correctAnswer := strings.ToLower(strings.TrimSpace(level.PlantName))
	
	isCorrect := userAnswer == correctAnswer

	response := dto.SubmitAnswerResponse{
		IsCorrect:      isCorrect,
		LevelCompleted: false,
		RewardGained:   0,
		TotalRewards:   userReward.TotalRewards,
	}

	if isCorrect {
		// Complete the level and add reward
		err = s.plantRepo.CompleteLevel(uint(userIDUint), req.LevelID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Failed to complete level",
			})
			return
		}

		// Get updated user reward
		updatedUserReward, err := s.plantRepo.GetUserReward(uint(userIDUint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Failed to get updated rewards",
			})
			return
		}

		response.Message = "Correct! Level completed!"
		response.LevelCompleted = true
		response.RewardGained = level.Reward
		response.TotalRewards = updatedUserReward.TotalRewards
	} else {
		response.Message = "Incorrect answer. Try again!"
		response.CorrectAnswer = level.PlantName
	}

	c.JSON(http.StatusOK, response)
}

// GetUserRewardHandler godoc
// @Summary      Get user reward details
// @Description  Returns user's total rewards and level reached
// @Tags         Game
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} dto.UserRewardResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /game/rewards [get]
func (s *PlantService) GetUserRewardHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	id, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
		return
	}

	userReward, err := s.plantRepo.GetOrCreateUserReward(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to fetch user rewards",
		})
		return
	}

	response := dto.UserRewardResponse{
		ID:           userReward.ID,
		UserID:       userReward.UserID,
		TotalRewards: userReward.TotalRewards,
		LevelReached: userReward.LevelReached,
		CreatedAt:    userReward.CreatedAt,
		UpdatedAt:    userReward.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// Admin Level Management Handlers

// CreateLevelHandler godoc
// @Summary      Create a new level (Admin)
// @Description  Creates a new plant identification level
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body dto.CreateLevelRequest true "Level data"
// @Success      201 {object} dto.LevelResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /admin/levels [post]
func (s *PlantService) CreateLevelHandler(c *gin.Context) {
	var req dto.CreateLevelRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request payload",
		})
		return
	}

	level := &infrastructure.Level{
		Riddle:    req.Riddle,
		PlantName: req.PlantName,
		Reward:    req.Reward,
	}

	if err := s.plantRepo.CreateLevel(level); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create level",
		})
		return
	}

	response := dto.LevelResponse{
		ID:        level.ID,
		Riddle:    level.Riddle,
		PlantName: level.PlantName,
		Reward:    level.Reward,
		CreatedAt: level.CreatedAt,
		UpdatedAt: level.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetAllLevelsHandler godoc
// @Summary      Get all levels (Admin)
// @Description  Retrieves all plant identification levels
// @Tags         Admin
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} dto.LevelListResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /admin/levels [get]
func (s *PlantService) GetAllLevelsHandler(c *gin.Context) {
	levels, err := s.plantRepo.GetAllLevels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to fetch levels",
		})
		return
	}

	total, err := s.plantRepo.GetLevelsCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to count levels",
		})
		return
	}

	levelResponses := make([]dto.LevelResponse, len(levels))
	for i, level := range levels {
		levelResponses[i] = dto.LevelResponse{
			ID:        level.ID,
			Riddle:    level.Riddle,
			PlantName: level.PlantName,
			Reward:    level.Reward,
			CreatedAt: level.CreatedAt,
			UpdatedAt: level.UpdatedAt,
		}
	}

	response := dto.LevelListResponse{
		Levels: levelResponses,
		Total:  total,
	}

	c.JSON(http.StatusOK, response)
}

// GetLevelByIDHandler godoc
// @Summary      Get level by ID (Admin)
// @Description  Retrieves a specific level by its ID
// @Tags         Admin
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Level ID"
// @Success      200 {object} dto.LevelResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /admin/levels/{id} [get]
func (s *PlantService) GetLevelByIDHandler(c *gin.Context) {
	levelIDStr := c.Param("id")
	levelID, err := strconv.ParseUint(levelIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid level ID",
		})
		return
	}

	level, err := s.plantRepo.GetLevelByID(uint(levelID))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Level not found",
		})
		return
	}

	response := dto.LevelResponse{
		ID:        level.ID,
		Riddle:    level.Riddle,
		PlantName: level.PlantName,
		Reward:    level.Reward,
		CreatedAt: level.CreatedAt,
		UpdatedAt: level.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateLevelHandler godoc
// @Summary      Update a level (Admin)
// @Description  Updates an existing level
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Level ID"
// @Param        request body dto.UpdateLevelRequest true "Updated level data"
// @Success      200 {object} dto.LevelResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /admin/levels/{id} [put]
func (s *PlantService) UpdateLevelHandler(c *gin.Context) {
	levelIDStr := c.Param("id")
	levelID, err := strconv.ParseUint(levelIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid level ID",
		})
		return
	}

	var req dto.UpdateLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request payload",
		})
		return
	}

	// Get existing level
	level, err := s.plantRepo.GetLevelByID(uint(levelID))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Level not found",
		})
		return
	}

	// Update fields if provided
	if req.Riddle != "" {
		level.Riddle = req.Riddle
	}
	if req.PlantName != "" {
		level.PlantName = req.PlantName
	}
	if req.Reward > 0 {
		level.Reward = req.Reward
	}

	if err := s.plantRepo.UpdateLevel(level); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update level",
		})
		return
	}

	response := dto.LevelResponse{
		ID:        level.ID,
		Riddle:    level.Riddle,
		PlantName: level.PlantName,
		Reward:    level.Reward,
		CreatedAt: level.CreatedAt,
		UpdatedAt: level.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteLevelHandler godoc
// @Summary      Delete a level (Admin)
// @Description  Deletes an existing level
// @Tags         Admin
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Level ID"
// @Success      200 {object} dto.SuccessResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /admin/levels/{id} [delete]
func (s *PlantService) DeleteLevelHandler(c *gin.Context) {
	levelIDStr := c.Param("id")
	levelID, err := strconv.ParseUint(levelIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid level ID",
		})
		return
	}

	// Check if level exists
	_, err = s.plantRepo.GetLevelByID(uint(levelID))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Level not found",
		})
		return
	}

	if err := s.plantRepo.DeleteLevel(uint(levelID)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to delete level",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Level deleted successfully",
	})
}