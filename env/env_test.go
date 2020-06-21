package env

import (
	// "github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestGetHostname(t *testing.T) {
	log.Printf("FQDN: %s\n", GetHostname())
}

func TestGetUser(t *testing.T) {
	log.Printf("User: %s\n", GetUser())
}

func TestGetOrigin(t *testing.T) {
	log.Printf("Origin: %s\n", GetOrigin())
}
