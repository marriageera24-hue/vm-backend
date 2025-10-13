package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
)

type SMS struct {
	Model

	ToUserUUID        string    `json:"to_user_uuid"`
	ToMobile          string    `json:"mobile"`
	FromUserUUID      string    `json:"-"`
	Type              string    `json:"type"` // type can be OTP, invite, welcome
	Message           string    `json:"message"`
	Status            string    `json:"status"` // waiting, in-process, delivered, expired
	ValidTill         time.Time `json:"valid_till"`
	ProcessedByMobile string    `json:"processed_by_mobile"`
}

func generateOTP(u User, phone string) error {

	var s SMS

	db.Where("to_user_uuid = ? AND to_mobile = ? AND valid_till > NOW()", u.UUID, phone).Order("id DESC").First(&s)

	if s.ID == 0 {

		m, err := generateRandomString(8)

		if err != nil {
			return err
		}

		s.ToUserUUID = u.UUID
		s.ToMobile = u.Phone
		s.Type = "otp"
		s.Message = m
		s.Status = "waiting"
		s.ValidTill = time.Now().Add(time.Minute * 30)

		db.Save(&s)
	}

	return nil
}

func getOTP(ctx echo.Context) error {

	var (
		ss []SMS
	)

	err := db.Where("status = 'waiting' AND valid_till > NOW()").Order("id ASC").Limit(20).Find(&ss).Error
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, ss)
}

func updateOTPStatus(ctx echo.Context) error {

	var s SMS

	// Populate object from JSON
	err := ctx.Bind(&s)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	return nil
}
