package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
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

// Assuming DBConn *sql.DB is globally available or passed to the handler
func apiAdminSearchHandler(ctx echo.Context) error {

	// 1. Authentication is handled by middleware (httpAuth)

	// 2. Bind the query parameters (from URL query string)
	// var params SearchQuery
	// if err := ctx.Bind(&params); err != nil {
	// 	return ctx.JSON(http.StatusBadRequest, "Invalid query parameters: "+err.Error())
	// }

	// 3. (Optional) Enforce default limit
	// if params.Limit <= 0 || params.Limit > 50 {
	// 	params.Limit = 10
	// }

	// 4. --- START: DIRECT DB LOGIC ---
	var users []User

	db = db.Where("deleted_at IS NULL").Order("id DESC")
	// .Find(&users) executes the query and scans results directly into the slice
	/*************  ✨ Windsurf Command 🌟  *************/
	result := db.Find(&users)
	/*******  17322107-6f58-4f23-8112-37095443682f  *******/
	if result.Error != nil {
		log.Printf("PostgreSQL GORM Query Error: %v", result.Error)
		return ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("DB query failed: %v", result.Error))
	}
	// 5. Success
	return ctx.JSON(http.StatusOK, users)
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

func apiAdminDeleteHandler(ctx echo.Context) error {
	// 1. Declare the variable as a single struct, not a slice.
	var (
		user User
		err  error
		// data struct {
		// 	Email string `json:"email"`
		// 	Token string `json:"token"`
		// }
		data struct {
			Uuid string `json:"uuid"`
		}
	)

	err = ctx.Bind(&data)
	if err != nil {
		return returnInvalidData(ctx, err)
	}
	print("in error 238")
	print(ctx.Param("uuid"))
	// 2. Fetch the user's record into the struct. Pass the pointer to GORM.
	// If successful, 'user' struct is populated.
	result := db.Where("uuid = ?", data.Uuid).First(&user) // Pass pointer: &user

	// 3. Check the result for errors (including not found)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ctx.NoContent(http.StatusNotFound) // Return 404 if not found
		}
		// Handle other DB errors (e.g., connection issues)
		return ctx.JSON(http.StatusInternalServerError, "Database error during fetch")
	}

	// 5. Delete the *fetched* struct. GORM uses the struct's ID for the WHERE clause.
	res := db.Exec("UPDATE users SET deleted_at = now() WHERE uuid = ?", data.Uuid)
	if res.Error != nil {
		// There was a database execution error.
		return ctx.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return ctx.JSON(http.StatusOK, gettext("User has been Deleted", ctx))
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
