package infrastructure

import (
	"errors"
	"time"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Notification CRUD operations
func (r *NotificationRepository) CreateNotification(notification *Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) GetNotificationByID(id uint) (*Notification, error) {
	var notification Notification
	err := r.db.Where("id = ?", id).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) GetUserNotifications(userID uint, limit, offset int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) GetUnreadNotifications(userID uint) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("user_id = ? AND is_read = false", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkAsRead(notificationID uint) error {
	now := time.Now().UTC()
	return r.db.Model(&Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	now := time.Now().UTC()
	return r.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *NotificationRepository) UpdateNotificationStatus(notificationID uint, status NotificationStatus) error {
	return r.db.Model(&Notification{}).
		Where("id = ?", notificationID).
		Update("status", status).Error
}

func (r *NotificationRepository) GetPendingNotifications(limit int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("status = ?", Pending).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) DeleteNotification(notificationID uint) error {
	return r.db.Delete(&Notification{}, notificationID).Error
}

func (r *NotificationRepository) GetNotificationsByType(userID uint, notificationType NotificationType, limit int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("user_id = ? AND type = ?", userID, notificationType).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// FCM Token management
func (r *NotificationRepository) UpsertFCMToken(userID uint, token string) error {
	var fcmToken UserFCMToken
	err := r.db.Where("user_id = ?", userID).First(&fcmToken).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new token
			fcmToken = UserFCMToken{
				UserID:   userID,
				Token:    token,
				IsActive: true,
			}
			return r.db.Create(&fcmToken).Error
		}
		return err
	}
	
	// Update existing token
	fcmToken.Token = token
	fcmToken.IsActive = true
	return r.db.Save(&fcmToken).Error
}

func (r *NotificationRepository) GetUserFCMToken(userID uint) (string, error) {
	var fcmToken UserFCMToken
	err := r.db.Where("user_id = ? AND is_active = true", userID).First(&fcmToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("FCM token not found")
		}
		return "", err
	}
	return fcmToken.Token, nil
}

func (r *NotificationRepository) DeactivateFCMToken(userID uint) error {
	return r.db.Model(&UserFCMToken{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error
}

func (r *NotificationRepository) GetAllUserFCMTokens(userIDs []uint) ([]UserFCMToken, error) {
	var tokens []UserFCMToken
	err := r.db.Where("user_id IN ? AND is_active = true", userIDs).Find(&tokens).Error
	return tokens, err
}

// Notification Preferences
func (r *NotificationRepository) GetUserPreferences(userID uint) (*UserNotificationPreference, error) {
	var prefs UserNotificationPreference
	err := r.db.Where("user_id = ?", userID).First(&prefs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default preferences
			prefs = UserNotificationPreference{
				UserID:              userID,
				FriendRequests:      true,
				GameRewards:         true,
				WeeklyChallenges:    true,
				DailyLoginRewards:   true,
				LevelCompletes:      true,
				AchievementUnlocks:  true,
				SystemAnnouncements: true,
				PlantIdentified:     true,
			}
			err = r.db.Create(&prefs).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &prefs, nil
}

func (r *NotificationRepository) UpdateUserPreferences(prefs *UserNotificationPreference) error {
	return r.db.Save(prefs).Error
}

func (r *NotificationRepository) IsNotificationTypeEnabled(userID uint, notificationType NotificationType) (bool, error) {
	prefs, err := r.GetUserPreferences(userID)
	if err != nil {
		return false, err
	}

	switch notificationType {
	case FriendRequest:
		return prefs.FriendRequests, nil
	case GameReward:
		return prefs.GameRewards, nil
	case WeeklyChallenge:
		return prefs.WeeklyChallenges, nil
	case DailyLoginReward:
		return prefs.DailyLoginRewards, nil
	case LevelComplete:
		return prefs.LevelCompletes, nil
	case AchievementUnlocked:
		return prefs.AchievementUnlocks, nil
	case SystemAnnouncement:
		return prefs.SystemAnnouncements, nil
	case PlantIdentified:
		return prefs.PlantIdentified, nil
	default:
		return true, nil
	}
}

func (r *NotificationRepository) GetNotificationsWithFilters(userID uint, notificationType string, limit, offset int) ([]Notification, int64, error) {
	var notifications []Notification
	var totalCount int64
	
	query := r.db.Where("user_id = ?", userID)
	
	if notificationType != "" && notificationType != "all" {
		query = query.Where("type = ?", notificationType)
	}
	
	// Get total count
	err := query.Model(&Notification{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	
	return notifications, totalCount, err
}

func (r *NotificationRepository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}
