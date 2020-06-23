package rest

import (
	"fmt"
	"net/http"
)

func getLogEntryByOid(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, r.RequestURI)
}

func StartWebServer() {
	http.HandleFunc("/", getLogEntryByOid)
	http.ListenAndServe("127.0.0.1:8880", nil)
}
