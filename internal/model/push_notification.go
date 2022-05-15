package model

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"github.com/Hrtnet/social-activities/internal/logger"
	"time"
)

var firebaseApp *firebase.App

// InitializeFirebaseAdminSDK must be invoked at app start before invoking
// any other PushNotification methods.
// https://firebase.google.com/docs/admin/setup#go
// https://firebase.google.com/docs/cloud-messaging/auth-server#provide-credentials-manually
func InitializeFirebaseAdminSDK() {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		logger.Logger.LogError("failed to initialize firebase app",
			"initializing firebase admin sdk", err)
	}
	firebaseApp = app
}

type Topic string

var Heartnet Topic = "heartnet"

type PushNotification struct {
	messaging.Notification

	// Data holds any extra data.
	// One important data field is url,
	// which can be used to pass an app deep link or web link.
	Data map[string]string
}

// SendToUser sends message to user attached to token
// Notifications sent to single user are usually of great importance.
// https://firebase.google.com/docs/cloud-messaging/send-message#send-messages-to-specific-devices
func (p PushNotification) SendToUser(token string) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client, err := firebaseApp.Messaging(ctx)
	if err != nil {
		logger.Logger.LogError(
			"failed to obtain firebase cloud messaging client",
			"send push notification to users", err)
	}

	// messaging.AndroidConfig.TTL not set here so we can take advantage of the default.
	message := &messaging.Message{
		Data: p.Data,
		Notification: &messaging.Notification{
			Title:    p.Title,
			Body:     p.Body,
			ImageURL: p.ImageURL,
		},
		Token: token,
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		logger.Logger.LogError(
			"failed to send push notification", "send to multiple users", err)
	}
}

// SendToMultipleUsers sends message to multiple user that are
// subscribed to topic.
// https://firebase.google.com/docs/cloud-messaging/send-message#send-messages-to-topics
// To send message to multiple users through their tokens, check the resource at
// https://firebase.google.com/docs/cloud-messaging/send-message#send-messages-to-multiple-devices
func (p PushNotification) SendToMultipleUsers(topic Topic) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client, err := firebaseApp.Messaging(ctx)
	if err != nil {
		logger.Logger.LogError(
			"failed to obtain firebase cloud messaging client",
			"send push notification messages to users", err)
		return
	}

	ttl := time.Hour * 24
	message := &messaging.Message{
		Data: p.Data,
		Notification: &messaging.Notification{
			Title:    p.Title,
			Body:     p.Body,
			ImageURL: p.ImageURL,
		},
		Android: &messaging.AndroidConfig{
			TTL: &ttl,
		},
		Topic: fmt.Sprintf("%s", topic),
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		logger.Logger.LogError(
			"failed to send push notification to multiple users", "send to multiple users", err)
		return
	}
}
