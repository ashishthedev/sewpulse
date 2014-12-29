package sewpulse

import (
	"appengine"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
)

type urlStruct struct {
	handler      func(w http.ResponseWriter, r *http.Request)
	templatePath string
}

type apiStruct struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

var urlMaps map[string]urlStruct
var apiMaps map[string]apiStruct
var templates = make(map[string]*template.Template)

func initRootUrlMaps() {
	urlMaps = map[string]urlStruct{
		"/": urlStruct{
			handler:      rootHandler,
			templatePath: "templates/home.html",
		},
	}

	for _, urlBlob := range urlMaps {
		templatePath := urlBlob.templatePath
		if templatePath != "" {
			templates[templatePath] = template.Must(template.ParseFiles(templatePath))
		}
	}

	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
	return
}

func initRootApiMaps() {
	return
}

func init() {
	initRootApiMaps()
	initRootUrlMaps()
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

	return
}
