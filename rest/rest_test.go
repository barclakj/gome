package rest

import (
	"log"
	"testing"
	"time"
)

func TestRest(t *testing.T) {
	log.Printf("Starting webserver...")
	go StartWebServer()
	log.Printf("Webserver running...")

	time.Sleep(15 * time.Second)
}
