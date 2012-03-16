// Copyright 2011 The goauth2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program makes a call to the buzz API, authenticated with OAuth2.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"code.google.com/p/goauth2/oauth"
)

var (
	code         = flag.String("code", "", "Authorization Code")
	token        = flag.String("token", "", "Access Token")
	clientId     = flag.String("id", "", "Client ID")
	clientSecret = flag.String("secret", "", "Client Secret")
)

const usageMsg = `
You must specify at least -id and -secret.
To obtain these details, see the "OAuth 2 Credentials" section under
the "API Access" tab on this page: https://code.google.com/apis/console/
`

const activities = "https://www.googleapis.com/buzz/v1/activities/@me/@public?max-results=1&alt=json"

func main() {
	flag.Parse()
	if *clientId == "" || *clientSecret == "" {
		flag.Usage()
		fmt.Fprint(os.Stderr, usageMsg)
		return
	}

	// Set up a configuration
	config := &oauth.Config{
		ClientId:     *clientId,
		ClientSecret: *clientSecret,
		Scope:        "https://www.googleapis.com/auth/buzz",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		RedirectURL:  "http://localhost/",
	}

	// Step one, get an authorization code from the data provider.
	// ("Please ask the user if I can access this resource.")
	if *code == "" && *token == "" {
		url := config.AuthCodeURL("")
		fmt.Println("Visit this URL to get a code, then run again with -code=YOUR_CODE")
		fmt.Println(url)
		return
	}

	// Set up a Transport with our config.
	t := &oauth.Transport{Config: config}

	// Step two, exchange the authorization code for an access token.
	// ("Here's the code you gave the user, now give me a token!")
	if *token == "" {
		tok, err := t.Exchange(*code)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Now run again with -token=%s\n", tok.AccessToken)
		return
		// We needn't return here; we could just use the Transport
		// to make authenticated requests straight away.
		// The process has been split up to demonstrate how one might
		// restore Credentials that have been previously stored.
	}

	// Step three, make the actual request using the token to authenticate.
	// ("Here's the token, let me in!")
	// First, re-instate our Token (typically this would be stored on disk).
	t.Token = &oauth.Token{
		AccessToken: *token,
		// If you were storing this information somewhere,
		// you'd want to store the RefreshToken field as well.
	}
	// Make the request.
	r, err := t.Client().Get(activities)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	// Write the response to standard output.
	io.Copy(os.Stdout, r.Body)
	fmt.Println()
}
