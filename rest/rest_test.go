package rest

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"realizr.io/gome/ctrl"
)

func TestRestGET(t *testing.T) {
	log.Printf("Starting webserver...")

	ctrl := ctrl.LogEntryController{}

	le := ctrl.Save("TEST", "application/octetstream", []byte("version 1"))

	prestr := le.Data
	log.Printf("expecting: %s", string(prestr))

	log.Printf("ID: %s %d", le.Oid, le.Branch)

	log.Printf("Data: %s", string(le.Data))

	go StartWebServer()
	log.Printf("Webserver running...")
	time.Sleep(1 * time.Second)

	url := "http://localhost:17456/log/" + le.Oid
	log.Printf("Requesting: " + url)

	req, _ := http.NewRequest("GET", url, nil)

	res, _ := http.DefaultClient.Do(req)
	assert.Equal(t, res.StatusCode, 200)
	fbranch, _ := strconv.ParseInt(res.Header.Get(X_GOME_BRANCH), 10, 64)
	assert.Equal(t, fbranch, le.Branch)
	assert.Equal(t, res.Header.Get(X_GOME_HASH), string(le.Hash))
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	log.Printf("response: %s", string(body))

	assert.Equal(t, string(body), string(prestr))

}

func TestRestPut(t *testing.T) {
	log.Printf("Starting webserver...")

	ctrl := ctrl.LogEntryController{}

	le := ctrl.Save("TEST", "application/octetstream", []byte("version 1"))

	log.Printf("ID: %s %d", le.Oid, le.Branch)
	log.Printf("Data: %s", string(le.Data))

	url := "http://localhost:17456/log/" + le.Oid
	log.Printf("Updating by URL: %s", url)

	req, _ := http.NewRequest("PUT", url, strings.NewReader("This is some updated text!"))
	log.Printf("Updating on branch: %d", le.Branch)
	req.Header.Add(X_GOME_BRANCH, strconv.FormatInt(le.Branch, 10))
	log.Printf("Updating from hash: %s", le.Hash)
	req.Header.Add(X_GOME_HASH, le.Hash)

	log.Print("Doing... ")
	res, err := http.DefaultClient.Do(req)
	log.Printf("Done!")
	if err == nil {
		log.Printf("Response StatusCode: %d", res.StatusCode)
		assert.Equal(t, 200, res.StatusCode)
	} else {
		log.Fatalf("Error updating resource: %s", err)
	}
}
