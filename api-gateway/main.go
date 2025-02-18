package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func productHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	proxyReq, err := http.NewRequest(req.Method, "http://product.api:3000/", bytes.NewReader(body))
	if err != nil {
		http.Error(res, "There are issues with request url", http.StatusBadRequest)
	}
	log.Println("after making")
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}

	client := &http.Client{}
	response, err := client.Do(proxyReq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	log.Println(body, response)
	original_header := res.Header()
	for h, val := range response.Header {
		original_header[h] = val
	}
	io.Copy(res, response.Body)
	response.Body.Close()
}

func vendorHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	proxyReq, err := http.NewRequest(req.Method, "http://vendor.api:3001/", bytes.NewReader(body))
	if err != nil {
		http.Error(res, "There are issues with request url", http.StatusBadRequest)
	}
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}

	client := &http.Client{}
	response, err := client.Do(proxyReq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	original_header := res.Header()
	for h, val := range response.Header {
		original_header[h] = val
	}
	io.Copy(res, response.Body)
	response.Body.Close()
}
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/product", productHandler)
	router.HandleFunc("/vendor", vendorHandler)

	// To handle host based routing
	// productRouter := router.Host("product.api").Subrouter()
	// productRouter.PathPrefix("/product").HandlerFunc(productHandler)

	// vendorRouter := router.Host("vendor.api").Subrouter()
	// vendorRouter.PathPrefix("/vendor").HandlerFunc(vendorHandler)

	http.Handle("/", router)

	log.Println("Starting API gateway on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))

}
