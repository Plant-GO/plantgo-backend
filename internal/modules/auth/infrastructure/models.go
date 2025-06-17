package infrastructure

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey" db:"id"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null;size:255" db:"username"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null;size:255" db:"email"`
	PasswordHash string         `json:"-" gorm:"column:password_hash;size:255" db:"password_hash"`
	AndroidID    *string        `json:"android_id,omitempty" gorm:"column:android_id;size:255" db:"android_id"`
	GoogleID     *string        `json:"google_id,omitempty" gorm:"column:google_id;size:255" db:"google_id"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"` 
}

func (User) TableName() string {
	return "users"
}


func (u *User) BeforeCreate(tx *gorm.DB) error {
	// BeforeCreate hook

	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // BeforeUpdate hook 
    
	u.UpdatedAt = time.Now().UTC()
	return nil
}