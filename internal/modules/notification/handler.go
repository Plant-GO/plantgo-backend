package notification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"plantgo-backend/internal/modules/notification/infrastructure"
)

type NotificationHandler struct {
	service *NotificationService
}

func NewNotificationHandler(service *NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetUserNotifications godoc
// @Summary      Get user notifications
// @Description  Retrieves paginated notifications for a user
// @Tags         Notifications
// @Produce      json
// @Param        userId path int true "User ID"
// @Param        limit query int false "Number of notifications per page" default(20)
// @Param        offset query int false "Offset for pagination" default(0)
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{userId} [get]
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	notifications, err := h.service.GetUserNotifications(uint(userID), limit, offset)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to fetch notifications", err)
		return
	}

	h.sendSuccess(c, "Notifications retrieved successfully", notifications)
}

// GetUnreadNotifications godoc
// @Summary      Get unread notifications
// @Description  Retrieves all unread notifications for a user
// @Tags         Notifications
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{userId}/unread [get]
func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	notifications, err := h.service.GetUnreadNotifications(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to fetch unread notifications", err)
		return
	}

	h.sendSuccess(c, "Unread notifications retrieved successfully", notifications)
}

// GetUnreadCount godoc
// @Summary      Get unread notification count
// @Description  Retrieves the count of unread notifications for a user
// @Tags         Notifications
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{userId}/count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	count, err := h.service.GetUnreadNotificationCount(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to fetch unread count", err)
		return
	}

	h.sendSuccess(c, "Unread count retrieved successfully", map[string]interface{}{
		"count": count,
	})
}

// MarkAsRead godoc
// @Summary      Mark notification as read
// @Description  Marks a specific notification as read
// @Tags         Notifications
// @Produce      json
// @Param        id path int true "Notification ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid notification ID", err)
		return
	}

	err = h.service.MarkAsRead(uint(id))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to mark notification as read", err)
		return
	}

	h.sendSuccess(c, "Notification marked as read successfully", nil)
}

// MarkAllAsRead godoc
// @Summary      Mark all notifications as read
// @Description  Marks all notifications as read for a user
// @Tags         Notifications
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{userId}/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	err = h.service.MarkAllAsRead(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to mark all notifications as read", err)
		return
	}

	h.sendSuccess(c, "All notifications marked as read successfully", nil)
}

// DeleteNotification godoc
// @Summary      Delete notification
// @Description  Deletes a specific notification
// @Tags         Notifications
// @Produce      json
// @Param        id path int true "Notification ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid notification ID", err)
		return
	}

	err = h.service.DeleteNotification(uint(id))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to delete notification", err)
		return
	}

	h.sendSuccess(c, "Notification deleted successfully", nil)
}

// UpdateFCMToken godoc
// @Summary      Update FCM token
// @Description  Updates the FCM token for push notifications
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        request body FCMTokenRequest true "FCM token update request"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/fcm-token [post]
func (h *NotificationHandler) UpdateFCMToken(c *gin.Context) {
	var req FCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := h.service.UpdateFCMToken(req.UserID, req.Token)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to update FCM token", err)
		return
	}

	h.sendSuccess(c, "FCM token updated successfully", nil)
}

// GetNotificationsWithFilters godoc
// @Summary      Get user notifications with filters
// @Description  Retrieves paginated notifications for a user with type filtering
// @Tags         Notifications
// @Produce      json
// @Param        userId path int true "User ID"
// @Param        type query string false "Notification type filter" Enums(all,friend_request,game_reward,level_complete,daily_login_reward,weekly_challenge,achievement_unlocked,system_announcement,plant_identified)
// @Param        limit query int false "Number of notifications per page" default(20)
// @Param        offset query int false "Offset for pagination" default(0)
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{userId} [get]
func (h *NotificationHandler) GetNotificationsWithFilters(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse query parameters
	notificationType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	notifications, totalCount, err := h.service.GetNotificationsWithFilters(uint(userID), notificationType, limit, offset)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to fetch notifications", err)
		return
	}

	unreadCount, _ := h.service.GetUnreadNotificationCount(uint(userID))
	hasMore := offset+limit < int(totalCount)

	response := map[string]interface{}{
		"notifications": notifications,
		"totalCount":    totalCount,
		"unreadCount":   unreadCount,
		"hasMore":       hasMore,
	}

	h.sendSuccess(c, "Notifications retrieved successfully", response)
}

// GetUserPreferences godoc
// @Summary      Get user notification preferences
// @Description  Retrieves notification preferences for a user
// @Tags         Notifications
// @Produce      json
// @Param        userId path int true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{userId}/preferences [get]
func (h *NotificationHandler) GetUserPreferences(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	preferences, err := h.service.GetUserPreferences(uint(userID))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to get preferences", err)
		return
	}

	h.sendSuccess(c, "Preferences retrieved successfully", preferences)
}

// UpdateUserPreferences godoc
// @Summary      Update user notification preferences
// @Description  Updates notification preferences for a user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        userId path int true "User ID"
// @Param        preferences body infrastructure.UserNotificationPreference true "User preferences"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /notifications/{userId}/preferences [put]
func (h *NotificationHandler) UpdateUserPreferences(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var preferences infrastructure.UserNotificationPreference
	if err := c.ShouldBindJSON(&preferences); err != nil {
		h.sendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	preferences.UserID = uint(userID)
	err = h.service.UpdateUserPreferences(&preferences)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to update preferences", err)
		return
	}

	h.sendSuccess(c, "Preferences updated successfully", preferences)
}

// Request/Response structures
type FCMTokenRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

type UpdatePreferencesRequest struct {
	FriendRequests      *bool `json:"friend_requests,omitempty"`
	GameRewards         *bool `json:"game_rewards,omitempty"`
	WeeklyChallenges    *bool `json:"weekly_challenges,omitempty"`
	DailyLoginRewards   *bool `json:"daily_login_rewards,omitempty"`
	LevelCompletes      *bool `json:"level_completes,omitempty"`
	AchievementUnlocks  *bool `json:"achievement_unlocks,omitempty"`
	SystemAnnouncements *bool `json:"system_announcements,omitempty"`
	PlantIdentified     *bool `json:"plant_identified,omitempty"`
}

// Helper methods
func (h *NotificationHandler) sendError(c *gin.Context, statusCode int, message string, err error) {
	response := Response{
		Success: false,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	c.JSON(statusCode, response)
}

func (h *NotificationHandler) sendSuccess(c *gin.Context, message string, data interface{}) {
	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}
