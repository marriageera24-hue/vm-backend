package main

import "time"

type IPInfoLog struct {
	ID               uint      `gorm:"primary_key" json:"id"`
	OrganizationUUID string    `gorm:"type:uuid;" json:"organization_uuid"`
	CreatedAt        time.Time `json:"created_at"`

	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message"`

	OrganizationName string `gorm:"-" json:"organization_name"`
}
