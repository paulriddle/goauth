package main

import (
	"html/template"
	"net/http"
)

type Info struct {
	Address string
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		addr := Info{"127.0.0.1:8080"}
		t, err := template.ParseFiles("index.html")
		if err != nil {
			panic(err)
		}
		t.Execute(w, addr)
	})
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
	})
	mux.HandleFunc("/approve", func(w http.ResponseWriter, r *http.Request) {
	})
	http.ListenAndServe(":8080", mux)
}
