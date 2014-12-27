package sewpulse

import (
	"appengine"
	"appengine/user"
	"net/http"
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

//A related operation is the need to get the hostname part of a URL to the application. You can use the appengine.DefaultVersionHostname function for this purpose.
