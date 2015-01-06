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

func initRRKDailyCashUrlMaps() {
	urlMaps = map[string]urlStruct{
		"/rrk/daily-cash": urlStruct{
			handler:      rrkGeneralPageHander,
			templatePath: "templates/rrk_daily_cash.html",
		},
	}

	for path, urlBlob := range urlMaps {
		templates[path] = template.Must(template.ParseFiles(urlBlob.templatePath))
		http.HandleFunc(path, urlBlob.handler)
	}
}

func initRRKDailyCashApiMaps() {
	apiMaps = map[string]apiStruct{
		"/api/rrkDailyCashEmailApi": apiStruct{
			handler: rrkDailyCashEmailApiHandler,
		},
	}

	for path, apiBlob := range apiMaps {
		http.HandleFunc(path, apiBlob.handler)
	}
}

func init() {
	initRRKDailyCashApiMaps()
	initRRKDailyCashUrlMaps()
	return
}

type CashTransaction struct {
	BillNumber  string
	Amount      int64
	Nature      string
	Description string
}

type cashTxsJSONFormat struct {
	DateOfTransactionAsUTC  int64
	SubmissionDateTimeAsUTC int64
	OpeningBalance          int64
	ClosingBalance          int64
	Items                   []CashTransaction
}

func rrkDailyCashEmailApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var cashTxsAsJson cashTxsJSONFormat
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&cashTxsAsJson); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	submissionTimeAsUTC := cashTxsAsJson.SubmissionDateTimeAsUTC
	logTime := time.Unix(submissionTimeAsUTC/1000, 0)
	logDateDDMMYY := DDMMYYFromUTC(submissionTimeAsUTC)

	logMsg := LogMsgShownForLogTime(logTime, time.Now())
	dateOfTxAsDDMMYY := DDMMYYFromUTC(cashTxsAsJson.DateOfTransactionAsUTC)

	openingBalance := cashTxsAsJson.OpeningBalance
	closingBalance := openingBalance
	for _, ct := range cashTxsAsJson.Items {
		closingBalance += ct.Amount
	}

	if closingBalance != cashTxsAsJson.ClosingBalance {
		http.Error(w, fmt.Sprintf("Application error: Closing Balance is not consistent on client and server. %v != %v", closingBalance, cashTxsAsJson.ClosingBalance), http.StatusInternalServerError)
		return
	}

	htmlTable := fmt.Sprintf(`
	<table border=1 cellpadding=5>
	<caption>
	<u><h1>%s</h1></u>
	<u><h3>%s</h3></u>
	</caption>
	<thead>
	<tr bgcolor=#838468>
	<th><font color='#000000'> Date </font></th>
	<th><font color='#000000'> Nature </font></th>
	<th><font color='#000000'> Amount </font></th>
	<th><font color='#000000'> Bill# </font></th>
	<th><font color='#000000'> Description </font></th>
	</tr>
	</thead>
	<tfoot>
		<tr>
			<td colspan=2>Closing Balance:</td>
			<td colspan=3><font color="#DD472F"><b>%v</b></font></td>
		</tr>
	</tfoot>
	`,
		logDateDDMMYY,
		logMsg,
		closingBalance,
	)

	htmlTable +=
		fmt.Sprintf(`
		<tr>
		<td colspan=2>%v</td>
		<td>%v</td>
		<td></td>
		<td></td>
		</tr>`, "Opening Balance", openingBalance)

	for _, ct := range cashTxsAsJson.Items {
		htmlTable +=
			fmt.Sprintf(`
		<tr>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		</tr>`, dateOfTxAsDDMMYY, ct.Nature, ct.Amount, ct.BillNumber, ct.Description)
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
		Subject:  fmt.Sprintf("Rs.%v in hand as on %s evening [SEWPULSE][RRKDC]", closingBalance, logDateDDMMYY),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
