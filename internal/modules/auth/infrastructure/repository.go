package infrastructure

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByGoogleID(googleID string) (*User, error) {
	var user User
	err := r.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with Google ID %s not found", googleID)
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByAndroidID(androidID string) (*User, error) {
	var user User
	err := r.db.Where("android_id = ?", androidID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with Android ID %s not found", androidID)
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) DeleteUser(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

func (r *UserRepository) GetAllUsers(limit, offset int) ([]User, error) {
	var users []User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *UserRepository) UserExists(email, googleID string) (*User, bool) {
	var user User
	var err error
	
	if googleID != "" {
		err = r.db.Where("google_id = ?", googleID).First(&user).Error
	} else if email != "" {
		err = r.db.Where("email = ?", email).First(&user).Error
	}
	
	if err != nil {
		return nil, false
	}
	return &user, true
}

func (r *UserRepository) CreateOrUpdateUser(user *User) (*User, error) {
	// Check if user exists by Google ID or email
	var existingUser User
	var err error
	
	if user.GoogleID != nil && *user.GoogleID != "" {
		err = r.db.Where("google_id = ?", *user.GoogleID).First(&existingUser).Error
	} else if user.Email != "" {
		err = r.db.Where("email = ?", user.Email).First(&existingUser).Error
	}
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// User doesn't exist, create new
		if err := r.db.Create(user).Error; err != nil {
			return nil, err
		}
		return user, nil
	}
	
	// User exists, update
	existingUser.Username = user.Username
	existingUser.Email = user.Email
	if user.AndroidID != nil {
		existingUser.AndroidID = user.AndroidID
	}
	if user.GoogleID != nil {
		existingUser.GoogleID = user.GoogleID
	}
	
	if err := r.db.Save(&existingUser).Error; err != nil {
		return nil, err
	}
	
	return &existingUser, nil
}