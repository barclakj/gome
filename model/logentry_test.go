package model

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestCreateLE(t *testing.T) {
	data := []byte("This is a simple test...")
	log.Printf("Testing Create LE!\n")
	le := NewLogEntry(data, "origin")
	assert.Equal(t,uint64(1), le.Seq)

	var nille *LogEntry = nil
	le = NewLogEntry(data, "")
	assert.Equal(t,nille, le)

	le = NewLogEntry(nil, "origin")
	assert.Equal(t, nille, le)
}

func TestLEJSON(t  *testing.T) {
	log.Printf("Testing JSON serialization/deserialization\n")
	le := NewLogEntry([]byte("Test"), "origin")
	jsonString := le.ToJSON()
	log.Printf(jsonString)

	le2 := FromJSON([]byte(jsonString))
	assert.Equal(t, "origin", le2.Origin)
	assert.Equal(t, le.Ts, le2.Ts)
	assert.Equal(t, le.Uuid, le2.Uuid)
	assert.Equal(t, le.RemoteTs, le2.RemoteTs)
	assert.Equal(t, []byte("Test"), le2.Data)
}

func TestUpdateLE(t *testing.T) {
	log.Printf("Testing update of LE\n")
	le := NewLogEntry([]byte("new le"), "origin")

	s := le.Seq

	le.Update([]byte("updated data"), "origin2")
	assert.Equal(t, s+1, le.Seq)
	assert.Equal(t,"origin2", le.Origin)
}
