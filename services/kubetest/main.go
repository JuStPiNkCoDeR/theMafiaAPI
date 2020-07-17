package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const port = 8080

func testHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello!")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", testHandler)

	fmt.Print("Starting...")
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)

	if err != nil {
		log.Fatalf("%v", err)
	}
}
