package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Staring Zubat API...")
	server := NewServer()
	http.ListenAndServe(":1337", server.Router)
}
