package sewpulse

import (
	"appengine"
	"appengine/mail"
	"appengine/user"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
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
	Quantity  int
	Unit      string
	Remarks   string
}

type JSONValues struct {
	DateTimeAsUTCMilliSeconds int64
	Items                     []ProducedItem
}

var (
	DaysMsg = map[string]string{
		"ON_TIME":     "Submitted on same day",
		"ONE_DAY_OLD": "Submitted after 1 day",
		"X_DAYS_OLD":  "Submitted after %d days",
	}
)

func LogMsgShownForLogTime(logTime time.Time, nowTime time.Time) string {
	l := logTime
	n := nowTime
	newLogTime := time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)
	newNow := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	duration := newNow.Sub(newLogTime)
	hrsDiff := int64(duration.Hours())

	if hrsDiff == 0 {
		return DaysMsg["ON_TIME"]
	} else if hrsDiff == 24 {
		return DaysMsg["ONE_DAY_OLD"]
	} else {
		noOfDays := hrsDiff / 24
		return fmt.Sprintf(DaysMsg["X_DAYS_OLD"], noOfDays)
	}
	panic("Should not reach here")
}
func rrkDailyProdEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var jsonValues JSONValues
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&jsonValues); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	myDebug(r, fmt.Sprintf("%#v", jsonValues))

	logTime := time.Unix(jsonValues.DateTimeAsUTCMilliSeconds/1000, 0)
	logDateYYYYMMMDD := logTime.Format("2006-Jan-02")
	logMsg := LogMsgShownForLogTime(logTime, time.Now())

	htmlTable := fmt.Sprintf(`
	<table border=1 cellpadding=5>
	<caption>
	<u><h1>%s</h1></u>
	<u><h3>%s</h3></u>
	</caption>
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
	`, logDateYYYYMMMDD, logMsg)

	for _, pi := range jsonValues.Items {
		htmlTable +=
			fmt.Sprintf(`
		<tr>
		<td>%s</td>
		<td>%d</td>
		<td>%s</td>
		<td>%s</td>
		</tr>`, pi.ModelName, pi.Quantity, pi.Unit, pi.Remarks)
	}
	htmlTable += "</table>"

	finalHTML := fmt.Sprintf("<html><head></head><body>%s</body></html>", htmlTable)

	bccAddr := Reverse("moc.liamg@dnanatodhsihsa")
	toAddr := Reverse("moc.liamg@ztigihba")

	c := appengine.NewContext(r)
	u := user.Current(c)
	msg := &mail.Message{
		Sender:   u.String() + "<" + u.Email + ">",
		To:       []string{toAddr},
		Bcc:      []string{bccAddr},
		Subject:  fmt.Sprintf("[SEWPULSE][RRKDP] %s", logDateYYYYMMMDD),
		HTMLBody: finalHTML,
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
