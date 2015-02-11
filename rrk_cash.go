package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/user"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"time"
)

func initRRKDailyCashUrlMaps() {
	urlMaps = map[string]urlStruct{
		"/rrk/daily-cash": urlStruct{
			handler:      generalPageHander,
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
		"/api/rrkDailyCashOpeningBalanceApi": apiStruct{
			handler: rrkDailyCashGetOpeningBalanceHandler,
		},
		"/api/rrkDailyCashGetUnsettledAdvancesApi": apiStruct{
			handler: rrkDailyCashGetUnsettledAdvancesHandler,
		},
		"/api/rrkDailyCashSettleAccForOneEntryApi": apiStruct{
			handler: rrkDailyCashSettleAccForOneEntryApiHandler,
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
	DateUTC     int64
}

type cashTxsJSONFormat struct {
	DateOfTransactionAsUTC  int64
	SubmissionDateTimeAsUTC int64
	OpeningBalance          int64
	ClosingBalance          int64
	Items                   []CashTransaction
}

type CashGAERollingCounter struct {
	Amount                 int64
	DateOfTransactionAsUTC int64
	SetByUserName          string
}

type RRKUnsettledAdvances struct {
	Items []CashTransaction
}

func RRKUnsettledAdvancesKey(r *http.Request) *datastore.Key {
	return SEWNewKey("RRKUnsettledAdvances", "UnsettledAdvances", 0, r)
}

func GetPreviousRRKUnsettledAdvances(r *http.Request) (*RRKUnsettledAdvances, error) {
	c := appengine.NewContext(r)
	k := RRKUnsettledAdvancesKey(r)
	e := new(RRKUnsettledAdvances)
	if err := datastore.Get(c, k, e); err != nil {
		if err == datastore.ErrNoSuchEntity {
			e.Items = []CashTransaction{}
			return e, nil
		}
		return e, err
	}
	return e, nil
}

func CashRollingCounterKey(r *http.Request) *datastore.Key {
	return SEWNewKey("CashGAERollingCounter", "cashCounter", 0, r)
}

func GetPreviousCashRollingCounter(r *http.Request) (*CashGAERollingCounter, error) {
	c := appengine.NewContext(r)
	k := CashRollingCounterKey(r)
	e := new(CashGAERollingCounter)
	err := datastore.Get(c, k, e)
	return e, err
}

func rrkDailyCashSettleAccForOneEntryApiHandler(w http.ResponseWriter, r *http.Request) {
	var ctx CashTransaction
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prevUnsettledAdv, err := GetPreviousRRKUnsettledAdvances(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, x := range prevUnsettledAdv.Items {
		if math.Abs(float64(x.Amount)) == math.Abs(float64(ctx.Amount)) && x.Description == ctx.Description && x.BillNumber == ctx.BillNumber && x.DateUTC == ctx.DateUTC {
			//Found the entry
			c := appengine.NewContext(r)
			err := datastore.RunInTransaction(c, func(c appengine.Context) error {
				var err1 error
				//1. Delete this entry
				e := &RRKUnsettledAdvances{Items: append(prevUnsettledAdv.Items[:i], prevUnsettledAdv.Items[i+1:]...)}
				k := RRKUnsettledAdvancesKey(r)
				if _, err1 := datastore.Put(c, k, e); err1 != nil {
					return err1
				}
				//2. Increment the total by same value in datastore
				cashRollingCounter, err1 := GetPreviousCashRollingCounter(r)
				if err1 != nil {
					return err1
				}
				cashRollingCounter.Amount += int64(math.Abs(float64(x.Amount)))
				cashGAERollingCounterKey := CashRollingCounterKey(r)

				if _, err1 := datastore.Put(c, cashGAERollingCounterKey, cashRollingCounter); err1 != nil {
					return err1
				}

				emailSubject := fmt.Sprintf("Settling Rs.%d advance given - %s [SEWPULSE][RRKADVSTL]", x.Amount, x.Description)
				htmlBody := fmt.Sprintf("Settling Rs.%d advance given - %s", x.Amount, x.Description)
				if err1 := SendMailToDesignatedPeopleNow(r, emailSubject, htmlBody); err1 != nil {
					return err1
				}

				return nil
			}, nil)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
	}
}

func rrkDailyCashGetUnsettledAdvancesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	prevUnsettledAdv, err := GetPreviousRRKUnsettledAdvances(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(*prevUnsettledAdv)
}

func rrkDailyCashGetOpeningBalanceHandler(w http.ResponseWriter, r *http.Request) {
	type JsonOpeningBal struct {
		Initialized    bool
		OpeningBalance int64
	}

	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	cashRollingCounter, err := GetPreviousCashRollingCounter(r)
	if err == datastore.ErrNoSuchEntity {
		json.NewEncoder(w).Encode(JsonOpeningBal{Initialized: false, OpeningBalance: 0})
		return

	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(JsonOpeningBal{Initialized: true, OpeningBalance: cashRollingCounter.Amount})
}

func SendMailNow(r *http.Request, toAddr []string, ccAddr []string, bccAddr []string, emailSubject string, htmlBody string) error {

	c := appengine.NewContext(r)
	u := user.Current(c)
	msg := &mail.Message{
		Sender:   u.String() + "<" + u.Email + ">",
		To:       toAddr,
		Bcc:      bccAddr,
		Subject:  emailSubject,
		HTMLBody: htmlBody,
	}
	return mail.Send(c, msg)
}

func SendMailToDesignatedPeopleNow(r *http.Request, emailSubject string, htmlBody string) error {
	bccAddr := []string{}
	bccAddr = append(bccAddr, Reverse("moc.liamg@dnanatodhsihsa"))
	toAddr := []string{}
	if IsLocalHostedOrOnDevBranch(r) {
		toAddr = append(toAddr, Reverse("moc.liamg@dnanatodhsihsa"))
	} else {
		toAddr = append(toAddr, Reverse("moc.liamg@ztigihba"))
	}
	var ccAddr []string = nil

	return SendMailNow(r, toAddr, ccAddr, bccAddr, emailSubject, htmlBody)

}

func rrkSaveUnsettledAdvanceEntryInDataStore(ctx CashTransaction, r *http.Request) error {
	if ctx.Nature != "Unsettled Advance" {
		return errors.New("Trying to enter " + ctx.Nature + " as 'Unsettled Advance' entry. This should not happen.")
	}

	e, err := GetPreviousRRKUnsettledAdvances(r)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			e = &RRKUnsettledAdvances{Items: []CashTransaction{}}
		} else {
			return err
		}
	}

	(*e).Items = append((*e).Items, ctx)
	k := RRKUnsettledAdvancesKey(r)
	c := appengine.NewContext(r)
	if _, err := datastore.Put(c, k, e); err != nil {
		return err
	}

	return nil
}

func rrkDailyCashEmailApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var cashTxsAsJson cashTxsJSONFormat
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&cashTxsAsJson); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	submissionTimeAsUTC := cashTxsAsJson.SubmissionDateTimeAsUTC
	logDateDDMMYY := DDMMYYFromUTC(submissionTimeAsUTC)

	logTime := time.Unix(submissionTimeAsUTC/1000, 0)
	logMsg := LogMsgShownForLogTime(logTime, time.Now())

	dateOfTxAsDDMMYY := DDMMYYFromUTC(cashTxsAsJson.DateOfTransactionAsUTC)

	openingBalance := cashTxsAsJson.OpeningBalance //TODO: Do not read cash opening balance from jSon. Read from server.
	closingBalance := openingBalance
	for _, ct := range cashTxsAsJson.Items {
		closingBalance += ct.Amount
	}

	if closingBalance != cashTxsAsJson.ClosingBalance {
		http.Error(w, fmt.Sprintf("Application error: Closing Balance is not consistent on client and server. %v != %v", closingBalance, cashTxsAsJson.ClosingBalance), http.StatusInternalServerError)
		return
	}

	c := appengine.NewContext(r)
	cashGAERollingCounter := CashGAERollingCounter{
		Amount:                 closingBalance,
		DateOfTransactionAsUTC: cashTxsAsJson.DateOfTransactionAsUTC,
		SetByUserName:          user.Current(c).String(),
	}

	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		for _, ct := range cashTxsAsJson.Items {
			//Save any unsettled amount in the datastore
			if ct.Nature == "Unsettled Advance" {
				if err1 := rrkSaveUnsettledAdvanceEntryInDataStore(ct, r); err1 != nil {
					return err1
				}
			}
		}

		if _, err1 := datastore.Put(c, CashRollingCounterKey(r), &cashGAERollingCounter); err1 != nil {
			return err1
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

		emailSubject := fmt.Sprintf("Rs.%v in hand as on %s evening [SEWPULSE][RRKDC]", closingBalance, logDateDDMMYY)
		if err1 := SendMailToDesignatedPeopleNow(r, emailSubject, finalHTML); err1 != nil {
			return err1
		}
		return nil
	}, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
