package main

import (
	"time"
)

type Model struct {
	ID   uint   `gorm:"primary_key" json:"-"`
	UUID string `gorm:"type:uuid; default:uuid_generate_v4(); index;" json:"uuid"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (m *Model) zeroID() {
	m.ID = 0
	m.UUID = ""
}

func (m *Model) postRead() {
}

func (m *Model) postSearch() {
}

func (m *Model) applyDefaults() {

}
