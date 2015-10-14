package main

import (
	"fmt"
	"log"
	"net/http"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-Access-Token")
	if token == "magic" {
		fmt.Fprintf(w, "You have some magic in you\n")
		log.Println("Allowed an access attempt")
	} else {
		http.Error(w, "You don't have enough magic in you", http.StatusForbidden)
		log.Println("Denied an access attempt")
	}
}

func main() {
	http.HandleFunc("/", mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
