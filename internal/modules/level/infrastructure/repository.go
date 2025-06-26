package infrastructure

import (
	"errors"
	"fmt"
	"time"
	"gorm.io/gorm"
)

type PlantRepository struct {
	db *gorm.DB
}

func NewPlantRepository(db *gorm.DB) *PlantRepository {
	return &PlantRepository{db: db}
}

// Level CRUD operations
func (r *PlantRepository) CreateLevel(level *Level) error {
	return r.db.Create(level).Error
}

func (r *PlantRepository) GetLevelByID(id uint) (*Level, error) {
	var level Level
	err := r.db.First(&level, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("level with ID %d not found", id)
		}
		return nil, err
	}
	return &level, nil
}

func (r *PlantRepository) GetAllLevels() ([]Level, error) {
	var levels []Level
	err := r.db.Order("id ASC").Find(&levels).Error
	return levels, err
}

func (r *PlantRepository) UpdateLevel(level *Level) error {
	return r.db.Save(level).Error
}

func (r *PlantRepository) DeleteLevel(id uint) error {
	return r.db.Delete(&Level{}, id).Error
}

func (r *PlantRepository) GetLevelsCount() (int64, error) {
	var count int64
	err := r.db.Model(&Level{}).Count(&count).Error
	return count, err
}

// UserLevelProgress CRUD operations  
func (r *PlantRepository) GetUserProgress(userID uint) ([]UserLevelProgress, error) {
	var progressList []UserLevelProgress
	err := r.db.Where("user_id = ?", userID).
		Preload("Level").
		Order("level_id ASC").
		Find(&progressList).Error
	return progressList, err
}

func (r *PlantRepository) GetCompletedLevels(userID uint) ([]UserLevelProgress, error) {
	var progressList []UserLevelProgress
	err := r.db.Where("user_id = ? AND is_completed = ?", userID, true).
		Preload("Level").
		Order("level_id ASC").
		Find(&progressList).Error
	return progressList, err
}

func (r *PlantRepository) IsLevelCompleted(userID, levelID uint) bool {
	var count int64
	r.db.Model(&UserLevelProgress{}).
		Where("user_id = ? AND level_id = ? AND is_completed = ?", userID, levelID, true).
		Count(&count)
	return count > 0
}

func (r *PlantRepository) CompleteLevel(userID, levelID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Check if already completed
		var existing UserLevelProgress
		err := tx.Where("user_id = ? AND level_id = ?", userID, levelID).First(&existing).Error
		
		now := time.Now().UTC()
		
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new progress record
			progress := UserLevelProgress{
				UserID:      userID,
				LevelID:     levelID,
				IsCompleted: true,
				CompletedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			err = tx.Create(&progress).Error
		} else if err == nil && !existing.IsCompleted {
			// Update existing record
			existing.IsCompleted = true
			existing.CompletedAt = &now
			existing.UpdatedAt = now
			err = tx.Save(&existing).Error
		}
		
		if err != nil {
			return err
		}
		
		// Get level reward
		var level Level
		err = tx.First(&level, levelID).Error
		if err != nil {
			return err
		}
		
		// Update user rewards
		err = r.addRewardToUser(tx, userID, level.Reward, int(levelID))
		if err != nil {
			return err
		}
		
		return nil
	})
}

// UserReward operations
func (r *PlantRepository) GetOrCreateUserReward(userID uint) (*UserReward, error) {
	var reward UserReward
	err := r.db.Where("user_id = ?", userID).First(&reward).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new user reward record
		reward = UserReward{
			UserID:       userID,
			TotalRewards: 0,
			LevelReached: 1,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		err = r.db.Create(&reward).Error
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &reward, nil
}

func (r *PlantRepository) addRewardToUser(tx *gorm.DB, userID uint, rewardPoints int, levelID int) error {
	var userReward UserReward
	err := tx.Where("user_id = ?", userID).First(&userReward).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new record
		userReward = UserReward{
			UserID:       userID,
			TotalRewards: rewardPoints,
			LevelReached: levelID,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		return tx.Create(&userReward).Error
	} else if err != nil {
		return err
	}
	
	// Update existing record
	userReward.TotalRewards += rewardPoints
	if levelID > userReward.LevelReached {
		userReward.LevelReached = levelID
	}
	userReward.UpdatedAt = time.Now().UTC()
	
	return tx.Save(&userReward).Error
}

func (r *PlantRepository) GetUserReward(userID uint) (*UserReward, error) {
	var reward UserReward
	err := r.db.Where("user_id = ?", userID).First(&reward).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user reward not found for user %d", userID)
		}
		return nil, err
	}
	return &reward, nil
}

func (r *PlantRepository) UpdateUserReward(reward *UserReward) error {
	return r.db.Save(reward).Error
}

// Game loading operations
func (r *PlantRepository) GetGameData(userID uint) (map[string]interface{}, error) {
	// Get user rewards
	userReward, err := r.GetOrCreateUserReward(userID)
	if err != nil {
		return nil, err
	}
	
	// Get all levels
	levels, err := r.GetAllLevels()
	if err != nil {
		return nil, err
	}
	
	// Get completed levels
	completedLevels, err := r.GetCompletedLevels(userID)
	if err != nil {
		return nil, err
	}
	
	// Create a map of completed level IDs for quick lookup
	completedMap := make(map[uint]bool)
	for _, progress := range completedLevels {
		completedMap[progress.LevelID] = true
	}
	
	// Prepare level data with completion status
	levelData := make([]map[string]interface{}, len(levels))
	for i, level := range levels {
		levelData[i] = map[string]interface{}{
			"id":           level.ID,
			"reward":       level.Reward,
			"is_completed": completedMap[level.ID],
			"is_unlocked":  int(level.ID) <= userReward.LevelReached,
		}
	}
	
	return map[string]interface{}{
		"user_reward":      userReward,
		"levels":           levelData,
		"completed_levels": len(completedLevels),
		"total_levels":     len(levels),
	}, nil
}

func (r *PlantRepository) GetLevelDetails(levelID uint) (map[string]interface{}, error) {
	level, err := r.GetLevelByID(levelID)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"id":         level.ID,
		"riddle":     level.Riddle,
		"plant_name": level.PlantName,
		"reward":     level.Reward,
	}, nil
}