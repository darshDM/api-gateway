package main

import (
	"io"
	"log"
	"net/http"
)

func introduce(w http.ResponseWriter, r *http.Request) {
	log.Println("request came to handler")
	w.Header().Set("test", "header")
	io.WriteString(w, "Hii!, this is product service\n")
}
func main() {
	http.HandleFunc("/", introduce)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
