package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	clientAddress = "http://localhost:8080"
	serverAddress = "http://localhost:8081"
)

type clientInfo struct {
	id            string
	secret        string
	scope         string
	redirect_uris []string
}

type httpError struct {
	StatusCode    int
	StatusMessage string
	ErrorMessage  string
}

var clients = map[string]clientInfo{
	"goauth": {
		id:            "goauth",
		secret:        "random-string",
		scope:         "all",
		redirect_uris: []string{clientAddress + "/callback"},
	},
}

var requests map[int]string

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(rootHandler))
	mux.HandleFunc("/authorize", http.HandlerFunc(authorizeHandler))
	mux.HandleFunc("/approve", func(w http.ResponseWriter, r *http.Request) {
	})
	fmt.Println("Listening on " + serverAddress)
	http.ListenAndServe(":8081", mux)
}

func newTemplate(filename string) *template.Template {
	templ, err := template.ParseFiles(filename)
	if err != nil {
		log.Fatalln(err)
	}
	return templ
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	indexTempl := newTemplate("index.gohtml")
	indexTempl.Execute(w, nil)
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	client_id := r.URL.Query().Get("client_id")
	client, ok := clients[client_id]
	if !ok {
		errMessage := fmt.Sprintf("Unkown client %s", client_id)
		renderError(w, http.StatusForbidden, errMessage)
		return
	}
	redirect_uri := r.URL.Query().Get("redirect_uri")
	if !contains(client.redirect_uris, redirect_uri) {
		errMessage := fmt.Sprintf("Mismatched redirect URI, expected one of %+v, got %s",
			client.redirect_uris,
			redirect_uri)
		renderError(w, http.StatusForbidden, errMessage)
		return
	}
	// TODO: Add support for a slice for scopes
	scope := r.URL.Query().Get("scope")
	// TODO: Client might not have any scopes
	if client.scope != scope {
		errMessage := fmt.Sprintf("Invalid scope, expected %s, got %s",
			client.scope, scope)
		renderError(w, http.StatusForbidden, errMessage)
		return
	}
	rand.Seed(time.Now().UnixNano())
	reqID := rand.Intn(65536)
	requests[reqID] = r.URL.Query().Encode()
	approveTempl := newTemplate("approve.gohtml")
	approveTemplData := struct {
		client clientInfo
		reqID  int
		scope  string
	}{
		client,
		reqID,
		scope,
	}
	approveTempl.Execute(w, approveTemplData)
}

func renderError(w http.ResponseWriter, status int, msg string) {
	errorTempl := newTemplate("error.gohtml")
	w.WriteHeader(status)
	whatHappened := httpError{
		StatusCode:    status,
		StatusMessage: http.StatusText(status),
		ErrorMessage:  msg,
	}
	errorTempl.Execute(w, whatHappened)
}

func contains(slice []string, target string) bool {
	for _, val := range slice {
		if val == target {
			return true
		}
	}
	return false
}
