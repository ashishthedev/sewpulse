package sewpulse

import (
	"html/template"
	"net/http"
)

func initGZBUrlMaps() {
	urlMaps = map[string]urlStruct{
		"/gzb": urlStruct{
			handler:      generalPageHander,
			templatePath: "templates/gzb.html",
		},
	}

	for path, urlBlob := range urlMaps {
		templates[path] = template.Must(template.ParseFiles(urlBlob.templatePath))
		http.HandleFunc(path, urlBlob.handler)
	}
	return
}

func initGZBApiMaps() {
}

func init() {
	initGZBApiMaps()
	initGZBUrlMaps()
	return
}
