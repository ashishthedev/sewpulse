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

func initRRKUrlMaps() {
	urlMaps = map[string]urlStruct{
		"/rrk/daily-production": urlStruct{
			handler:      rrkSubmitDailyProductionHandler,
			templatePath: "templates/rrk_daily_production.html",
		},
		"/rrk": urlStruct{
			handler:      rrkHomePageHandler,
			templatePath: "templates/rrk.html",
		},
	}

	for _, urlBlob := range urlMaps {
		//Making sure templatePath exists and caching them
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

func initRRKApiMaps() {
	apiMaps = map[string]apiStruct{
		"/api/rrkDailyProdEmailSendApi": apiStruct{
			handler: rrkDailyProdEmailSendApiHandler,
		},
	}
	return
}

func init() {
	initRRKApiMaps()
	initRRKUrlMaps()
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

type ProducedItemsJSONValues struct {
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

	var producedItemsAsJson ProducedItemsJSONValues
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&producedItemsAsJson); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logTime := time.Unix(producedItemsAsJson.DateTimeAsUTCMilliSeconds/1000, 0)
	logDateYYYYMMMDD := logTime.Format("2006-Jan-02")
	logMsg := LogMsgShownForLogTime(logTime, time.Now())

	totalQuantityProduced := 0
	for _, pi := range producedItemsAsJson.Items {
		totalQuantityProduced += pi.Quantity
	}

	htmlTable := fmt.Sprintf(`
	<table border=1 cellpadding=5>
	<caption>
	<u><h1>%s</h1></u>
	<u><h3>%s</h3></u>
	</caption>
	<thead>
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
	</thead>
	<tfoot>
		<tr>
			<td>Total:</td>
			<td colspan=3><font color="#DD472F"><b>%v</b></font></td>
		</tr>
	</tfoot>
	`,
		logDateYYYYMMMDD,
		logMsg,
		totalQuantityProduced,
	)

	for _, pi := range producedItemsAsJson.Items {
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
	toAddr := ""
	if IsLocalHostedOrOnDevBranch(r) {
		toAddr = Reverse("moc.liamg@dnanatodhsihsa")
	} else {
		toAddr = Reverse("moc.liamg@ztigihba")
	}

	c := appengine.NewContext(r)
	u := user.Current(c)
	msg := &mail.Message{
		Sender:   u.String() + "<" + u.Email + ">",
		To:       []string{toAddr},
		Bcc:      []string{bccAddr},
		Subject:  fmt.Sprintf("%s [SEWPULSE][RRKDP]", logDateYYYYMMMDD),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func rrkHomePageHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	template := templates[urlMaps[urlPath].templatePath]
	err := template.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func rrkSubmitDailyProductionHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	template := templates[urlMaps[urlPath].templatePath]
	err := template.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
