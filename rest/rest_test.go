package rest

import (
	"log"
	"testing"
	"time"

	"realizr.io/gome/ctrl"
)

func TestRest(t *testing.T) {
	log.Printf("Starting webserver...")

	ctrl := ctrl.LogEntryController{}

	le := ctrl.Save([]byte("version 1"))
	log.Printf("ID: %s %d", le.Oid, le.Branch)

	log.Printf("Data: %s", string(le.Data))

	go StartWebServer()
	log.Printf("Webserver running...")

	time.Sleep(15 * time.Second)
}
