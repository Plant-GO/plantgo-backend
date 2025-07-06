# PlantGo Notification System Documentation

## Overview
The PlantGo notification system provides a comprehensive solution for managing user notifications with Firebase push notification support. The system is designed to be scalable, maintainable, and compatible with Flutter frontend applications.

## Architecture

### Components
1. **Models** (`infrastructure/models.go`) - Database entities and types
2. **Repository** (`infrastructure/repository.go`) - Data access layer
3. **Service** (`service.go`) - Business logic layer
4. **Handler** (`handler.go`) - HTTP handlers for API endpoints
5. **Firebase Service** (`firebase_service.go`) - Firebase FCM integration

### Database Schema

#### Notifications Table
```sql
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    read_at TIMESTAMP
);
```

#### User Notification Preferences Table
```sql
CREATE TABLE user_notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    friend_requests BOOLEAN DEFAULT true,
    game_rewards BOOLEAN DEFAULT true,
    weekly_challenges BOOLEAN DEFAULT true,
    daily_login_rewards BOOLEAN DEFAULT true,
    level_completes BOOLEAN DEFAULT true,
    achievement_unlocks BOOLEAN DEFAULT true,
    system_announcements BOOLEAN DEFAULT true,
    plant_identified BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### User FCM Tokens Table
```sql
CREATE TABLE user_fcm_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Notification Types

The system supports the following notification types:

1. **friend_request** - Friend requests from other users
2. **game_reward** - General game rewards
3. **weekly_challenge** - Weekly challenge completions
4. **daily_login_reward** - Daily login streaks
5. **level_complete** - Level completion notifications
6. **achievement_unlocked** - Achievement unlocks
7. **system_announcement** - System-wide announcements
8. **plant_identified** - Plant identification results

## API Endpoints

### Notification Management

#### Get User Notifications
```http
GET /api/v1/notifications/{userId}?limit=20&offset=0
```

#### Get Unread Notifications
```http
GET /api/v1/notifications/{userId}/unread
```

#### Get Unread Count
```http
GET /api/v1/notifications/{userId}/count
```

#### Mark Notification as Read
```http
PUT /api/v1/notifications/{id}/read
```

#### Mark All Notifications as Read
```http
PUT /api/v1/notifications/{userId}/read-all
```

#### Delete Notification
```http
DELETE /api/v1/notifications/{id}
```

### FCM Token Management

#### Update FCM Token
```http
POST /api/v1/notifications/fcm-token
Content-Type: application/json

{
    "user_id": 123,
    "token": "firebase_fcm_token_here"
}
```

### User Preferences

#### Get User Notification Preferences
```http
GET /api/v1/notifications/{userId}/preferences
```

#### Update User Notification Preferences
```http
PUT /api/v1/notifications/{userId}/preferences
Content-Type: application/json

{
    "friend_requests": true,
    "game_rewards": true,
    "weekly_challenges": false,
    "daily_login_rewards": true,
    "level_completes": true,
    "achievement_unlocks": true,
    "system_announcements": true,
    "plant_identified": true
}
```

## Service Layer Methods

### Notification Generation

```go
// Generate level completion notification
func (s *NotificationService) GenerateLevelCompleteNotification(userID uint, levelNumber int, reward int) error

// Generate daily login reward notification
func (s *NotificationService) GenerateDailyLoginReward(userID uint, reward int, streak int) error

// Generate weekly challenge completion notification
func (s *NotificationService) GenerateWeeklyChallengeComplete(userID uint, challengeName string, reward int) error

// Generate friend request notification
func (s *NotificationService) GenerateFriendRequestNotification(userID uint, fromUserID uint, fromUsername string) error

// Generate achievement unlock notification
func (s *NotificationService) GenerateAchievementUnlocked(userID uint, achievementName string, reward int) error

// Generate system announcement
func (s *NotificationService) GenerateSystemAnnouncement(userID uint, title, message string) error

// Generate plant identification notification
func (s *NotificationService) GeneratePlantIdentifiedNotification(userID uint, plantName string, confidence float64) error

// Generate general game reward notification
func (s *NotificationService) GenerateGameRewardNotification(userID uint, rewardType string, reward int, description string) error
```

### Notification Retrieval

```go
// Get paginated notifications for a user
func (s *NotificationService) GetUserNotifications(userID uint, limit, offset int) ([]infrastructure.Notification, error)

// Get unread notifications for a user
func (s *NotificationService) GetUnreadNotifications(userID uint) ([]infrastructure.Notification, error)

// Get count of unread notifications
func (s *NotificationService) GetUnreadNotificationCount(userID uint) (int64, error)
```

## Firebase Configuration

### Environment Variables

Add these variables to your `.env` file:

```bash
# Firebase Configuration for Push Notifications
FIREBASE_PROJECT_ID=your-firebase-project-id
FIREBASE_CREDENTIALS_PATH=./firebase-service-account-key.json
```

### Firebase Service Account Key

1. Go to Firebase Console
2. Select your project
3. Go to Project Settings > Service accounts
4. Click "Generate new private key"
5. Save the JSON file as `firebase-service-account-key.json` in your project root

## Integration Examples

### Level Completion Integration

```go
// In your level completion handler
func (h *LevelHandler) CompleteLevel(c *gin.Context) {
    // ... level completion logic ...
    
    // Generate notification
    if h.notificationService != nil {
        err := h.notificationService.GenerateLevelCompleteNotification(
            userID,
            level.LevelNumber,
            level.Reward,
        )
        if err != nil {
            log.Printf("Failed to generate level completion notification: %v", err)
        }
    }
    
    // ... rest of handler ...
}
```

### Plant Identification Integration

```go
// In your plant scanning handler
func (h *PlantHandler) ScanImage(c *gin.Context) {
    // ... scanning logic ...
    
    // Generate notification for successful identification
    if confidence > 0.7 && h.notificationService != nil {
        err := h.notificationService.GeneratePlantIdentifiedNotification(
            userID,
            plantName,
            confidence,
        )
        if err != nil {
            log.Printf("Failed to generate plant identification notification: %v", err)
        }
    }
    
    // ... rest of handler ...
}
```

## Error Handling

The system includes comprehensive error handling:

1. **Database Errors** - Proper GORM error handling with meaningful messages
2. **Firebase Errors** - Graceful degradation when Firebase is unavailable
3. **Validation Errors** - Input validation with clear error responses
4. **Permission Errors** - User preference checking before notification creation

## Testing

### Unit Tests

```go
func TestNotificationService_GenerateLevelCompleteNotification(t *testing.T) {
    // Test notification generation
    service := NewNotificationService(mockRepo, mockFirebase)
    
    err := service.GenerateLevelCompleteNotification(1, 5, 100)
    assert.NoError(t, err)
    
    // Verify notification was created
    notifications, _ := service.GetUserNotifications(1, 10, 0)
    assert.Len(t, notifications, 1)
    assert.Equal(t, "Level Complete! 🎉", notifications[0].Title)
}
```

### Integration Tests

```go
func TestNotificationAPI_GetUserNotifications(t *testing.T) {
    // Test API endpoint
    req := httptest.NewRequest("GET", "/api/v1/notifications/1", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    var response Response
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.True(t, response.Success)
}
```

## Performance Considerations

1. **Database Indexing** - Indexes on user_id, type, and created_at columns
2. **Pagination** - All list endpoints support pagination
3. **Async Processing** - Push notifications are sent asynchronously
4. **Bulk Operations** - Support for bulk notification creation
5. **Connection Pooling** - Efficient database connection management

## Security

1. **User Authorization** - Ensure users can only access their own notifications
2. **Input Validation** - Validate all input parameters
3. **Rate Limiting** - Implement rate limiting for notification endpoints
4. **Firebase Security** - Secure Firebase service account key storage

## Monitoring and Logging

The system includes comprehensive logging:

```go
log.Printf("Successfully sent FCM message: %s", response)
log.Printf("Failed to send FCM message: %v", err)
log.Printf("Firebase client not initialized, skipping push notification")
```

## Future Enhancements

1. **Notification Templates** - Dynamic notification templates
2. **Scheduling** - Scheduled notifications
3. **Analytics** - Notification engagement tracking
4. **Webhook Support** - Third-party integrations
5. **Rich Media** - Image and action button support
6. **Notification Grouping** - Smart notification bundling
7. **A/B Testing** - Notification content testing
8. **Localization** - Multi-language support

## Troubleshooting

### Common Issues

1. **Firebase Not Initialized**: Check FIREBASE_CREDENTIALS_PATH environment variable
2. **No Push Notifications**: Verify FCM token is properly stored
3. **Permission Denied**: Check user notification preferences
4. **Database Errors**: Verify auto-migration completed successfully

### Debug Commands

```bash
# Check Firebase credentials
echo $FIREBASE_CREDENTIALS_PATH

# Test notification creation
curl -X POST http://localhost:8080/api/v1/notifications/fcm-token \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "token": "test_token"}'

# Check notification preferences
curl http://localhost:8080/api/v1/notifications/1/preferences
```

This notification system provides a robust foundation for user engagement in the PlantGo application, with full Firebase integration for real-time push notifications to your Flutter frontend.
      "type": "level_complete",
      "title": "Level Complete! 🎉",
      "message": "Congratulations! You completed level 5 and earned 100 coins!",
      "data": "{\"level_number\":5,\"reward\":100}",
      "status": "sent",
      "is_read": false,
      "created_at": "2025-01-06T10:30:00Z",
      "updated_at": "2025-01-06T10:30:00Z",
      "read_at": null
    }
  ]
}
```

### 2. Get Unread Notifications
```http
GET /notifications/{userId}/unread
```

### 3. Get Unread Count
```http
GET /notifications/{userId}/unread/count
```

**Response:**
```json
{
  "success": true,
  "message": "Unread count retrieved successfully",
  "data": {
    "count": 5
  }
}
```

### 4. Mark Notification as Read
```http
PUT /notifications/{notificationId}/read
```

### 5. Mark All Notifications as Read
```http
PUT /notifications/{userId}/read-all
```

### 6. Delete Notification
```http
DELETE /notifications/{notificationId}
```

### 7. Update FCM Token
```http
POST /notifications/fcm-token
```

**Request Body:**
```json
{
  "user_id": 123,
  "token": "fcm_token_here"
}
```

### 8. Get User Preferences
```http
GET /notifications/{userId}/preferences
```

**Response:**
```json
{
  "success": true,
  "message": "Preferences retrieved successfully",
  "data": {
    "id": 1,
    "user_id": 123,
    "friend_requests": true,
    "game_rewards": true,
    "weekly_challenges": true,
    "daily_login_rewards": true,
    "level_completes": true,
    "achievement_unlocks": true,
    "system_announcements": true,
    "plant_identified": true,
    "created_at": "2025-01-06T10:30:00Z",
    "updated_at": "2025-01-06T10:30:00Z"
  }
}
```

### 9. Update User Preferences
```http
PUT /notifications/{userId}/preferences
```

**Request Body:**
```json
{
  "friend_requests": true,
  "game_rewards": false,
  "weekly_challenges": true,
  "daily_login_rewards": true,
  "level_completes": true,
  "achievement_unlocks": true,
  "system_announcements": false,
  "plant_identified": true
}
```

## Database Schema

### Notifications Table
```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    INDEX idx_is_read (is_read)
);
```

### User Notification Preferences Table
```sql
CREATE TABLE user_notification_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    friend_requests BOOLEAN DEFAULT true,
    game_rewards BOOLEAN DEFAULT true,
    weekly_challenges BOOLEAN DEFAULT true,
    daily_login_rewards BOOLEAN DEFAULT true,
    level_completes BOOLEAN DEFAULT true,
    achievement_unlocks BOOLEAN DEFAULT true,
    system_announcements BOOLEAN DEFAULT true,
    plant_identified BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### User FCM Tokens Table
```sql
CREATE TABLE user_fcm_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
);
```

## Integration with Game Events

The notification system automatically generates notifications for various game events:

### Level Completion
```go
// Triggered when a user completes a level
notificationService.GenerateLevelCompleteNotification(userID, levelNumber, reward)
```

### Daily Login
```go
// Triggered on daily login
notificationService.GenerateDailyLoginReward(userID, reward, streak)
```

### Plant Identification
```go
// Triggered when ML model identifies a plant
notificationService.GeneratePlantIdentifiedNotification(userID, plantName, confidence)
```

### Achievement Unlocked
```go
// Triggered when user unlocks an achievement
notificationService.GenerateAchievementUnlocked(userID, achievementName, reward)
```

### Friend Requests
```go
// Triggered when user receives a friend request
notificationService.GenerateFriendRequestNotification(userID, fromUserID, fromUsername)
```

## Error Handling

All API endpoints return standardized error responses:

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message"
}
```

Common HTTP status codes:
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `404` - Not Found (resource doesn't exist)
- `500` - Internal Server Error

## Performance Considerations

1. **Pagination**: Use `limit` and `offset` parameters for large notification lists
2. **Indexing**: Database indexes on `user_id`, `created_at`, and `is_read` for fast queries
3. **Bulk Operations**: Use bulk endpoints for multiple operations
4. **Cleanup**: Implement periodic cleanup of old read notifications

## Security

1. **User Validation**: All endpoints validate user ownership of resources
2. **Input Sanitization**: All inputs are validated and sanitized
3. **Rate Limiting**: Consider implementing rate limiting for notification creation
4. **Token Security**: FCM tokens are stored securely and can be deactivated

## Monitoring and Analytics

The system provides notification statistics through the stats endpoint:

```http
GET /notifications/{userId}/stats
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total": 150,
    "unread": 5,
    "by_type": [
      {"type": "level_complete", "count": 45},
      {"type": "daily_login_reward", "count": 30},
      {"type": "achievement_unlocked", "count": 15}
    ]
  }
}
```

## Environment Variables

Add these to your `.env` file:

```env
# Firebase Configuration for Push Notifications
FIREBASE_PROJECT_ID=your-firebase-project-id
FIREBASE_CREDENTIALS_PATH=./firebase-service-account-key.json
```

## Testing

You can test the notification system using curl:

```bash
# Get user notifications
curl -X GET "http://localhost:8080/api/v1/notifications/1?limit=10&offset=0"

# Update FCM token
curl -X POST "http://localhost:8080/api/v1/notifications/fcm-token" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "token": "your-fcm-token"}'

# Mark notification as read
curl -X PUT "http://localhost:8080/api/v1/notifications/1/read"

# Update preferences
curl -X PUT "http://localhost:8080/api/v1/notifications/1/preferences" \
  -H "Content-Type: application/json" \
  -d '{"friend_requests": true, "game_rewards": false}'
```

## Migration Guide

If upgrading from a system without notifications:

1. The system automatically creates default preferences for existing users
2. All existing users start with all notification types enabled
3. Database migrations run automatically on startup
4. No data loss occurs during the upgrade process

## Troubleshooting

### Common Issues

1. **Notifications not being created**: Check if user preferences allow the notification type
2. **FCM token not working**: Ensure Firebase credentials are properly configured
3. **Database errors**: Verify database migrations have run successfully
4. **Performance issues**: Check database indexes and consider pagination

### Debug Mode

Enable debug logging by setting the log level to debug in your application configuration.

### Health Check

The system includes health checks accessible via:
```http
GET /health
```

This will show database connectivity and migration status.
