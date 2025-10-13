package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

type Wallet struct {
	gorm.Model
	UserUUID string  `gorm:"type:uuid;index;" json:"user_uuid"`
	Balance  float64 `json:"balance"`
}

func getBalance(ctx echo.Context) error {

	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	var w Wallet
	UUID := u.UUID
	log.Println(UUID)

	err = db.Model(&Wallet{}).Where("user_uuid = ?", UUID).First(&w).Error

	if err != nil {
		log.Println(err.Error())
		return ctx.NoContent(http.StatusNotFound)
	}
	fmt.Print(w)
	return ctx.JSON(http.StatusOK, w)

}
func updateWallet() {
	// Create the trigger
	err := db.Exec(`
        CREATE FUNCTION add_amount_to_wallet() RETURNS TRIGGER AS $$
        BEGIN
		IF EXISTS (SELECT 1 FROM wallets WHERE user_uuid = NEW.user_uuid) THEN
            UPDATE wallets
            SET balance = balance + (NEW.data->>'amount'::text)::numeric
            WHERE user_uuid = NEW.user_uuid;
            RETURN NEW;
		ELSE
			INSERT INTO wallets (user_uuid, balance)
			VALUES (NEW.user_uuid, (NEW.data->>'amount'::text)::numeric);
			RETURN NEW;
		END IF;
        END;
        $$ LANGUAGE plpgsql;

        CREATE TRIGGER add_amount_trigger
        AFTER INSERT ON payments
        FOR EACH ROW
        EXECUTE PROCEDURE add_amount_to_wallet();
    `)
	if err != nil {
		panic(err)
	}

	fmt.Println("Trigger created successfully")
}
