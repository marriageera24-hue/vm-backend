package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	e := godotenv.Load()
	if e != nil {
		log.Print("error while loading env vars ", e)
	}

	env := os.Getenv("VM_ENVIRONMENT")

	if isOneOf(env, []string{"local", "production"}) {
		e = godotenv.Load(getEnvFilePath(".env"))
		if e != nil {
			log.Println("error while loading env vars ", e)
		}
	}
}

func main() {

	err := initI18n()
	if err != nil {
		log.Fatal(err)
	}

	openDatabaseConnection()
	migrate()
	setupCron()
	startServer()
}
