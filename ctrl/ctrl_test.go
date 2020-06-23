package ctrl

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var wg sync.WaitGroup

func TestSaveNewLE(t *testing.T) {
	log.Printf("Testing save of new LE\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save([]byte("new le"))

	assert.Equal(t, true, le != nil)
	if le != nil {
		assert.Equal(t, uint64(1), le.Seq)
	}
	time.Sleep(2 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}

func TestUpdateLE(t *testing.T) {
	log.Printf("Testing save of new LE\n")

	ctrl := LogEntryController{}
	ctrl.Init(&wg)

	le := ctrl.Save([]byte("new le"))

	assert.Equal(t, true, le != nil)
	if le != nil {
		assert.Equal(t, uint64(1), le.Seq)

		le = ctrl.Update(le.Oid, 0, []byte("this is some new data"))
		assert.Equal(t, true, le != nil)
		assert.Equal(t, uint64(2), le.Seq)

		le = ctrl.Update(le.Oid, 0, []byte("and some more data"))
		assert.Equal(t, true, le != nil)
		assert.Equal(t, uint64(3), le.Seq)

	}
	time.Sleep(2 * time.Second)
	ctrl.Alive = false
	wg.Wait()
}
