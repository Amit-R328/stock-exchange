package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting server...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
