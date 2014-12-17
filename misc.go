package sewpulse

import (
	"appengine"
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

