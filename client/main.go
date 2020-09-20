package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	serverAddress     = "http://localhost:9000"
	authServerAddress = "http://localhost:9001"
	protectedResource = "http://localhost:9002"
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
	id:            "oauth-client-1",
	secret:        "oauth-client-secret-1",
	scope:         "foo",
	redirect_uris: []string{serverAddress + "/callback"},
}

var authServer = authServerInfo{
	authEndpoint:  authServerAddress + "/authorize",
	tokenEndpoint: authServerAddress + "/token",
}

func main() {
	rand.Seed(time.Now().UnixNano())
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(rootHandler))
	mux.Handle("/authorize", http.HandlerFunc(authorizeHandler))
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
	})
	mux.HandleFunc("/fetch_resource", http.HandlerFunc(fetchResourceHandler))
	fmt.Println("Listening on " + serverAddress)
	log.Fatal(http.ListenAndServe(":9000", mux))
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

var symbols = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var symbolsLen = len(symbols)

// TODO: Use more advanced cryptography
func randomstring(n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = symbols[rand.Intn(symbolsLen)]
	}
	return string(result)
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
	v.Set("state", randomstring(8))
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
