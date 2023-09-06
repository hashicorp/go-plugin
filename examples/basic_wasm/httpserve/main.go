package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("p", 8080, "port")
	dir := flag.String("dir", ".", "root directory")
	flag.Parse()

	handler := http.FileServer(http.Dir(*dir))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "basic", "greeter":
			w.Header().Set("Content-Type", "application/wasm")
		}
		handler.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Println("serving at " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
