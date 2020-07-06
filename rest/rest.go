package rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"realizr.io/gome/model"

	"strconv"

	"github.com/gorilla/mux"
	"realizr.io/gome/ctrl"
)

const X_GOME_BRANCH = "X_GOME_BRANCH"
const X_GOME_HASH = "X_GOME_HASH"
const X_GOME_TYPE = "X_GOME_TYPE"
const X_GOME_ID = "X_GOME_ID"
const X_GOME_SEQ = "X_GOME_SEQ"
const DEFAULT_CONTENT_TYPE = "application/octetstream"
const REST_PORT = ":17456"

var controller *ctrl.LogEntryController

// Returns the int batch number if provided in request or default value if not.
func getRequestBatch(r *http.Request) int64 {
	batch, err := strconv.ParseInt(r.Header.Get(X_GOME_BRANCH), 10, 64)
	if err != nil {
		return int64(model.DEFAULT_BRANCH)
	} else {
		return batch
	}
}

// Returns the current entity type of the document.
func getRequestEntityType(r *http.Request) string {
	return r.Header.Get(X_GOME_TYPE)
}

// Returns the current hash of the document.
func getRequestHash(r *http.Request) string {
	return r.Header.Get(X_GOME_HASH)
}

func getRequestData(r *http.Request) []byte {
	body, _ := ioutil.ReadAll(r.Body)
	return body
}

// Returns the content type or default if not founmd.
func getRequestContentType(r *http.Request) string {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		return ct
	} else {
		return DEFAULT_CONTENT_TYPE
	}
}

func createArticle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Creating document...")
	entityType := getRequestEntityType(r)
	body := getRequestData(r)
	contentType := getRequestContentType(r)

	le := controller.Save(entityType, contentType, body)

	if le == nil {
		log.Fatalf("Failed to save document... ")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.Header().Add(X_GOME_BRANCH, strconv.FormatInt(le.Branch, 10))
		w.Header().Add(X_GOME_SEQ, strconv.FormatUint(le.Seq, 10))
		w.Header().Add(X_GOME_HASH, le.Hash)
		w.Header().Add(X_GOME_TYPE, le.EntityType)
		w.Header().Add(X_GOME_ID, le.Oid)
		w.WriteHeader(http.StatusOK)

	}
}

func updateArticle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Updating document...")
	vars := mux.Vars(r)
	oid := vars["oid"]
	branch := getRequestBatch(r)
	hash := getRequestHash(r)
	body := getRequestData(r)

	le := controller.Update(oid, branch, hash, body)

	if le == nil {
		log.Fatalf("Failed to updated document (id, branch, hash): %s %d %s", oid, branch, hash)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.Header().Add(X_GOME_BRANCH, strconv.FormatInt(le.Branch, 10))
		w.Header().Add(X_GOME_SEQ, strconv.FormatUint(le.Seq, 10))
		w.Header().Add(X_GOME_HASH, le.Hash)
		w.Header().Add(X_GOME_TYPE, le.EntityType)
		w.Header().Add(X_GOME_ID, le.Oid)
		w.WriteHeader(http.StatusOK)
	}
}

func fetchArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["oid"]
	branch := getRequestBatch(r)
	le := controller.Fetch(key, branch)

	if le != nil {
		//log.Printf("Content-type %s", le.ContentType)
		//log.Printf("Content-Length %s", strconv.Itoa(len(le.Data)))
		//log.Printf("%s %s", X_GOME_BRANCH, string(le.Branch))
		//log.Printf("%s %s", X_GOME_HASH, le.Hash)
		log.Printf("Serving document %s %d", key, branch)
		w.Header().Add("Content-Type", le.ContentType)
		w.Header().Add("Content-Length", strconv.FormatInt(int64(len(le.Data)), 10))
		w.Header().Add(X_GOME_BRANCH, strconv.FormatInt(le.Branch, 10))
		w.Header().Add(X_GOME_SEQ, strconv.FormatUint(le.Seq, 10))
		w.Header().Add(X_GOME_HASH, le.Hash)
		w.Header().Add(X_GOME_TYPE, le.EntityType)
		w.Header().Add(X_GOME_ID, le.Oid)
		w.WriteHeader(http.StatusOK)
		w.Write(le.Data)
	} else {
		log.Printf("Document not found %s %d", key, branch)
		w.WriteHeader(http.StatusNotFound)
	}
}

func fetchFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleRequests(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	log.Printf("Webserver listening on %s", REST_PORT)

	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	// 	myRouter.HandleFunc("/", homePage)
	//	myRouter.HandleFunc("/logs", returnAllArticles)
	myRouter.HandleFunc("/log/{oid}", fetchArticle).Methods("GET")
	myRouter.HandleFunc("/log/", createArticle).Methods("POST")
	myRouter.HandleFunc("/log/{oid}", updateArticle).Methods("PUT")
	myRouter.HandleFunc("/favicon.ico", fetchFavicon).Methods("GET")

	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	http.ListenAndServe(REST_PORT, myRouter)
}

func getLogEntryByOid(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, r.RequestURI)
}

func StartWebServer(ctrlr *ctrl.LogEntryController, wg *sync.WaitGroup) {
	controller = ctrlr

	go handleRequests(wg)
	//	http.HandleFunc("/", getLogEntryByOid)
	//	http.ListenAndServe("127.0.0.1:8880", nil)
}
