package db

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"realizr.io/gome/model"
)

func TestAppend(t *testing.T) {
	le := model.NewLogEntry([]byte("simple test"), "origin")
	log.Printf("UUID: %s SEQ: %d\n", le.Uuid, le.Seq)
	assert.Equal(t,true,Append(le))
	le.Update([]byte("my data"), "origin2")
	log.Printf("UUID: %s SEQ: %d\n", le.Uuid, le.Seq)
	assert.Equal(t,true,Append(le))
	log.Printf("Done testing!\n")
}
