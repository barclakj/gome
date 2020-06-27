package model

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	data := []byte("This is a simple test...")
	le := NewLogEntry(data, "")
	assert.Equal(t, false, le.Validate())
	le = NewLogEntry([]byte(""), "test")
	assert.Equal(t, false, le.Validate())
	le = NewLogEntry(data, "test")
	le.Oid = ""
	assert.Equal(t, false, le.Validate())
	le = NewLogEntry(data, "test")
	assert.Equal(t, true, le.Validate())

	le = NewLogEntry(data, "test")
	le.Seq = 0
	assert.Equal(t, false, le.Validate())

	le = NewLogEntry(data, "test")
	le.Branch = -1
	assert.Equal(t, false, le.Validate())
}

func TestCreateLE(t *testing.T) {
	data := []byte("This is a simple test...")
	log.Printf("Testing Create LE!\n")
	le := NewLogEntry(data, "origin")
	assert.Equal(t, uint64(1), le.Seq)
}

func TestLEJSON(t *testing.T) {
	log.Printf("Testing JSON serialization/deserialization\n")
	le := NewLogEntry([]byte("Test"), "origin")
	jsonString := le.ToJSON()
	log.Printf(jsonString)

	le2 := FromJSON([]byte(jsonString))
	assert.Equal(t, "origin", le2.Origin)
	assert.Equal(t, le.Ts, le2.Ts)
	assert.Equal(t, le.Oid, le2.Oid)
	assert.Equal(t, le.OriginTs, le2.OriginTs)
	assert.Equal(t, []byte("Test"), le2.Data)
}
