// Package facebook implement OAuth2 authentication for Facebook Graph
// providing an handler (GraphHandler) which perform OAuth token authorization
// and exchange.
package facebook

import (
    "fmt"

    "strconv"
    "strings"

    "time"

    "net/url"
    "net/http"

    "io/ioutil"
)

const (
    authorizationURL   = "https://www.facebook.com/dialog/oauth?"
    tokenExchangeURL = "https://graph.facebook.com/oauth/access_token?"
)

type Token struct {
    Token string
    ExpireAt time.Time
}

type GraphHandler struct {
    // App Key/ID
    Key,

    // App Secret
    Secret,

    // Login URL
    RedirectURI string

    // https://developers.facebook.com/docs/reference/login/
    Scope []string

    // SuccessCallback is executed when TokenExchange succeed
    SuccessCallback func(http.ResponseWriter, *http.Request, *Token)

    // ErrorCallback is executed when any of the OAuth step fails
    ErrorCallback func(http.ResponseWriter, *http.Request, error)
}

func (h *GraphHandler) AuthorizeRedirectUrl() string {
    qs := url.Values{}

    qs.Set("redirect_uri", h.RedirectURI)
    qs.Set("client_id", h.Key)
    qs.Set("scope", strings.Join(h.Scope, ","))

    return fmt.Sprintf("%v%v", authorizationURL, qs.Encode())
}

func (h *GraphHandler) tokenExchangeUrl(authcode string) string {
    qs := url.Values{}

    qs.Set("redirect_uri", h.RedirectURI)
    qs.Set("client_id", h.Key)
    qs.Set("client_secret", h.Secret)
    qs.Set("code", authcode)

    return fmt.Sprintf("%v%v", tokenExchangeURL, qs.Encode())
}

func (h *GraphHandler) TokenExchange(authcode string) (string, time.Time, error) {
    tokenURL := h.tokenExchangeUrl(authcode)

    bailout := func(err error) (string, time.Time, error) {
        return "", time.Now(), fmt.Errorf("%v: %v", tokenURL, err)
    }

    resp, err := http.Get(tokenURL)
    if err != nil {
        return bailout(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return bailout(err)
    }

    qs, err := url.ParseQuery(string(body))
    if err != nil {
        return bailout(err)
    }

    expireInSecs, err := strconv.Atoi(qs.Get("expires"))
    if err != nil {
        return bailout(err)
    }

    expireAt := time.Unix(int64(expireInSecs) + time.Now().Unix(), 0)

    return qs.Get("access_token"), expireAt, nil
}

// If no auth code is found, then redirect to facebook authorization endpoint,
// otherwise try to exchange the auth code with a bearer token, by invoking
// GraphHandler.readToken.
// On success token is passed to GraphHandler.SuccessCallback,
// otherwise error is passed to GraphHandler.ErrorCallback
// (error is a string - error_code: error_description).
func (h *GraphHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    qs := r.URL.Query()

    code := qs.Get("code")

    if code == "" {
        http.Redirect(w, r, h.AuthorizeRedirectUrl(), 302)
        return
    }

    var err error

    token := &Token{}

    if token.Token, token.ExpireAt, err = h.TokenExchange(code); err != nil {
        h.ErrorCallback(w, r, err)
        return
    } else {
        h.SuccessCallback(w, r, token)
    }
}
