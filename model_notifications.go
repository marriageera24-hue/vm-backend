package main

import (
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo"
)

type Notification struct {
	Model

	SenderID         string    `gorm:"index" json:"sender_uuid"`
	ReceiverID       string    `gorm:"index" json:"receiver_uuid"`
	Message          string    `json:"message"`
	ReferenceID      uint      `json:"-"`
	Status           bool      `json:"status"`
	NotificationDate time.Time `json:"notification_date"`
	FirebaseToken    string    `json:"-"`
	FirebaseStatus   bool      `json:"-"`
	Sender           User      `gorm:"-" json:"-"`
	Receiver         User      `gorm:"-" json:"-"`
}

func (n *Notification) createAndSend() {
	// generate the notification
	n.generate()

	if n.FirebaseToken != "" {
		n.FirebaseStatus = sendToClientDevices(n)
	}

	err := db.Save(&n).Error

	if err != nil {
		log.Println("Error while saving notification to DB: ", err.Error())
	}
}

func (n *Notification) generate() {
	n.Status = UNREAD

	// fetch users from db

	if len(n.Sender.UUID) == 0 {
		db.Model(&User{}).Where("uuid=?", n.SenderID).Find(&n.Sender)
	}

	if len(n.Receiver.UUID) == 0 {
		db.Model(&User{}).Where("uuid=?", n.ReceiverID).Find(&n.Receiver)
	}

	switch n.ReferenceID {
	case 1:
	case 2:
		n.Message = fmt.Sprintf("%s has shown interest in your profile.", n.Sender.getName())
	case 3:
		n.Message = fmt.Sprintf("%s has visited your profile.", n.Sender.getName())
	}

	if token, exist := n.Receiver.OtherInfo["fcmToken"]; exist && len(token.(string)) > 0 {
		n.FirebaseToken = token.(string)
	}
}

func myNotifications(ctx echo.Context) error {

	return nil
}
