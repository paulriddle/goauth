package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
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
	State        string
	AccessToken  string
	RefreshToken string
	Scope        string
}

type httpError struct {
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

var Credentials credentials

func main() {
	rand.Seed(time.Now().UnixNano())
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(rootHandler))
	mux.Handle("/authorize", http.HandlerFunc(authorizeHandler))
	mux.HandleFunc("/callback", http.HandlerFunc(callbackHandler))
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
	indexTempl.Execute(w, Credentials)
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
	Credentials.State = randomstring(8)
	v.Set("state", Credentials.State)
	authorizeURL.RawQuery = v.Encode()

	http.RedirectHandler(authorizeURL.String(), http.StatusFound).ServeHTTP(w, r)
}

func fetchResourceHandler(w http.ResponseWriter, r *http.Request) {
	indexTempl := newTemplate("index.gohtml")
	errorTempl := newTemplate("error.gohtml")

	if Credentials.AccessToken == "" {
		renderError(w, http.StatusForbidden, "Missing access token")
	}

	accessToken, err := r.Cookie("accessToken")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		whatHappened := httpError{
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

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	recievedState := r.URL.Query().Get("state")
	if recievedState != Credentials.State {
		errMsg := fmt.Sprintf("State does not match: expected %s got %s",
			Credentials.State, recievedState)
		renderError(w, http.StatusForbidden, errMsg)
		return
	}

	code := r.URL.Query().Get("code")
	params := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {client.redirect_uris[0]},
	}
	payload := strings.NewReader(params.Encode())
	req, err := http.NewRequest("POST", authServer.tokenEndpoint, payload)
	if err != nil {
		errMsg := fmt.Sprintf(`Failure: http.NewRequest("POST", %s, %+v)\n%+v`,
			authServer.tokenEndpoint, payload, err)
		renderError(w, http.StatusInternalServerError, errMsg)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(url.QueryEscape(client.id), url.QueryEscape(client.secret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Failure: http.DefaultClient.Do(%+v)\n%+v", req, err)
		renderError(w, http.StatusInternalServerError, errMsg)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var accessTokenData struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			Scope        string
		}
		if err := json.NewDecoder(resp.Body).Decode(&accessTokenData); err != nil {
			errMsg := fmt.Sprintf("Error parsing JSON: %+v", err)
			renderError(w, http.StatusInternalServerError, errMsg)
			return
		}

		Credentials.AccessToken = accessTokenData.AccessToken
		Credentials.RefreshToken = accessTokenData.RefreshToken
		Credentials.Scope = accessTokenData.Scope

		indexTempl := newTemplate("index.gohtml")
		indexTempl.Execute(w, Credentials)
	} else {
		renderError(w, resp.StatusCode, "Error requesting access token")
	}
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
