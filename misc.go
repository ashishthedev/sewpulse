package sewpulse

import (
	"appengine"
	"appengine/user"
	"net/http"
	"strings"
)

func myDebug(r *http.Request, s string) {
	c := appengine.NewContext(r)
	c.Debugf(s)
	return
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func IsLocalHostedOrOnDevBranch(r *http.Request) bool {
	if appengine.IsDevAppServer() {
		return true
	}
	if strings.ToLower(r.URL.Host[:4]) == "dev." {
		return true
	}
	return false
}

func IsUserAdminOrDev(r *http.Request) bool {
	if appengine.IsDevAppServer() {
		return true
	}
	c := appengine.NewContext(r)
	if u := user.Current(c); u != nil {
		return u.Admin
	}
	return false
}
