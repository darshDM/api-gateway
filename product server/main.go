package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func introduce(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Header.Get("x-forwarded-for"), r.RemoteAddr)
	io.WriteString(w, "Hii!, this is product service\n")
}
func main() {
	http.HandleFunc("/", introduce)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
