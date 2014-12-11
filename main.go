package sewpulse

import (
	"appengine"
	"appengine/user"
	"fmt"
	"net/http"
)

type urlStruct struct {
	handler func(w http.ResponseWriter, r *http.Request)
	path    string
}

var urlMaps map[string]urlStruct

func init() {
	urlMaps = map[string]urlStruct{
		"/": urlStruct{
			handler:      rootHandler,
			path:         "/",
			templatePath: "templates/home.html",
		},

		"/rrk/submit-daily-production": urlStruct{
			handler:      rrkSubmitDailyProductionHandler,
			path:         "/rrk/submit-daily-production",
			templatePath: "/templates/rrk_daily_production.html",
		},
	}

	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
}

func rrkSubmitDailyProductionHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	fmt.Fprintf(w, "<h1>Welcom to rrk, %v</h1>!", u)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	fmt.Fprintf(w, "Hello, %v!", u)
}
