package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func main() {
	http.HandleFunc("/", getRoot)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3473"
	}
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
