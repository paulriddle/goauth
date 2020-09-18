package main

import (
	"html/template"
	"net/http"
)

type Info struct {
	Address string
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// render root template
		addr := Info{"127.0.0.1:8080"}
		t, err := template.ParseFiles("index.html")
		if err != nil {
			panic(err)
		}
		t.Execute(w, addr)
	})
	http.ListenAndServe(":8080", nil)
}
