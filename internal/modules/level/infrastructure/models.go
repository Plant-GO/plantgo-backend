// infrastructure/models.go
package infrastructure

import (
	"time"
	"gorm.io/gorm"
)

type Level struct {
	ID          uint           `json:"id" gorm:"primaryKey" db:"id"`
	LevelNumber int            `json:"level_number" gorm:"not null;uniqueIndex" db:"level_number"`
	Riddle      string         `json:"riddle" gorm:"not null;size:500" db:"riddle"`
	PlantName   string         `json:"plant_name" gorm:"not null;size:255" db:"plant_name"`
	Reward      int            `json:"reward" gorm:"not null;default:0" db:"reward"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Level) TableName() string {
	return "levels"
}

type UserLevelProgress struct {
	ID          uint      `json:"id" gorm:"primaryKey" db:"id"`
	UserID      uint      `json:"user_id" gorm:"not null;index" db:"user_id"`
	LevelID     uint      `json:"level_id" gorm:"not null;index" db:"level_id"`
	IsCompleted bool      `json:"is_completed" gorm:"default:false" db:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Level Level `json:"level,omitempty" gorm:"foreignKey:LevelID"`
}

func (UserLevelProgress) TableName() string {
	return "user_level_progress"
}

type UserReward struct {
	ID           uint      `json:"id" gorm:"primaryKey" db:"id"`
	UserID       uint      `json:"user_id" gorm:"not null;uniqueIndex" db:"user_id"`
	TotalRewards int       `json:"total_rewards" gorm:"default:0" db:"total_rewards"`
	LevelReached int       `json:"level_reached" gorm:"default:1" db:"level_reached"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (UserReward) TableName() string {
	return "user_rewards"
}

// GORM Hooks
func (l *Level) BeforeCreate(tx *gorm.DB) error {
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (l *Level) BeforeUpdate(tx *gorm.DB) error {
	l.UpdatedAt = time.Now().UTC()
	return nil
}

func (ulp *UserLevelProgress) BeforeCreate(tx *gorm.DB) error {
	if ulp.CreatedAt.IsZero() {
		ulp.CreatedAt = time.Now().UTC()
	}
	if ulp.UpdatedAt.IsZero() {
		ulp.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (ulp *UserLevelProgress) BeforeUpdate(tx *gorm.DB) error {
	ulp.UpdatedAt = time.Now().UTC()
	return nil
}

func (ur *UserReward) BeforeCreate(tx *gorm.DB) error {
	if ur.CreatedAt.IsZero() {
		ur.CreatedAt = time.Now().UTC()
	}
	if ur.UpdatedAt.IsZero() {
		ur.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (ur *UserReward) BeforeUpdate(tx *gorm.DB) error {
	ur.UpdatedAt = time.Now().UTC()
	return nil
}