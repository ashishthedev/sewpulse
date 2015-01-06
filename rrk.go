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
		"/rrk/daily-polish": urlStruct{
			handler:      rrkGeneralPageHander,
			templatePath: "templates/rrk_daily_polish.html",
		},
		"/rrk/daily-assembly": urlStruct{
			handler:      rrkGeneralPageHander,
			templatePath: "templates/rrk_daily_assembly.html",
		},
		"/rrk": urlStruct{
			handler:      rrkGeneralPageHander,
			templatePath: "templates/rrk.html",
		},
	}

	for path, urlBlob := range urlMaps {
		templates[path] = template.Must(template.ParseFiles(urlBlob.templatePath))
		http.HandleFunc(path, urlBlob.handler)
	}
	return
}

func initRRKApiMaps() {
	apiMaps = map[string]apiStruct{
		"/api/rrkDailyPolishEmailSendApi": apiStruct{
			handler: rrkDailyPolishEmailSendApiHandler,
		},
		"/api/rrkDailyAssemblyEmailSendApi": apiStruct{
			handler: rrkDailyAssemblyEmailSendApiHandler,
		},
	}

	for path, apiBlob := range apiMaps {
		http.HandleFunc(path, apiBlob.handler)
	}
	return
}

func init() {
	initRRKApiMaps()
	initRRKUrlMaps()
	return
}

type ProducedItem struct {
	ModelName        string
	Quantity         int
	Unit             string
	AssemblyLineName string
	Remarks          string
}

type ProducedItemsJSONValues struct {
	DateTimeAsUTCMilliSeconds int64
	Items                     []ProducedItem
}

func rrkDailyAssemblyEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
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

	submissionDateTimeAsUTC := producedItemsAsJson.DateTimeAsUTCMilliSeconds
	logTime := time.Unix(submissionDateTimeAsUTC/1000, 0)
	logDateDDMMYY := DDMMYYFromUTC(submissionDateTimeAsUTC)
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
	<tr bgcolor=#838468>
	<th><font color='#000000'> Product </font></th>
	<th><font color='#000000'> Line </font></th>
	<th><font color='#000000'> Quantity </font></th>
	<th><font color='#000000'> Units </font></th>
	<th><font color='#000000'> Remarks </font></th>
	</tr>
	</thead>
	<tfoot>
		<tr>
			<td colspan=2>Total:</td>
			<td colspan=3><font color="#DD472F"><b>%v</b></font></td>
		</tr>
	</tfoot>
	`,
		logDateDDMMYY,
		logMsg,
		totalQuantityProduced,
	)

	for _, pi := range producedItemsAsJson.Items {
		htmlTable +=
			fmt.Sprintf(`
		<tr>
		<td>%s</td>
		<td>%s</td>
		<td>%d</td>
		<td>%s</td>
		<td>%s</td>
		</tr>`, pi.ModelName, pi.AssemblyLineName, pi.Quantity, pi.Unit, pi.Remarks)
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
		Subject:  fmt.Sprintf("%s: %v pc [SEWPULSE][RRKDA]", logDateDDMMYY, totalQuantityProduced),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func rrkDailyPolishEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
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

	submissionDateTimeAsUTC := producedItemsAsJson.DateTimeAsUTCMilliSeconds
	logTime := time.Unix(submissionDateTimeAsUTC/1000, 0)
	logDateDDMMYY := DDMMYYFromUTC(submissionDateTimeAsUTC)
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
	<tr bgcolor=#838468>
	<th><font color='#000000'> Product </font></th>
	<th><font color='#000000'> Line </font></th>
	<th><font color='#000000'> Quantity </font></th>
	<th><font color='#000000'> Units </font></th>
	<th><font color='#000000'> Remarks </font></th>
	</tr>
	</thead>
	<tfoot>
		<tr>
			<td colspan=2>Total:</td>
			<td colspan=3><font color="#DD472F"><b>%v</b></font></td>
		</tr>
	</tfoot>
	`,
		logDateDDMMYY,
		logMsg,
		totalQuantityProduced,
	)

	for _, pi := range producedItemsAsJson.Items {
		htmlTable +=
			fmt.Sprintf(`
		<tr>
		<td>%s</td>
		<td>%s</td>
		<td>%d</td>
		<td>%s</td>
		<td>%s</td>
		</tr>`, pi.ModelName, pi.AssemblyLineName, pi.Quantity, pi.Unit, pi.Remarks)
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
		Subject:  fmt.Sprintf("%s: %v pc [SEWPULSE][RRKDP]", logDateDDMMYY, totalQuantityProduced),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func rrkGeneralPageHander(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	//TODO: if urlPath ends in / strip it off
	template := templates[urlPath]
	err := template.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
