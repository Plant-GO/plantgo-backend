// internal/modules/plant/handlers.go
package level

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"plantgo-backend/internal/modules/level/infrastructure"
	"plantgo-backend/internal/modules/notification"
)

type PlantHandler struct {
	repository          *infrastructure.PlantRepository
	notificationService *notification.NotificationService
}

func NewPlantHandler(repository *infrastructure.PlantRepository, notificationService *notification.NotificationService) *PlantHandler {
	return &PlantHandler{
		repository:          repository,
		notificationService: notificationService,
	}
}

// Response structures
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type LevelRequest struct {
	LevelNumber int    `json:"level_number"`
	Riddle      string `json:"riddle"`
	PlantName   string `json:"plant_name"`
	Reward      int    `json:"reward"`
}

type CompleteLevelRequest struct {
	UserID  uint `json:"user_id"`
	LevelID uint `json:"level_id"`
}

type CompleteLevelByNumberRequest struct {
	UserID      uint `json:"user_id"`
	LevelNumber int  `json:"level_number"`
}

// Helper functions
func (h *PlantHandler) sendError(c *gin.Context, statusCode int, message string, err error) {
	response := Response{
		Success: false,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	c.JSON(statusCode, response)
}

func (h *PlantHandler) sendSuccess(c *gin.Context, message string, data interface{}) {
	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}

// CreateLevel godoc
// @Summary      Create a new level
// @Description  Creates a new level with riddle, plant name, and reward
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        request body LevelRequest true "Level creation info"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /admin/levels [post]
func (h *PlantHandler) CreateLevel(c *gin.Context) {
	var req LevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if req.LevelNumber <= 0 {
		h.sendError(c, http.StatusBadRequest, "Level number must be greater than 0", nil)
		return
	}
	if strings.TrimSpace(req.Riddle) == "" {
		h.sendError(c, http.StatusBadRequest, "Riddle cannot be empty", nil)
		return
	}
	if strings.TrimSpace(req.PlantName) == "" {
		h.sendError(c, http.StatusBadRequest, "Plant name cannot be empty", nil)
		return
	}

	level := &infrastructure.Level{
		LevelNumber: req.LevelNumber,
		Riddle:      strings.TrimSpace(req.Riddle),
		PlantName:   strings.TrimSpace(req.PlantName),
		Reward:      req.Reward,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := h.repository.CreateLevel(level); err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to create level", err)
		return
	}

	h.sendSuccess(c, "Level created successfully", level)
}

// GetLevel godoc
// @Summary      Get level by ID
// @Description  Retrieves a level by its ID
// @Tags         Level
// @Produce      json
// @Param        id path int true "Level ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Router       /levels/{id} [get]
func (h *PlantHandler) GetLevel(c *gin.Context) {
	idStr := c.Param("id")
	
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid level ID", err)
		return
	}

	level, err := h.repository.GetLevelByID(uint(id))
	if err != nil {
		h.sendError(c, http.StatusNotFound, "Level not found", err)
		return
	}

	h.sendSuccess(c, "Level retrieved successfully", level)
}

// GetLevelByNumber godoc
// @Summary      Get level by number
// @Description  Retrieves a level by its level number
// @Tags         Level
// @Produce      json
// @Param        number path int true "Level Number"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Router       /levels/number/{number} [get]
func (h *PlantHandler) GetLevelByNumber(c *gin.Context) {
	levelNumberStr := c.Param("number")
	
	levelNumber, err := strconv.Atoi(levelNumberStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid level number", err)
		return
	}

	level, err := h.repository.GetLevelByNumber(levelNumber)
	if err != nil {
		h.sendError(c, http.StatusNotFound, "Level not found", err)
		return
	}

	h.sendSuccess(c, "Level retrieved successfully", level)
}

// GetAllLevels godoc
// @Summary      Get all levels
// @Description  Retrieves all levels in the system
// @Tags         Level
// @Produce      json
// @Success      200 {object} Response
// @Failure      500 {object} Response
// @Router       /levels [get]
func (h *PlantHandler) GetAllLevels(c *gin.Context) {
	levels, err := h.repository.GetAllLevels()
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to retrieve levels", err)
		return
	}

	h.sendSuccess(c, "Levels retrieved successfully", levels)
}

// UpdateLevel godoc
// @Summary      Update level
// @Description  Updates an existing level by ID
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id path int true "Level ID"
// @Param        request body LevelRequest true "Level update info"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Failure      500 {object} Response
// @Router       /admin/levels/{id} [put]
func (h *PlantHandler) UpdateLevel(c *gin.Context) {
	idStr := c.Param("id")
	
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid level ID", err)
		return
	}

	// Get existing level
	existingLevel, err := h.repository.GetLevelByID(uint(id))
	if err != nil {
		h.sendError(c, http.StatusNotFound, "Level not found", err)
		return
	}

	var req LevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Update fields if provided
	if req.LevelNumber > 0 {
		existingLevel.LevelNumber = req.LevelNumber
	}
	if strings.TrimSpace(req.Riddle) != "" {
		existingLevel.Riddle = strings.TrimSpace(req.Riddle)
	}
	if strings.TrimSpace(req.PlantName) != "" {
		existingLevel.PlantName = strings.TrimSpace(req.PlantName)
	}
	if req.Reward >= 0 {
		existingLevel.Reward = req.Reward
	}
	existingLevel.UpdatedAt = time.Now().UTC()

	if err := h.repository.UpdateLevel(existingLevel); err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to update level", err)
		return
	}

	h.sendSuccess(c, "Level updated successfully", existingLevel)
}

// DeleteLevel godoc
// @Summary      Delete level
// @Description  Deletes a level by ID
// @Tags         Admin
// @Produce      json
// @Param        id path int true "Level ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Failure      500 {object} Response
// @Router       /admin/levels/{id} [delete]
func (h *PlantHandler) DeleteLevel(c *gin.Context) {
	idStr := c.Param("id")
	
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid level ID", err)
		return
	}

	// Check if level exists
	_, err = h.repository.GetLevelByID(uint(id))
	if err != nil {
		h.sendError(c, http.StatusNotFound, "Level not found", err)
		return
	}

	if err := h.repository.DeleteLevel(uint(id)); err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to delete level", err)
		return
	}

	h.sendSuccess(c, "Level deleted successfully", nil)
}

// GetUserProgress godoc
// @Summary      Get user progress
// @Description  Retrieves the progress of a user across all levels
// @Tags         Game
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /game/progress/{userId} [get]
func (h *PlantHandler) GetUserProgress(c *gin.Context) {
	userIDStr := c.Param("userId")
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	progress, err := h.repository.GetUserProgress(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to retrieve user progress", err)
		return
	}

	h.sendSuccess(c, "User progress retrieved successfully", progress)
}

// GetCompletedLevels godoc
// @Summary      Get completed levels
// @Description  Retrieves all levels completed by a user
// @Tags         Game
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /game/completed/{userId} [get]
func (h *PlantHandler) GetCompletedLevels(c *gin.Context) {
	userIDStr := c.Param("userId")
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	completedLevels, err := h.repository.GetCompletedLevels(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to retrieve completed levels", err)
		return
	}

	h.sendSuccess(c, "Completed levels retrieved successfully", completedLevels)
}

// CompleteLevel godoc
// @Summary      Complete level
// @Description  Marks a level as completed for a user
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param        request body CompleteLevelRequest true "Level completion info"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Failure      409 {object} Response
// @Failure      500 {object} Response
// @Router       /game/complete [post]
func (h *PlantHandler) CompleteLevel(c *gin.Context) {
	var req CompleteLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.UserID == 0 || req.LevelID == 0 {
		h.sendError(c, http.StatusBadRequest, "User ID and Level ID are required", nil)
		return
	}

	// Check if level exists
	level, err := h.repository.GetLevelByID(req.LevelID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, "Level not found", err)
		return
	}

	// Check if already completed
	if h.repository.IsLevelCompleted(req.UserID, req.LevelID) {
		h.sendError(c, http.StatusConflict, "Level already completed", nil)
		return
	}

	if err := h.repository.CompleteLevel(req.UserID, req.LevelID); err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to complete level", err)
		return
	}

	// Generate notification for level completion
	if h.notificationService != nil {
		err := h.notificationService.GenerateLevelCompleteNotification(
			req.UserID,
			level.LevelNumber,
			level.Reward,
		)
		if err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to generate level completion notification: %v", err)
		}
	}

	responseData := map[string]interface{}{
		"user_id":      req.UserID,
		"level_id":     req.LevelID,
		"level_number": level.LevelNumber,
		"reward":       level.Reward,
		"completed_at": time.Now().UTC(),
	}

	h.sendSuccess(c, "Level completed successfully", responseData)
}

// CompleteLevelByNumber godoc
// @Summary      Complete level by number
// @Description  Marks a level as completed for a user using level number
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param        request body CompleteLevelByNumberRequest true "Level completion info"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Failure      409 {object} Response
// @Failure      500 {object} Response
// @Router       /game/complete-by-number [post]
func (h *PlantHandler) CompleteLevelByNumber(c *gin.Context) {
	var req CompleteLevelByNumberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.UserID == 0 || req.LevelNumber <= 0 {
		h.sendError(c, http.StatusBadRequest, "User ID and Level Number are required", nil)
		return
	}

	// Get level by number
	level, err := h.repository.GetLevelByNumber(req.LevelNumber)
	if err != nil {
		h.sendError(c, http.StatusNotFound, "Level not found", err)
		return
	}

	// Check if already completed
	if h.repository.IsLevelCompletedByNumber(req.UserID, req.LevelNumber) {
		h.sendError(c, http.StatusConflict, "Level already completed", nil)
		return
	}

	if err := h.repository.CompleteLevel(req.UserID, level.ID); err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to complete level", err)
		return
	}

	// Generate notification for level completion
	if h.notificationService != nil {
		err := h.notificationService.GenerateLevelCompleteNotification(
			req.UserID,
			level.LevelNumber,
			level.Reward,
		)
		if err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to generate level completion notification: %v", err)
		}
	}

	responseData := map[string]interface{}{
		"user_id":      req.UserID,
		"level_id":     level.ID,
		"level_number": level.LevelNumber,
		"reward":       level.Reward,
		"completed_at": time.Now().UTC(),
	}

	h.sendSuccess(c, "Level completed successfully", responseData)
}

// GetUserReward godoc
// @Summary      Get user reward
// @Description  Retrieves the total reward points for a user
// @Tags         Game
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /game/rewards/{userId} [get]
func (h *PlantHandler) GetUserReward(c *gin.Context) {
	userIDStr := c.Param("userId")
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	reward, err := h.repository.GetOrCreateUserReward(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to retrieve user reward", err)
		return
	}

	h.sendSuccess(c, "User reward retrieved successfully", reward)
}

// GetLevelDetails godoc
// @Summary      Get level details
// @Description  Retrieves detailed information about a level for a specific user
// @Tags         Game
// @Produce      json
// @Param        userId path int true "User ID"
// @Param        number path int true "Level Number"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Router       /game/level/{userId}/{number} [get]
func (h *PlantHandler) GetLevelDetails(c *gin.Context) {
	userIDStr := c.Param("userId")
	levelNumberStr := c.Param("number")
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	levelNumber, err := strconv.Atoi(levelNumberStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid level number", err)
		return
	}

	levelDetails, err := h.repository.GetLevelDetailsByNumber(uint(userID), levelNumber)
	if err != nil {
		h.sendError(c, http.StatusNotFound, "Level details not found", err)
		return
	}

	h.sendSuccess(c, "Level details retrieved successfully", levelDetails)
}

// GetGameData godoc
// @Summary      Get game data
// @Description  Retrieves comprehensive game data for a user including progress and rewards
// @Tags         Game
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /game/data/{userId} [get]
func (h *PlantHandler) GetGameData(c *gin.Context) {
	userIDStr := c.Param("userId")
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	gameData, err := h.repository.GetGameData(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to retrieve game data", err)
		return
	}

	h.sendSuccess(c, "Game data retrieved successfully", gameData)
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns the health status of the service
// @Tags         System
// @Produce      json
// @Success      200 {object} Response
// @Failure      500 {object} Response
// @Router       /plant/health [get]
func (h *PlantHandler) HealthCheck(c *gin.Context) {
	// Get total levels count
	count, err := h.repository.GetLevelsCount()
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Health check failed", err)
		return
	}

	healthData := map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now().UTC(),
		"total_levels": count,
	}

	h.sendSuccess(c, "Service is healthy", healthData)
}