package main

import (
	"log"
	"os"
	"strconv"

	"time"

	"github.com/robfig/cron"
)

func setupCron() {
	var e error

	ce := false
	if os.Getenv("VM_ENABLE_CRON") != "" {
		ce, e = strconv.ParseBool(os.Getenv("VM_ENABLE_CRON"))

		if e != nil {
			log.Print("Cron jobs schedule err ", e)
			return
		}
	}

	cs := os.Getenv("VM_CRONS_RUN_AT")

	if ce && cs != "" {
		log.Print("Cron jobs schedule is enabled at ", cs)

		// Init
		c := cron.New()

		// every scheduled slot

		e = c.AddFunc(cs, sequentialJobs)
		if e != nil {
			return
		}

		log.Printf("%+v\n", c.Location())

		// Start
		c.Start()
	}
}

func sequentialJobs() {
	log.Print("Sequential jobs started at: ", time.Now().String)

	log.Print("Sequential jobs ended at: ", time.Now().String)
}
