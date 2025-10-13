package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

type Payment struct {
	gorm.Model
	UserUUID string          `gorm:"type:uuid;index;" json:"user_uuid"`
	Success  bool            `json:"success"`
	Code     string          `json:"code"`
	Message  string          `json:"message"`
	Data     json.RawMessage `json:"data"` // Store PaymentData as JSON
}

func savePayments(ctx echo.Context) error {
	u, errr := verifySession(ctx)
	if errr != nil {
		return ctx.JSON(http.StatusUnauthorized, errr.Error())
	}

	var (
		p              Payment
		paymentData    map[string]json.RawMessage
		responseBytes  []byte
		responseString string
		decodedBody    []byte
		err            error
	)

	// Read the request body as a byte array
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		log.Println("Error reading request body:", err.Error())
		return returnInvalidData(ctx, err)
	}

	// Decode the JSON data directly
	err = json.Unmarshal(body, &paymentData)
	if err != nil {
		log.Println("Error unmarshalling JSON data:", err.Error())
		return returnInvalidData(ctx, err)
	}

	// Access the "response" data
	responseBytes, ok := paymentData["response"]
	if !ok {
		log.Println("Error: Missing \"response\" key in request body")
		return returnInvalidData(ctx, err)
	}

	responseString = strings.TrimSpace(string(responseBytes))
	responseString = responseString[1 : len(responseString)-1]

	// Attempt to decode the base64 data with error handling
	decodedBody, err = base64.StdEncoding.DecodeString(responseString)
	if err != nil {
		log.Println("Error decoding base64 data:", err.Error())
		// Handle the error appropriately, e.g., return an error to the client
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid base64-encoded data")
	}

	// Unmarshal the decoded data into the Payment struct
	err = json.Unmarshal(decodedBody, &p)
	if err != nil {
		log.Println("Error unmarshalling JSON data:", err.Error())
		return returnInvalidData(ctx, err)
	}
	p.CreatedAt = time.Now()
	p.UserUUID = u.UUID

	// Use sql.NullInt64 to handle potential NULL values from the database
	var nullID sql.NullInt64
	err = db.Save(&p).Scan(&nullID).Error // Save data and scan for the ID

	if err != nil {
		log.Println("Error while saving payment to DB:", err.Error())
		return err
	}

	// Check if the ID is valid (not NULL) before using it
	if nullID.Valid {
		fmt.Printf("Payment saved successfully with ID: %d\n", nullID.Int64)
	} else {
		fmt.Println("Payment saved, but ID is NULL")
	}

	return nil
}
