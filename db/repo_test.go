package db

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"realizr.io/gome/model"
)

func TestAppend(t *testing.T) {
	le := model.LogEntry{}
	le.Uuid = "uuid"
	le.Seq = 1
	le.RemoteTs = 1
	le.Ts = 1
	le.Origin = "origin"
	assert.Equal(t,true,Append(&le))
	log.Printf("Done testing!\n")
}
