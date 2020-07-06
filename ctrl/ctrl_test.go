package ctrl

import (
	"log"
	"sync"
	"testing"
	"time"

	"realizr.io/gome/db"
	"realizr.io/gome/env"
	"realizr.io/gome/model"

	"github.com/stretchr/testify/assert"
)

var wg sync.WaitGroup

func TestSaveNewLE(t *testing.T) {
	log.Printf("Testing save of new LE\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save("TEST", "application/octetstream", []byte("new le"))

	assert.Equal(t, true, le != nil)
	if le != nil {
		assert.Equal(t, uint64(1), le.Seq)
	}
	time.Sleep(1 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}

func TestUpdateLE(t *testing.T) {
	log.Printf("Testing save of new LE\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save("TEST", "application/octetstream", []byte("new le"))

	assert.Equal(t, true, le != nil)
	if le != nil {
		assert.Equal(t, uint64(1), le.Seq)

		le = ctrl.Update(le.Oid, 0, le.Hash, []byte("this is some new data"))
		assert.Equal(t, true, le != nil)
		assert.Equal(t, uint64(2), le.Seq)

		le = ctrl.Update(le.Oid, 0, le.Hash, []byte("and some more data"))
		assert.Equal(t, true, le != nil)
		assert.Equal(t, uint64(3), le.Seq)

	}
	time.Sleep(1 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}

func TestLETree(t *testing.T) {
	log.Printf("Testing tree of log entries\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save("TEST", "application/octetstream", []byte("version 1"))
	log.Printf("%s", le.ToJSON())
	le.Origin = "noone@0.0.0.0"

	le.Seq = 2
	le.Data = []byte("version 2")
	le.Hash = model.Hash(le.Oid, le.Hash, le.Data)

	ctrl.FromRemote(le)
	log.Printf("%s", le.ToJSON())

	leJSON := le.ToJSON()

	le.Seq = 3
	le.Data = []byte("version 3")
	le.Hash = model.Hash(le.Oid, le.Hash, le.Data)

	ctrl.FromRemote(le)
	log.Printf("%s", le.ToJSON())

	leJSON3 := le.ToJSON()

	le.Seq = 4
	le.Data = []byte("version 4")
	le.Hash = model.Hash(le.Oid, le.Hash, le.Data)
	ctrl.FromRemote(le)
	log.Printf("%s", le.ToJSON())

	le2 := model.FromJSON([]byte(leJSON))
	le2.Seq = 3
	le2.Data = []byte("version 3.1")
	le2.Hash = model.Hash(le2.Oid, le2.Hash, le2.Data)
	ctrl.FromRemote(le2)
	log.Printf("%s", le2.ToJSON())

	le3 := model.FromJSON([]byte(leJSON3))
	le3.Seq = 4
	le3.Data = []byte("version 4.2")
	le3.Hash = model.Hash(le3.Oid, le3.Hash, le3.Data)
	ctrl.FromRemote(le3)
	log.Printf("%s", le3.ToJSON())

	le3.Seq = 5
	le3.Data = []byte("version 5.2")
	le3.Hash = model.Hash(le3.Oid, le3.Hash, le3.Data)
	ctrl.FromRemote(le3)
	log.Printf("%s", le3.ToJSON())

	leJSON4 := le3.ToJSON()

	le3.Seq = 6
	le3.Data = []byte("version 6.2")
	le3.Hash = model.Hash(le3.Oid, le3.Hash, le3.Data)
	ctrl.FromRemote(le3)
	log.Printf("%s", le3.ToJSON())

	le4 := model.FromJSON([]byte(leJSON4))
	le4.Seq = 6
	le4.Data = []byte("version 6.3")
	le4.Hash = model.Hash(le4.Oid, le4.Hash, le4.Data)
	ctrl.FromRemote(le4)
	log.Printf("%s", le4.ToJSON())

	assert.Equal(t, int64(0), le.Branch)
	assert.Equal(t, int64(1), le2.Branch)
	assert.Equal(t, int64(2), le3.Branch)
	assert.Equal(t, int64(3), le4.Branch)
	assert.Equal(t, le3.Branch, le4.PreviousBranch)
	assert.Equal(t, le.Branch, le2.PreviousBranch)

	time.Sleep(1 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}

func TestLEStash(t *testing.T) {
	var nile *model.LogEntry
	log.Printf("Testing stash of log entries\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save("TEST", "application/octetstream", []byte("version 1"))
	log.Printf("%s", le.ToJSON())
	le.Origin = "noone@0.0.0.0"

	le.Seq = 2
	le.Data = []byte("version 2")
	le.Hash = model.Hash(le.Oid, le.Hash, le.Data)
	le2JSON := le.ToJSON()

	le.Seq = 3
	le.Data = []byte("version 3")
	le.Hash = model.Hash(le.Oid, le.Hash, le.Data)
	le3JSON := le.ToJSON()

	le.Seq = 4
	le.Data = []byte("version 4")
	le.Hash = model.Hash(le.Oid, le.Hash, le.Data)
	le4JSON := le.ToJSON()

	le = model.FromJSON([]byte(le4JSON))
	assert.Equal(t, nile, ctrl.FromRemote(le))

	le = model.FromJSON([]byte(le3JSON))
	assert.Equal(t, nile, ctrl.FromRemote(le))

	le = model.FromJSON([]byte(le2JSON))
	le = ctrl.FromRemote(le)
	if le != nil {
		assert.Equal(t, uint64(4), le.Seq)
	} else {
		assert.Equal(t, true, false)
	}

	time.Sleep(1 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}

func TestLEDuplicate(t *testing.T) {
	var nile *model.LogEntry
	log.Printf("Testing duplicate log entries\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save("TEST", "application/octetstream", []byte("version 1"))
	le.Origin = "noone@0.0.0.0"
	log.Printf("%s", le.ToJSON())

	le.Seq = 2
	le.Data = []byte("version 2")
	le.Hash = model.Hash(le.Oid, le.Hash, le.Data)

	assert.NotEqual(t, nile, ctrl.FromRemote(le)) // should save ok
	assert.Equal(t, nile, ctrl.FromRemote(le))    // should ignore

	time.Sleep(1 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}

func TestLENewRemote(t *testing.T) {
	var nile *model.LogEntry
	log.Printf("Testing new log from remote\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := model.NewLogEntry("TEST", "application/octetstream", []byte("version 1"), "noone@0.0.0.0")
	le.Hash = model.Hash(le.Oid, "", le.Data)
	log.Printf("%s", le.ToJSON())

	assert.NotEqual(t, nile, ctrl.FromRemote(le)) // should save ok

	time.Sleep(1 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}

func TestCMDReplay(t *testing.T) {
	log.Printf("Testing REPLAY COMMAND\n")
	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save("TEST", "application/octetstream", []byte("version 1"))
	db.AddObserver(le.Oid, "127.0.0.1:7456")
	for i := 0; i < 10; i++ {
		le = ctrl.Update(le.Oid, le.Branch, le.Hash, []byte("new"))
	}
	env.AllowLocal = true

	ctrl.RequestReplay(le.Oid, le.Branch)

	time.Sleep(1 * time.Second)
	env.AllowLocal = false
	ctrl.Alive = false
	wg.Wait()
}
