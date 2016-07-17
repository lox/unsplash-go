package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"golang.org/x/oauth2"
)

var ()

var (
	serverBind = "localhost:9900"
	oAuthConf  = &oauth2.Config{
		ClientID:     "0b82b42c05108a5c98558dce5674f4effec7734476049ae3204b8b4a40672143",
		ClientSecret: "0011c75ddaea36324b9035f4807f51168d714c455aba2305793d5b31d9053213",
		Scopes:       []string{"public", "read_photos", "read_collections"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://unsplash.com/oauth/authorize",
			TokenURL: "https://unsplash.com/oauth/token",
		},
		RedirectURL: "http://" + serverBind,
	}
	oAuthStateString string
)

func init() {
	oAuthStateString = fmt.Sprintf("state-%d", time.Now().UnixNano())
}

func newOAuthClient() (*http.Client, error) {
	var token *oauth2.Token

	if isLoggedIn() {
		var err error
		token, err = loadToken()
		if err != nil {
			return nil, err
		}
	} else {
		ln, err := net.Listen("tcp", serverBind)
		if err != nil {
			return nil, err
		}

		tokenCh := make(chan *oauth2.Token)
		go http.Serve(ln, unsplashHandler(tokenCh))

		openURL(oAuthConf.AuthCodeURL(oAuthStateString, oauth2.AccessTypeOnline))

		token = <-tokenCh
		ln.Close()

		if err = saveToken(token); err != nil {
			return nil, err
		}
	}

	return oAuthConf.Client(oauth2.NoContext, token), nil
}

func unsplashHandler(tokenCh chan *oauth2.Token) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		if state != oAuthStateString {
			http.Error(w, "Invalid oauth state", http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")
		token, err := oAuthConf.Exchange(oauth2.NoContext, code)
		if err != nil {
			http.Error(w, "OAuth exchange failed with "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Authorization completed, close browser and return to console")
		tokenCh <- token
	})
}

func isLoggedIn() bool {
	t, err := loadToken()
	if err != nil || time.Now().After(t.Expiry) {
		return false
	}
	return true
}

func loadToken() (*oauth2.Token, error) {
	tokenFile, err := os.Open("token.json")
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	defer tokenFile.Close()

	err = json.NewDecoder(tokenFile).Decode(token)
	if err != nil {
		return nil, err
	}
	return token, err
}

func saveToken(token *oauth2.Token) error {
	tokenFile, err := os.Create("token.json")
	if err != nil {
		return err
	}

	defer tokenFile.Close()
	return json.NewEncoder(tokenFile).Encode(token)
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Fatal("Error opening URL in browser.")
}
