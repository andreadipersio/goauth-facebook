goauth-facebook
==============

A Facebook Graph authentication library.

[![GoDoc](https://godoc.org/github.com/andreadipersio/goauth-facebook?status.png)](http://godoc.org/github.com/andreadipersio/goauth-facebook)

### Usage

```Go

package main

import (
    "fmt"
    "net/http"

    "github.com/andreadipersio/goauth-facebook/facebook"
)

func main() {
    fbHandler := &facebook.GraphHandler {
        Key: "my app ID/Key",
        Secret: "my app Secret",

        RedirectURI: "http://localhost:8001/oauth/facebook",

        Scope: []string{"email"},

        ErrorCallback: func(w http.ResponseWriter, r *http.Request, err error) {
            http.Error(w, fmt.Sprintf("OAuth error - %v", err), 500)
        },

        SuccessCallback: func(w http.ResponseWriter,  r *http.Request, token *facebook.Token) {
            http.SetCookie(w, &http.Cookie{
                Name: "facebook_token",
                Value: token.Token,
            })
        },
    }

    http.Handle("/oauth/facebook", fbHandler)
    http.ListenAndServe(":8001", nil)
}
```
