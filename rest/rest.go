package rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"realizr.io/gome/model"

	"strconv"

	"github.com/gorilla/mux"
	"realizr.io/gome/ctrl"
)

const X_GOME_BRANCH = "X_GOME_BRANCH"
const X_GOME_HASH = "X_GOME_HASH"

var controller ctrl.LogEntryController

// Returns the int batch number if provided in request or default value if not.
func getRequestBatch(r *http.Request) int64 {
	batch, err := strconv.ParseInt(r.Header.Get(X_GOME_BRANCH), 10, 64)
	if err != nil {
		return int64(model.DEFAULT_BRANCH)
	} else {
		return batch
	}
}

// Returns the current hash of the document.
func getRequestHash(r *http.Request) string {
	return r.Header.Get(X_GOME_HASH)
}

func getRequestData(r *http.Request) []byte {
	body, _ := ioutil.ReadAll(r.Body)
	return body
}

func createArticle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Creating document...")
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
	}
}

func fetchArticle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Fetching document...")
	vars := mux.Vars(r)
	key := vars["oid"]
	branch := getRequestBatch(r)
	le := controller.Fetch(key, branch)

	if le != nil {
		//log.Printf("Content-type %s", le.ContentType)
		//log.Printf("Content-Length %s", strconv.Itoa(len(le.Data)))
		//log.Printf("%s %s", X_GOME_BRANCH, string(le.Branch))
		//log.Printf("%s %s", X_GOME_HASH, le.Hash)
		w.Header().Add("Content-Type", le.ContentType)
		w.Header().Add("Content-Length", strconv.FormatInt(int64(len(le.Data)), 10))
		w.Header().Add(X_GOME_BRANCH, strconv.FormatInt(le.Branch, 10))
		w.Header().Add(X_GOME_HASH, le.Hash)
		w.WriteHeader(http.StatusOK)
		w.Write(le.Data)
	}
}

func handleRequests() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	// 	myRouter.HandleFunc("/", homePage)
	//	myRouter.HandleFunc("/logs", returnAllArticles)
	myRouter.HandleFunc("/log/{oid}", fetchArticle).Methods("GET")
	myRouter.HandleFunc("/log", createArticle).Methods("POST")
	myRouter.HandleFunc("/log/{oid}", updateArticle).Methods("PUT")

	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	http.ListenAndServe(":17456", myRouter)
}

func getLogEntryByOid(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, r.RequestURI)
}

func StartWebServer() {
	handleRequests()
	//	http.HandleFunc("/", getLogEntryByOid)
	//	http.ListenAndServe("127.0.0.1:8880", nil)
}
