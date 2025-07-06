package notification

import (
	"encoding/json"
	"fmt"
	"log"
	"plantgo-backend/internal/modules/notification/infrastructure"
)

type NotificationService struct {
	repo            *infrastructure.NotificationRepository
	firebaseService *FirebaseService
}

func NewNotificationService(repo *infrastructure.NotificationRepository, firebaseService *FirebaseService) *NotificationService {
	return &NotificationService{
		repo:            repo,
		firebaseService: firebaseService,
	}
}

type NotificationData struct {
	LevelID       *uint                  `json:"level_id,omitempty"`
	LevelNumber   *int                   `json:"level_number,omitempty"`
	Reward        *int                   `json:"reward,omitempty"`
	FriendID      *uint                  `json:"friend_id,omitempty"`
	AchievementID *uint                  `json:"achievement_id,omitempty"`
	PlantName     *string                `json:"plant_name,omitempty"`
	Confidence    *float64               `json:"confidence,omitempty"`
	ExtraData     map[string]interface{} `json:"extra_data,omitempty"`
}

// Helper method to create notification and send push notification
func (s *NotificationService) createAndSendNotification(notification *infrastructure.Notification) error {
	// Create notification in database
	if err := s.repo.CreateNotification(notification); err != nil {
		return err
	}

	// Send push notification via Firebase
	if s.firebaseService != nil {
		go func() {
			if err := s.firebaseService.SendPushNotification(notification); err != nil {
				log.Printf("Failed to send push notification: %v", err)
			}
		}()
	}

	return nil
}

// Generate different types of notifications
func (s *NotificationService) GenerateLevelCompleteNotification(userID uint, levelNumber int, reward int) error {
	// Check if user has this notification type enabled
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.LevelComplete)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil // User has disabled this notification type
	}

	data := NotificationData{
		LevelNumber: &levelNumber,
		Reward:      &reward,
		ExtraData: map[string]interface{}{
			"level_number": levelNumber,
			"reward_type":  "level_completion",
		},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.LevelComplete,
		Title:   "Level Complete! 🎉",
		Message: fmt.Sprintf("Congratulations! You completed level %d and earned %d coins!", levelNumber, reward),
		Data:    string(dataJSON),
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

func (s *NotificationService) GenerateDailyLoginReward(userID uint, reward int, streak int) error {
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.DailyLoginReward)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil
	}

	data := NotificationData{
		Reward: &reward,
		ExtraData: map[string]interface{}{
			"streak":      streak,
			"reward_type": "daily_login",
		},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.DailyLoginReward,
		Title:   "Daily Login Reward! 🌟",
		Message: fmt.Sprintf("Welcome back! You've earned %d coins. Login streak: %d days!", reward, streak),
		Data:    string(dataJSON),
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

func (s *NotificationService) GenerateWeeklyChallengeComplete(userID uint, challengeName string, reward int) error {
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.WeeklyChallenge)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil
	}

	data := NotificationData{
		Reward: &reward,
		ExtraData: map[string]interface{}{
			"challenge_name": challengeName,
			"reward_type":    "weekly_challenge",
		},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.WeeklyChallenge,
		Title:   "Weekly Challenge Complete! 🏆",
		Message: fmt.Sprintf("Amazing! You completed '%s' and earned %d coins!", challengeName, reward),
		Data:    string(dataJSON),
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

func (s *NotificationService) GenerateFriendRequestNotification(userID uint, fromUserID uint, fromUsername string) error {
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.FriendRequest)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil
	}

	data := NotificationData{
		FriendID: &fromUserID,
		ExtraData: map[string]interface{}{
			"from_username": fromUsername,
			"action":        "friend_request",
		},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.FriendRequest,
		Title:   "New Friend Request! 👥",
		Message: fmt.Sprintf("%s wants to be your friend in PlantGo!", fromUsername),
		Data:    string(dataJSON),
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

func (s *NotificationService) GenerateAchievementUnlocked(userID uint, achievementName string, reward int) error {
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.AchievementUnlocked)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil
	}

	data := NotificationData{
		Reward: &reward,
		ExtraData: map[string]interface{}{
			"achievement_name": achievementName,
			"reward_type":      "achievement",
		},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.AchievementUnlocked,
		Title:   "Achievement Unlocked! 🏅",
		Message: fmt.Sprintf("Congratulations! You unlocked '%s' and earned %d coins!", achievementName, reward),
		Data:    string(dataJSON),
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

func (s *NotificationService) GenerateSystemAnnouncement(userID uint, title, message string) error {
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.SystemAnnouncement)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil
	}

	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.SystemAnnouncement,
		Title:   title,
		Message: message,
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

func (s *NotificationService) GeneratePlantIdentifiedNotification(userID uint, plantName string, confidence float64) error {
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.PlantIdentified)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil
	}

	data := NotificationData{
		PlantName:  &plantName,
		Confidence: &confidence,
		ExtraData: map[string]interface{}{
			"plant_name": plantName,
			"confidence": confidence,
		},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.PlantIdentified,
		Title:   "Plant Identified! 🌿",
		Message: fmt.Sprintf("Great! We identified '%s' with %.2f%% confidence!", plantName, confidence*100),
		Data:    string(dataJSON),
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

func (s *NotificationService) GenerateGameRewardNotification(userID uint, rewardType string, reward int, description string) error {
	enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.GameReward)
	if err != nil {
		log.Printf("Error checking notification preferences: %v", err)
		return err
	}
	if !enabled {
		return nil
	}

	data := NotificationData{
		Reward: &reward,
		ExtraData: map[string]interface{}{
			"reward_type": rewardType,
			"description": description,
		},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	notification := &infrastructure.Notification{
		UserID:  userID,
		Type:    infrastructure.GameReward,
		Title:   "Game Reward! 💰",
		Message: fmt.Sprintf("You earned %d coins from %s!", reward, description),
		Data:    string(dataJSON),
		Status:  infrastructure.Pending,
	}
	
	return s.createAndSendNotification(notification)
}

// Bulk notification generation for system announcements
func (s *NotificationService) GenerateBulkSystemAnnouncement(userIDs []uint, title, message string) error {
	notifications := make([]*infrastructure.Notification, 0, len(userIDs))
	
	for _, userID := range userIDs {
		enabled, err := s.repo.IsNotificationTypeEnabled(userID, infrastructure.SystemAnnouncement)
		if err != nil {
			log.Printf("Error checking notification preferences for user %d: %v", userID, err)
			continue
		}
		if !enabled {
			continue
		}

		notification := &infrastructure.Notification{
			UserID:  userID,
			Type:    infrastructure.SystemAnnouncement,
			Title:   title,
			Message: message,
			Status:  infrastructure.Pending,
		}
		notifications = append(notifications, notification)
	}
	
	// Bulk create notifications
	for _, notification := range notifications {
		if err := s.repo.CreateNotification(notification); err != nil {
			log.Printf("Error creating bulk notification for user %d: %v", notification.UserID, err)
		}
	}
	
	return nil
}

// Helper methods for retrieving notifications
func (s *NotificationService) GetUserNotifications(userID uint, limit, offset int) ([]infrastructure.Notification, error) {
	return s.repo.GetUserNotifications(userID, limit, offset)
}

func (s *NotificationService) GetUnreadNotifications(userID uint) ([]infrastructure.Notification, error) {
	return s.repo.GetUnreadNotifications(userID)
}

func (s *NotificationService) GetUnreadNotificationCount(userID uint) (int64, error) {
	return s.repo.GetUnreadNotificationCount(userID)
}

func (s *NotificationService) MarkAsRead(notificationID uint) error {
	return s.repo.MarkAsRead(notificationID)
}

func (s *NotificationService) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}

func (s *NotificationService) DeleteNotification(notificationID uint) error {
	return s.repo.DeleteNotification(notificationID)
}

func (s *NotificationService) UpdateFCMToken(userID uint, token string) error {
	return s.repo.UpsertFCMToken(userID, token)
}

func (s *NotificationService) GetUserPreferences(userID uint) (*infrastructure.UserNotificationPreference, error) {
	return s.repo.GetUserPreferences(userID)
}

func (s *NotificationService) UpdateUserPreferences(prefs *infrastructure.UserNotificationPreference) error {
	return s.repo.UpdateUserPreferences(prefs)
}

func (s *NotificationService) GetNotificationsByType(userID uint, notificationType infrastructure.NotificationType, limit int) ([]infrastructure.Notification, error) {
	return s.repo.GetNotificationsByType(userID, notificationType, limit)
}

func (s *NotificationService) GetNotificationsWithFilters(userID uint, notificationType string, limit, offset int) ([]infrastructure.Notification, int64, error) {
	return s.repo.GetNotificationsWithFilters(userID, notificationType, limit, offset)
}

func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	return s.repo.GetUnreadCount(userID)
}
