package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintln(w, "Hello from "+vars["number"])
}

// Hello HTTP Server
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/hello/{number}", handler)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
