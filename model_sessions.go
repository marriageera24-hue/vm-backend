package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type Session struct {
	ID uint `gorm:"primary_key" json:"-"`

	UserID    uint
	Token     string
	IPAddress string
	ExpiresAt time.Time

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func verifySession(ctx echo.Context) (User, error) {
	var (
		// s Session
		u   User
		err error
		s   Session
	)

	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	id := uint(claims["id"].(float64))

	rf := db.Model(&Session{}).Where("user_id = ?", id).Find(&s).RecordNotFound()

	if rf {
		return u, fmt.Errorf("User session not found")
	}

	err = db.Model(&User{}).Where("id = ?", s.UserID).First(&u).Error

	return u, err
}
