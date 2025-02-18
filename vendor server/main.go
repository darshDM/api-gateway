package main

import (
	"io"
	"log"
	"net/http"
)

func introduce(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hii!, this is vendor service\n")
}
func main() {
	http.HandleFunc("/", introduce)
	log.Fatal(http.ListenAndServe(":3001", nil))
}
