package sewpulse

import (
	"appengine"
	"appengine/user"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type urlStruct struct {
	handler func(w http.ResponseWriter, r *http.Request)
	templatePath string
}

type apiStruct struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

var urlMaps map[string]urlStruct
var apiMaps map[string]apiStruct
var templates = make(map[string]*template.Template)

func initUrlMaps() {
	urlMaps = map[string]urlStruct{
		"/": urlStruct{
			handler:      rootHandler,
			templatePath: "templates/home.html",
		},

		"/rrk/daily-production": urlStruct{
			handler:      rrkSubmitDailyProductionHandler,
			templatePath: "templates/rrk_daily_production.html",
		},
	}

	for _, urlBlob := range urlMaps {
		templatePath := urlBlob.templatePath
		if templatePath != ""{
			templates[templatePath] = template.Must(template.ParseFiles(templatePath))
		}
	}
}

func initApiMaps() {
	apiMaps = map[string]apiStruct{
		"/api/rrkDailyProdEmailSendApi": apiStruct{
			handler: rrkDailyProdEmailSendApiHandler,
		},
	}
}

func init() {
	initApiMaps()
	initUrlMaps()
	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
}

func rrkDailyProdEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Println("Inside rrkDailyProdEmailSendApiHandler()")
	return
}
func rrkSubmitDailyProductionHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	template:=templates[urlMaps[urlPath].templatePath]
	template.Execute(w, nil)
	return
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
