package db

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"realizr.io/gome/model"
)

func TestTen(t *testing.T) {
	for i := 1; i < 10; i++ {
		le := model.NewLogEntry([]byte("simple test"), "origin")
		InsertLogEntry(le)
	}
}

func TestTenUpdates(t *testing.T) {
	le := model.NewLogEntry([]byte("simple test"), "origin")
	InsertLogEntry(le)
	for i := 1; i < 10; i++ {
		le.Seq = le.Seq + 1
		InsertLogEntry(le) // note that hash here will be wrong but we're testing DB, not hashed updates
	}

	le2 := FetchLatestLogEntry(le.Oid)
	log.Printf("Found object %s with latest seq %d \n", le2.Oid, le2.Seq)
	assert.Equal(t, le2.Seq, uint64(10))
	assert.Equal(t, le2.Oid, le.Oid)
}

func perfTest(num int) string {
	started := time.Now().UnixNano()
	le := model.NewLogEntry([]byte("simple test"), "origin")
	InsertLogEntry(le)
	for i := 1; i < num; i++ {
		le.Seq = le.Seq + 1
		InsertLogEntry(le) // note that hash here will be wrong but we're testing DB, not hashed updates
	}
	endInsert := time.Now().UnixNano()

	le2 := FetchLatestLogEntry(le.Oid)
	log.Printf("Found object %s with latest seq %d \n", le2.Oid, le2.Seq)
	endQuery := time.Now().UnixNano()
	log.Printf("%d records, Insert %dms, Fetch %dms, Total %dms", num, ((endInsert - started) / 1000000), ((endQuery - endInsert) / 1000000), ((endQuery - started) / 1000000))
	return le.Oid
}

func perfTestFetch(oid string) {
	started := time.Now().UnixNano()

	// oid := "urn:uuid:5f6c5a8c-64e7-46f5-af34-f2b862029133"

	le2 := FetchLatestLogEntry(oid)
	log.Printf("Found object %s with latest seq %d \n", le2.Oid, le2.Seq)
	endQuery := time.Now().UnixNano()
	log.Printf("%d records, Fetched %dms\n", le2.Seq, ((endQuery - started) / 1000000))
}

func TestPerf(t *testing.T) {
	perfTest(100)
	perfTest(200)
	perfTest(300)
	perfTest(400)
	oid := perfTest(500)

	for i := 1; i < 10; i++ {
		perfTestFetch(oid)
	}
}
