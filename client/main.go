package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type credentials struct {
	accessToken string
}

type serverError struct {
	StatusCode int
	Message    string
}

func main() {
	indexTempl := newTemplate("index.html")
	errorTempl := newTemplate("error.html")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var creds credentials
		indexTempl.Execute(w, creds)
	})

	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
	})
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
	})

	mux.HandleFunc("/fetch_resource", func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie("accessToken")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			whatHappened := serverError{
				StatusCode: http.StatusUnauthorized,
				Message:    http.StatusText(http.StatusUnauthorized),
			}
			errorTempl.Execute(w, whatHappened)
			return
		}
		creds := credentials{accessToken: accessToken.Name}
		indexTempl.Execute(w, creds)
	})

	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func newTemplate(filename string) *template.Template {
	templ, err := template.ParseFiles(filename)
	if err != nil {
		log.Fatalln(err)
	}
	return templ
}
