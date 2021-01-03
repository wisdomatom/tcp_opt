package main

import (
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, req *http.Request) {
	// num := req.Header.Get("num")
	// log.Printf("total: %v", num)
	// time.Sleep(time.Second * 1)
	w.Write([]byte(`{"status": "ok"}`))
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", nil))
}
