package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"plantgo-backend/internal/modules/notification/infrastructure"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

type FirebaseService struct {
	client *messaging.Client
	repo   *infrastructure.NotificationRepository
}

func NewFirebaseService(repo *infrastructure.NotificationRepository) (*FirebaseService, error) {
	// Initialize Firebase Admin SDK
	credentialsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credentialsPath == "" {
		log.Println("Warning: FIREBASE_CREDENTIALS_PATH not set, Firebase messaging will be disabled")
		return &FirebaseService{client: nil, repo: repo}, nil
	}

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase messaging client: %v", err)
	}

	log.Println("Firebase messaging service initialized successfully")
	return &FirebaseService{client: client, repo: repo}, nil
}

func (f *FirebaseService) SendPushNotification(notification *infrastructure.Notification) error {
	if f.client == nil {
		log.Println("Firebase client not initialized, skipping push notification")
		return nil
	}

	// Get user's FCM token
	fcmToken, err := f.repo.GetUserFCMToken(notification.UserID)
	if err != nil {
		log.Printf("Failed to get FCM token for user %d: %v", notification.UserID, err)
		f.repo.UpdateNotificationStatus(notification.ID, infrastructure.Failed)
		return err
	}

	// Parse notification data
	var data map[string]string
	if notification.Data != "" {
		var notificationData map[string]interface{}
		if err := json.Unmarshal([]byte(notification.Data), &notificationData); err == nil {
			data = make(map[string]string)
			for k, v := range notificationData {
				data[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Add basic notification metadata
	if data == nil {
		data = make(map[string]string)
	}
	data["notification_id"] = strconv.FormatUint(uint64(notification.ID), 10)
	data["type"] = string(notification.Type)
	data["user_id"] = strconv.FormatUint(uint64(notification.UserID), 10)

	// Create FCM message
	message := &messaging.Message{
		Token: fcmToken,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Message,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Icon:        "ic_notification",
				Color:       "#4CAF50",
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				ChannelID:   "plantgo_notifications",
				Priority:    messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: notification.Title,
						Body:  notification.Message,
					},
					Sound: "default",
					Badge: f.getUnreadCountForUser(notification.UserID),
				},
			},
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: notification.Title,
				Body:  notification.Message,
				Icon:  "/icons/icon-192x192.png",
				Badge: "/icons/badge-72x72.png",
			},
		},
	}

	// Send the message
	response, err := f.client.Send(context.Background(), message)
	if err != nil {
		log.Printf("Failed to send FCM message: %v", err)
		f.repo.UpdateNotificationStatus(notification.ID, infrastructure.Failed)
		return err
	}

	log.Printf("Successfully sent FCM message: %s", response)
	f.repo.UpdateNotificationStatus(notification.ID, infrastructure.Sent)
	return nil
}

func (f *FirebaseService) SendBulkPushNotifications(notifications []infrastructure.Notification) error {
	if f.client == nil {
		log.Println("Firebase client not initialized, skipping bulk push notifications")
		return nil
	}

	if len(notifications) == 0 {
		return nil
	}

	// Group notifications by FCM token to optimize sending
	tokenGroups := make(map[string][]infrastructure.Notification)

	for _, notification := range notifications {
		fcmToken, err := f.repo.GetUserFCMToken(notification.UserID)
		if err != nil {
			log.Printf("Failed to get FCM token for user %d: %v", notification.UserID, err)
			f.repo.UpdateNotificationStatus(notification.ID, infrastructure.Failed)
			continue
		}

		tokenGroups[fcmToken] = append(tokenGroups[fcmToken], notification)
	}

	// Send notifications for each token
	for fcmToken, userNotifications := range tokenGroups {
		// Send the most recent notification for each user
		latestNotification := userNotifications[len(userNotifications)-1]

		err := f.sendSingleNotification(fcmToken, &latestNotification)
		if err != nil {
			log.Printf("Failed to send bulk notification: %v", err)
		}
	}

	return nil
}

func (f *FirebaseService) sendSingleNotification(fcmToken string, notification *infrastructure.Notification) error {
	// Parse notification data
	var data map[string]string
	if notification.Data != "" {
		var notificationData map[string]interface{}
		if err := json.Unmarshal([]byte(notification.Data), &notificationData); err == nil {
			data = make(map[string]string)
			for k, v := range notificationData {
				data[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Add basic notification metadata
	if data == nil {
		data = make(map[string]string)
	}
	data["notification_id"] = strconv.FormatUint(uint64(notification.ID), 10)
	data["type"] = string(notification.Type)
	data["user_id"] = strconv.FormatUint(uint64(notification.UserID), 10)

	// Create FCM message
	message := &messaging.Message{
		Token: fcmToken,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Message,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Icon:        "ic_notification",
				Color:       "#4CAF50",
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				ChannelID:   "plantgo_notifications",
				Priority:    messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: notification.Title,
						Body:  notification.Message,
					},
					Sound: "default",
					Badge: f.getUnreadCountForUser(notification.UserID),
				},
			},
		},
	}

	// Send the message
	response, err := f.client.Send(context.Background(), message)
	if err != nil {
		log.Printf("Failed to send FCM message: %v", err)
		f.repo.UpdateNotificationStatus(notification.ID, infrastructure.Failed)
		return err
	}

	log.Printf("Successfully sent FCM message: %s", response)
	f.repo.UpdateNotificationStatus(notification.ID, infrastructure.Sent)
	return nil
}

func (f *FirebaseService) getUnreadCountForUser(userID uint) *int {
	count, err := f.repo.GetUnreadNotificationCount(userID)
	if err != nil {
		log.Printf("Failed to get unread count for user %d: %v", userID, err)
		return nil
	}
	intCount := int(count)
	return &intCount
}

// ValidateToken validates an FCM token
func (f *FirebaseService) ValidateToken(token string) error {
	if f.client == nil {
		return fmt.Errorf("firebase client not initialized")
	}

	// Create a test message to validate the token
	message := &messaging.Message{
		Token: token,
		Data: map[string]string{
			"test": "validation",
		},
	}

	// Validate the message (this doesn't send it)
	_, err := f.client.Send(context.Background(), message)
	return err
}

// SendTopicNotification sends a notification to a topic
func (f *FirebaseService) SendTopicNotification(topic, title, body string, data map[string]string) error {
	if f.client == nil {
		return fmt.Errorf("firebase client not initialized")
	}

	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Icon:        "ic_notification",
				Color:       "#4CAF50",
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				ChannelID:   "plantgo_notifications",
				Priority:    messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound: "default",
				},
			},
		},
	}

	response, err := f.client.Send(context.Background(), message)
	if err != nil {
		return fmt.Errorf("failed to send topic notification: %v", err)
	}

	log.Printf("Successfully sent topic notification: %s", response)
	return nil
}

// SubscribeToTopic subscribes tokens to a topic
func (f *FirebaseService) SubscribeToTopic(tokens []string, topic string) error {
	if f.client == nil {
		return fmt.Errorf("firebase client not initialized")
	}

	response, err := f.client.SubscribeToTopic(context.Background(), tokens, topic)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %v", err)
	}

	log.Printf("Successfully subscribed %d tokens to topic %s", response.SuccessCount, topic)
	return nil
}

// UnsubscribeFromTopic unsubscribes tokens from a topic
func (f *FirebaseService) UnsubscribeFromTopic(tokens []string, topic string) error {
	if f.client == nil {
		return fmt.Errorf("firebase client not initialized")
	}

	response, err := f.client.UnsubscribeFromTopic(context.Background(), tokens, topic)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from topic: %v", err)
	}

	log.Printf("Successfully unsubscribed %d tokens from topic %s", response.SuccessCount, topic)
	return nil
}

// ProcessPendingNotifications processes all pending notifications and sends push notifications
func (f *FirebaseService) ProcessPendingNotifications() error {
	pendingNotifications, err := f.repo.GetPendingNotifications(100) // Process up to 100 at a time
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %v", err)
	}

	if len(pendingNotifications) == 0 {
		return nil
	}

	log.Printf("Processing %d pending notifications", len(pendingNotifications))

	for _, notification := range pendingNotifications {
		err := f.SendPushNotification(&notification)
		if err != nil {
			log.Printf("Failed to send push notification for notification %d: %v", notification.ID, err)
		}
	}

	return nil
}
