package main

import (
	"errors"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

func apiCreate(ctx echo.Context, tx *gorm.DB, item apiObject, skipRequiredCheck bool) (int, error) {
	var err error

	item.zeroID()

	// Sanitize
	item.sanitize(ctx)

	// Check is valid
	err = item.validate(ctx, true)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Check if object exists
	exists := item.exists()
	if exists {
		return http.StatusConflict, errors.New(item.getExistsMessage(ctx))
	}

	tx.NewRecord(item)
	err = tx.Set("gorm:association_autoupdate", false).Create(item).Error
	if err != nil {
		return http.StatusInternalServerError, errors.New("Unable to create item")
	}

	return http.StatusCreated, nil
}

func apiUpdate(ctx echo.Context, tx *gorm.DB, item apiObject, skipRequiredCheck bool) (int, error) {
	var err error

	// Sanitize
	item.sanitize(ctx)

	// Check is valid
	err = item.validate(ctx, false)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Check if object exists
	exists := item.exists()
	if exists {
		return http.StatusConflict, errors.New(item.getExistsMessage(ctx))
	}

	err = tx.Set("gorm:association_autoupdate", false).Save(item).Error
	if err != nil {
		return http.StatusInternalServerError, errors.New("Unable to save item")
	}

	return http.StatusOK, nil
}
