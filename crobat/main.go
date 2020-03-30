package main

import (
	"fmt"
	"net/http"
	"go.elastic.co/apm/module/apmhttp"
)

func main() {
	fmt.Println("Staring Crobat API... You get it? Because Sonar? Sonar, Bats?")
	server := NewServer()
	http.ListenAndServe(":1338", apmhttp.Wrap(server.Router))
}
