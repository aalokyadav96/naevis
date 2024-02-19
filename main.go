package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/",Hello)
	http.ListenAndServe("localhost:4000",nil)
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"Hi")
}