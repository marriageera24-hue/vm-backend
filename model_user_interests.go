package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

type UserInterest struct {
	ID           uint   `gorm:"primary_key" json:"-"`
	FromUserUUID string `json:"from_user_uuid"`
	ToUserUUID   string `json:"to_user_uuid"`
	Type         string `json:"type"`
	UserInfo     User   `gorm:"-"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (u *UserInterest) sanitize(ctx echo.Context) {
	u.Type = sanitizeText(u.Type, 64)
}

func (u *UserInterest) validate(ctx echo.Context, skipRequiredCheck bool) error {

	if u.FromUserUUID == "" && !skipRequiredCheck {
		return errors.New("From UUID required")
	}

	if u.ToUserUUID == "" && !skipRequiredCheck {
		return errors.New("To UUID required")
	}

	if u.Type == "" && !skipRequiredCheck {
		return errors.New("Type is required")
	}

	return nil
}

func interests(ctx echo.Context) error {
	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	var uu []UserInterest

	db.Where("from_user_uuid = ?", u.UUID).Find(&uu)

	if len(uu) > 0 {
		for i, u := range uu {
			db.Model(&User{}).Where("uuid = ?", u.ToUserUUID).First(&u.UserInfo)
			db.Model(&Media{}).Where("user_uuid = ?", u.UserInfo.UUID).Find(&u.UserInfo.UserMedia)
			uu[i] = u
		}
	}

	return ctx.JSON(http.StatusOK, uu)
}

func interested(ctx echo.Context) error {
	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	var uu []UserInterest

	db.Where("to_user_uuid = ?", u.UUID).Find(&uu)

	if len(uu) > 0 {
		for i, u := range uu {
			db.Model(&User{}).Where("uuid = ?", u.FromUserUUID).First(&u.UserInfo)
			db.Model(&Media{}).Where("user_uuid = ?", u.UserInfo.UUID).Find(&u.UserInfo.UserMedia)
			uu[i] = u
		}
	}

	return ctx.JSON(http.StatusOK, uu)
}

func addInterest(ctx echo.Context) error {
	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	var (
		req UserInterest
	)

	// Populate object from JSON
	err = ctx.Bind(&req)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	req.FromUserUUID = u.UUID

	req.sanitize(ctx)
	err = req.validate(ctx, true)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	var ui UserInterest

	log.Println(req)
	db.Where("from_user_uuid = ? AND to_user_uuid = ?", req.FromUserUUID, req.ToUserUUID).First(&ui)
	ui.Type = req.Type
	ui.FromUserUUID = req.FromUserUUID
	ui.ToUserUUID = req.ToUserUUID
	err = db.Save(&ui).Error

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	n := Notification{
		SenderID:   ui.FromUserUUID,
		ReceiverID: ui.ToUserUUID,
	}

	if strings.ToLower(ui.Type) == "visited" {
		n.ReferenceID = PersonVisited
	} else if strings.ToLower(ui.Type) == "interested" {
		n.ReferenceID = PersonInterested
	}

	//n.createAndSend()
	return ctx.NoContent(http.StatusCreated)
}
