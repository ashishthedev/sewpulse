package sewpulse

import (
	"appengine"
	"appengine/mail"
	"appengine/user"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"io"
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
		if templatePath != "" {
			templates[templatePath] = template.Must(template.ParseFiles(templatePath))
		}
	}
	return
}

func initApiMaps() {
	apiMaps = map[string]apiStruct{
		"/api/rrkDailyProdEmailSendApi": apiStruct{
			handler: rrkDailyProdEmailSendApiHandler,
		},
	}
	return
}

func init() {
	initApiMaps()
	initUrlMaps()
	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
	for path, apiBlob := range apiMaps {
		http.HandleFunc(path, apiBlob.handler)
	}
	return
}
type ProducedItem struct {
	ModelName string
	Quantity int 
	Unit string
	Remarks string
}

func (x *ProducedItem) String() string {
	return fmt.Sprintf("%s %d%s %s", x.ModelName, x.Quantity, x.Unit, x.Remarks)
}

func rrkDailyProdEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}


	var pis []ProducedItem
	dec := json.NewDecoder(r.Body)
	for {
		if err := dec.Decode(&pis); err == io.EOF {
			break
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		myDebug(r, fmt.Sprintf("%#v", pis))
	}

	finalHTML :=`
	<table border=1 cellpadding=5>
	<tr bgcolor=#838468> <th>
	<font color='#000000'> Product </font>
	</th>
	<th>
	<font color='#000000'> Quantity </font>
	</th>
	<th>
	<font color='#000000'> Units </font>
	</th>
	<th>
	<font color='#000000'> Remarks </font>
	</th> </tr> 
	`

	for _, pi:= range pis{
		finalHTML +=
		fmt.Sprintf(`
		<tr>
		<td>%s</td>
		<td>%s</td>
		<td>%s</td>
		<td>%s</td>
		</tr>`, pi.ModelName, pi.Quantity, pi.Unit, pi.Remarks)
	}
	finalHTML += "</table>"

	finalHTML=fmt.Sprintf("<html><head></head><body>%s</body></html>", finalHTML)

	bccAddr := Reverse("moc.liamg@dnanatodhsihsa")
	toAddr := bccAddr //toAddr := Reverse("moc.liamg@ztigihba")

	c := appengine.NewContext(r)
	u := user.Current(c)
	msg := &mail.Message{
		Sender:  u.String() + "<" + u.Email + ">",
		To:      []string{toAddr},
		Bcc:     []string{bccAddr},
		Subject: "[SEWPULSE-RRK-DP]",
		HTMLBody:    finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func rrkSubmitDailyProductionHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	template := templates[urlMaps[urlPath].templatePath]
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
	return
}
