package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

const (
	serverAddress     = "http://localhost:8080"
	authServerAddress = "http://localhost:8081"
	protectedResource = "http://localhost:8082"
)

type clientInfo struct {
	id            string
	secret        string
	scope         string
	redirect_uris []string
}

type authServerInfo struct {
	authEndpoint  string
	tokenEndpoint string
}

type credentials struct {
	AccessToken string
}

type serverError struct {
	StatusCode    int
	StatusMessage string
	ErrorMessage  string
}

var client = clientInfo{
	id:            "goauth",
	secret:        "random-string",
	scope:         "all",
	redirect_uris: []string{serverAddress + "/callback"},
}

var authServer = authServerInfo{
	authEndpoint:  authServerAddress + "/authorize",
	tokenEndpoint: authServerAddress + "/token",
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(rootHandler))
	mux.Handle("/authorize", http.HandlerFunc(authorizeHandler))
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
	})
	mux.HandleFunc("/fetch_resource", http.HandlerFunc(fetchResourceHandler))
	fmt.Println("Listening on " + serverAddress)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func newTemplate(filename string) *template.Template {
	templ, err := template.ParseFiles(filename)
	if err != nil {
		log.Fatalln(err)
	}
	return templ
}

func newCredentials() credentials {
	return credentials{"NULL"}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	indexTempl := newTemplate("index.gohtml")
	creds := newCredentials()
	indexTempl.Execute(w, creds)
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	authorizeURL, err := url.Parse(authServer.authEndpoint)
	if err != nil {
		log.Fatalln(err)
	}

	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("scope", client.scope)
	v.Set("client_id", client.id)
	v.Set("redirect_uri", client.redirect_uris[0])
	authorizeURL.RawQuery = v.Encode()

	http.RedirectHandler(authorizeURL.String(), http.StatusFound).ServeHTTP(w, r)
}

func fetchResourceHandler(w http.ResponseWriter, r *http.Request) {
	indexTempl := newTemplate("index.gohtml")
	errorTempl := newTemplate("error.gohtml")

	accessToken, err := r.Cookie("accessToken")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		whatHappened := serverError{
			StatusCode:    http.StatusUnauthorized,
			StatusMessage: http.StatusText(http.StatusUnauthorized),
			ErrorMessage:  "Missing access token",
		}
		errorTempl.Execute(w, whatHappened)
		return
	}
	// http.Post(protectedResource, )
	creds := credentials{AccessToken: accessToken.Name}
	indexTempl.Execute(w, creds)
}
