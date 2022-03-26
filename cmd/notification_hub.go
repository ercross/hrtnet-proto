package main

import (
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/gorilla/websocket"
	"sync"
)

// NotificationHub implements a simple notification system.
type NotificationHub struct {

	// maps userId to connections
	connections map[string]*websocket.Conn

	// sync mechanism for connections
	connLock sync.RWMutex

	storage NotificationRepo
}

func NewNotificationHub(storage NotificationRepo) *NotificationHub {
	return &NotificationHub{
		connections: make(map[string]*websocket.Conn),
		connLock:    sync.RWMutex{},
		storage:     storage,
	}
}

// AddConnection adds a new websocket connection to the connection pool
func (hub *NotificationHub) AddConnection(userId string, conn *websocket.Conn) {
	hub.connLock.Lock()
	hub.connections[userId] = conn
	hub.connLock.Unlock()
}

// Dispatch dispatches notification if target user
// has an active websocket connection and save to NotificationRepo
// regardless
func (hub *NotificationHub) Dispatch(notification *model.Notification) {
	go func() {
		hub.connLock.RLock()
		conn, ok := hub.connections[notification.UserID]
		if ok {
			conn.WriteJSON(notification)
		}
		hub.connLock.RUnlock()
		if err := hub.storage.SaveNotification(notification); err != nil {
			logger.Logger.LogError("error saving notification message", "dispatch notification", err)
		}
	}()
}

func (hub *NotificationHub) DispatchAllUnread(forUserId string) {

	go func() {

		// fetch all unread notifications from storage
		notifications, err := hub.storage.FetchAllUnreadNotifications(forUserId)
		if err != nil {
			logger.Logger.LogError("unable to fetch user's all unread notifications", "dispatch all unread", err)
			return
		}

		// send
		hub.connLock.RLock()
		conn, ok := hub.connections[forUserId]

		if ok {
			conn.WriteJSON(&notifications)
		}
		hub.connLock.RUnlock()
	}()
}

// RemoveConnection removes inactive connections from connections.
func (hub *NotificationHub) RemoveConnection(userId string) {
	hub.connLock.Lock()
	for id, _ := range hub.connections {
		if id == userId {
			delete(hub.connections, userId)
		}
	}
	hub.connLock.Unlock()
}
