package main

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
)

// logged in user object
var LoggedInUser User

func apiCreateHandler(ctx echo.Context) error {
	var err error

	item, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// Populate object from JSON
	err = ctx.Bind(&item)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	tx := db.Begin()

	code, err := apiCreate(ctx, tx, item, true)
	if err != nil {
		tx.Rollback()
		return ctx.JSON(code, err.Error())
	}

	tx.Commit()

	return ctx.JSON(code, item)
}

func apiReadHandler(ctx echo.Context) error {
	var err error

	item, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	uuid := ctx.Param("uuid")
	if uuid == "new" {
		return ctx.JSON(http.StatusOK, item)
	}

	if db.Unscoped().Where("uuid = ?", uuid).First(item).RecordNotFound() {
		return ctx.NoContent(http.StatusNotFound)
	}

	item.postRead()

	return ctx.JSON(http.StatusOK, item)
}

func apiBulkReadHandler(ctx echo.Context) error {
	var (
		uuids []string
		err   error
	)

	item, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// Populate object from JSON
	err = ctx.Bind(&uuids)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	items, err := item.bulkRead(uuids)
	if err != nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	return ctx.JSON(http.StatusOK, items)
}

func apiSearchHandler(ctx echo.Context) error {

	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	} else {
		LoggedInUser = u
	}

	var params SearchQuery

	item, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// Populate object from JSON
	err = ctx.Bind(&params)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if params.Limit <= 0 || params.Limit >= 10 {
		params.Limit = 10
	}

	items, err := item.doSearch(params)
	if err != nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	return ctx.JSON(http.StatusOK, items)
}

func apiUpdateHandler(ctx echo.Context) error {
	var err error

	item, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	old, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	err = db.Where("uuid = ?", ctx.Param("uuid")).First(item).Error
	if err != nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	bb, err := json.Marshal(item)
	if err == nil {
		err = json.Unmarshal(bb, &old)
		if err != nil {
			return returnInvalidData(ctx, err)
		}
	}

	// Populate object from JSON
	err = ctx.Bind(&item)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	tx := db.Begin()

	code, err := apiUpdate(ctx, tx, item, true)
	if err != nil {
		tx.Rollback()
		return ctx.JSON(code, err.Error())
	}

	tx.Commit()

	return ctx.JSON(code, item)
}

func apiDeleteHandler(ctx echo.Context) error {
	var err error

	item, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	fail := db.Where("uuid = ?", ctx.Param("uuid")).First(item).RecordNotFound()
	if fail {
		return ctx.NoContent(http.StatusNotFound)
	}

	tx := db.Begin()

	err = tx.Delete(item).Error
	if err != nil {
		tx.Rollback()

		return ctx.JSON(http.StatusBadRequest, "Unable to delete item")
	}

	tx.Commit()

	return ctx.NoContent(http.StatusOK)
}

func apiRestoreHandler(ctx echo.Context) error {
	var err error

	item, err := getAPIItem(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	fail := db.Unscoped().Where("uuid = ?", ctx.Param("uuid")).First(item).RecordNotFound()
	if fail {
		return ctx.NoContent(http.StatusNotFound)
	}

	tx := db.Begin()

	err = item.restore(ctx, tx)
	if err != nil {
		tx.Rollback()

		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	tx.Commit()

	return ctx.JSON(http.StatusOK, item)
}
