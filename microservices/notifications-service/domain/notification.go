package domain

import (
	"github.com/gocql/gocql"
	"time"
)

type Notification struct {
	UserId         string     `json:"userId"`
	CreatedAt      time.Time  `json:"createdAt"`
	NotificationId gocql.UUID `json:"notificationId"`
	Message        string     `json:"message"`
	Status         string     `json:"status"`
}

type Notifications []*Notification
