package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/maddevsio/fcm"
)

// ref https://godoc.org/github.com/maddevsio/fcm
func sendToClientDevices(n *Notification) bool {
	// data := map[string]string{
	// 	"body":  "Hello World1",
	// 	"title": "Happy Day",
	// }
	serverKey := os.Getenv("FIREBASE_SERVER_KEY")
	// serverKey := "AIzaSyBp0-T6wZ1uKckMJMX-u3ISozMZYX3FT4g"
	c := fcm.NewFCM(serverKey)
	response, err := c.Send(fcm.Message{
		// RegistrationIDs: []string{n.FirebaseToken},
		To:          n.FirebaseToken,
		CollapseKey: "type_a",
		// Data:        data,
		// Priority: fcm.PriorityHigh,
		Notification: fcm.Notification{
			Title: "Matrimony",
			Body:  n.Message,
			Sound: "default",
			Badge: "3",
		},
	})

	if err != nil {
		log.Fatal(err.Error())
		return false
	}
	d, e := json.Marshal(response)

	if e != nil {
		log.Fatal(err.Error())
		return false
	}

	fmt.Println(string(d))
	return true
}
