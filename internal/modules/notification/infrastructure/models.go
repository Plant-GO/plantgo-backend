package infrastructure

import (
	"time"
	"gorm.io/gorm"
)

type NotificationType string

const (
	FriendRequest       NotificationType = "friend_request"
	GameReward          NotificationType = "game_reward"
	WeeklyChallenge     NotificationType = "weekly_challenge"
	DailyLoginReward    NotificationType = "daily_login_reward"
	LevelComplete       NotificationType = "level_complete"
	AchievementUnlocked NotificationType = "achievement_unlocked"
	SystemAnnouncement  NotificationType = "system_announcement"
	PlantIdentified     NotificationType = "plant_identified"
)

type NotificationStatus string

const (
	Pending NotificationStatus = "pending"
	Sent    NotificationStatus = "sent"
	Failed  NotificationStatus = "failed"
	Read    NotificationStatus = "read"
)

type Notification struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	UserID      uint               `json:"user_id" gorm:"not null;index"`
	Type        NotificationType   `json:"type" gorm:"not null"`
	Title       string             `json:"title" gorm:"not null"`
	Message     string             `json:"message" gorm:"not null"`
	Emoji       string             `json:"emoji" gorm:"size:10"`
	Data        string             `json:"data" gorm:"type:text"`
	ActionType  string             `json:"action_type" gorm:"size:50"`
	ActionData  string             `json:"action_data" gorm:"type:text"`
	DeepLinkURL string             `json:"deep_link_url" gorm:"size:255"`
	Status      NotificationStatus `json:"status" gorm:"default:pending"`
	IsRead      bool               `json:"is_read" gorm:"default:false"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	ReadAt      *time.Time         `json:"read_at"`
}

type UserNotificationPreference struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	UserID              uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	FriendRequests      bool      `json:"friend_requests" gorm:"default:true"`
	GameRewards         bool      `json:"game_rewards" gorm:"default:true"`
	WeeklyChallenges    bool      `json:"weekly_challenges" gorm:"default:true"`
	DailyLoginRewards   bool      `json:"daily_login_rewards" gorm:"default:true"`
	LevelCompletes      bool      `json:"level_completes" gorm:"default:true"`
	AchievementUnlocks  bool      `json:"achievement_unlocks" gorm:"default:true"`
	SystemAnnouncements bool      `json:"system_announcements" gorm:"default:true"`
	PlantIdentified     bool      `json:"plant_identified" gorm:"default:true"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type UserFCMToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (UserNotificationPreference) TableName() string {
	return "user_notification_preferences"
}

func (UserFCMToken) TableName() string {
	return "user_fcm_tokens"
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	if n.UpdatedAt.IsZero() {
		n.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (n *Notification) BeforeUpdate(tx *gorm.DB) error {
	n.UpdatedAt = time.Now().UTC()
	return nil
}

func (unp *UserNotificationPreference) BeforeCreate(tx *gorm.DB) error {
	if unp.CreatedAt.IsZero() {
		unp.CreatedAt = time.Now().UTC()
	}
	if unp.UpdatedAt.IsZero() {
		unp.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (unp *UserNotificationPreference) BeforeUpdate(tx *gorm.DB) error {
	unp.UpdatedAt = time.Now().UTC()
	return nil
}

func (uft *UserFCMToken) BeforeCreate(tx *gorm.DB) error {
	if uft.CreatedAt.IsZero() {
		uft.CreatedAt = time.Now().UTC()
	}
	if uft.UpdatedAt.IsZero() {
		uft.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (uft *UserFCMToken) BeforeUpdate(tx *gorm.DB) error {
	uft.UpdatedAt = time.Now().UTC()
	return nil
}
