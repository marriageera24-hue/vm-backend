package main

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func openDatabaseConnection() error {
	var err error

	db, err = gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)

	if os.Getenv("VM_ENVIRONMENT") != "production" {
		db.LogMode(true)
	}

	return err
}
