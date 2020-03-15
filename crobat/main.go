package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Staring Crobat API... You get it? Because Sonar? Sonar, Bats?")
	server := NewServer()
	http.ListenAndServe(":1338", server.Router)
}
