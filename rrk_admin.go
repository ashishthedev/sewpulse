package sewpulse

import (
	"appengine"
	"appengine/user"
	"html/template"
	"net/http"
)

func initRRKAdminUrlMaps() {
	urlMaps := map[string]urlStruct{
		"/rrk/a": urlStruct{
			handler:      rrkAdminHandler,
			templatePath: "templates/admin/rrk_admin.html",
		},
	}

	for urlPath, urlBlob := range urlMaps {
		templatePath := urlBlob.templatePath
		templates[urlPath] = template.Must(template.ParseFiles(templatePath))
	}

	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
	return
}

func initRRKAdminApiMaps() {
	return
}

func init() {
	initRRKAdminUrlMaps()
	initRRKAdminApiMaps()
	return
}

func rrkAdminHandler(w http.ResponseWriter, r *http.Request) {
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
	// TODO: Check if the user is an admin
	// TODO: Report if it was a failed attempt
	urlPath := r.URL.Path
	myDebug(r, urlPath)
	template := templates[urlPath]
	err := template.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
