package infrastructure

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey" db:"id"`
	Username     string         `json:"username" gorm:"not null;size:255" db:"username"`
	Email        *string        `json:"email,omitempty" gorm:"uniqueIndex;size:255" db:"email"` // Make nullable
	PasswordHash *string        `json:"-" gorm:"column:password_hash;size:255" db:"password_hash"` // Make nullable for guests
	AndroidID    *string        `json:"android_id,omitempty" gorm:"uniqueIndex;column:android_id;size:255" db:"android_id"` // Add unique index
	GoogleID     *string        `json:"google_id,omitempty" gorm:"uniqueIndex;column:google_id;size:255" db:"google_id"` // Add unique index
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"` 
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now().UTC()
	return nil
}