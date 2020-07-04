package rest

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"realizr.io/gome/ctrl"
)

func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	ctrl := ctrl.LogEntryController{}

	le := ctrl.Fetch(key, 0)

	if le != nil {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, le.ToJSON())
	}
}

func handleRequests() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	// 	myRouter.HandleFunc("/", homePage)
	//	myRouter.HandleFunc("/logs", returnAllArticles)
	myRouter.HandleFunc("/log/{id}", returnSingleArticle)
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	log.Fatal(http.ListenAndServe(":17456", myRouter))
}

func getLogEntryByOid(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, r.RequestURI)
}

func StartWebServer() {
	handleRequests()
	//	http.HandleFunc("/", getLogEntryByOid)
	//	http.ListenAndServe("127.0.0.1:8880", nil)
}
