package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"time"
)

func initGZBDailyCashUrlMaps() {
	urlMaps = map[string]urlStruct{
		"/gzb/daily-cash": urlStruct{
			handler:      generalPageHander,
			templatePath: "templates/gzb_daily_cash.html",
		},
	}

	for path, urlBlob := range urlMaps {
		templates[path] = template.Must(template.ParseFiles(urlBlob.templatePath))
		http.HandleFunc(path, urlBlob.handler)
	}
}

func initGZBDailyCashApiMaps() {
	apiMaps = map[string]apiStruct{
		"/api/gzbDailyCashEmailApi": apiStruct{
			handler: gzbDailyCashEmailApiHandler,
		},
		"/api/gzbDailyCashOpeningBalanceApi": apiStruct{
			handler: gzbDailyCashGetOpeningBalanceHandler,
		},
		"/api/gzbDailyCashGetUnsettledAdvancesApi": apiStruct{
			handler: gzbDailyCashGetUnsettledAdvancesHandler,
		},
		"/api/gzbDailyCashSettleAccForOneEntryApi": apiStruct{
			handler: gzbDailyCashSettleAccForOneEntryApiHandler,
		},
	}

	for path, apiBlob := range apiMaps {
		http.HandleFunc(path, apiBlob.handler)
	}
}

func init() {
	initGZBDailyCashApiMaps()
	initGZBDailyCashUrlMaps()
	return
}

type GZBUnsettledAdvances struct {
	//TODO: RRKUnsettledAdvances is just like GZBUnsettledAdvance. Fix it if found more clarity.
	Items []CashTransaction
}

func GZBUnsettledAdvancesKey(r *http.Request) *datastore.Key {
	return GZB_SEWNewKey("UnsettledAdvances", "UnsettledAdvances", 0, r)
}

func GZBGetPreviousUnsettledAdvances(r *http.Request) (*GZBUnsettledAdvances, error) {
	c := appengine.NewContext(r)
	k := GZBUnsettledAdvancesKey(r)
	e := new(GZBUnsettledAdvances)
	if err := datastore.Get(c, k, e); err != nil {
		if err == datastore.ErrNoSuchEntity {
			e.Items = []CashTransaction{}
			return e, nil
		}
		return e, err
	}
	return e, nil
}

func GZBCashRollingCounterKey(r *http.Request) *datastore.Key {
	return GZB_SEWNewKey("CashGAERollingCounter", "cashCounter", 0, r)
}

func GZBGetPreviousCashRollingCounter(r *http.Request) (*CashGAERollingCounter, error) {
	c := appengine.NewContext(r)
	k := GZBCashRollingCounterKey(r)
	e := new(CashGAERollingCounter)
	err := datastore.Get(c, k, e)
	return e, err
}

func gzbDailyCashSettleAccForOneEntryApiHandler(w http.ResponseWriter, r *http.Request) {
	var ctx CashTransaction
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prevUnsettledAdv, err := GZBGetPreviousUnsettledAdvances(r)
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
				e := &GZBUnsettledAdvances{Items: append(prevUnsettledAdv.Items[:i], prevUnsettledAdv.Items[i+1:]...)}
				k := GZBUnsettledAdvancesKey(r)
				if _, err1 := datastore.Put(c, k, e); err1 != nil {
					return err1
				}
				//2. Increment the total by same value in datastore
				cashRollingCounter, err1 := GZBGetPreviousCashRollingCounter(r)
				if err1 != nil {
					return err1
				}
				cashRollingCounter.Amount += int64(math.Abs(float64(x.Amount)))

				if _, err1 := datastore.Put(c, GZBCashRollingCounterKey(r), cashRollingCounter); err1 != nil {
					return err1
				}

				emailSubject := fmt.Sprintf("Settling Rs.%d advance given - %s [SEWPULSE][GZBADVSTL]", x.Amount, x.Description)
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

func gzbDailyCashGetUnsettledAdvancesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	prevUnsettledAdv, err := GZBGetPreviousUnsettledAdvances(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(*prevUnsettledAdv)
}

func gzbDailyCashGetOpeningBalanceHandler(w http.ResponseWriter, r *http.Request) {
	type JsonOpeningBal struct {
		Initialized    bool
		OpeningBalance int64
	}

	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	cashRollingCounter, err := GZBGetPreviousCashRollingCounter(r)
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

func gzbSaveUnsettledAdvanceEntryInDataStore(ctx CashTransaction, r *http.Request) error {
	if ctx.Nature != "Unsettled Advance" {
		return errors.New("Trying to enter " + ctx.Nature + " as 'Unsettled Advance' entry. This should not happen.")
	}

	e, err := GZBGetPreviousUnsettledAdvances(r)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			e = &GZBUnsettledAdvances{Items: []CashTransaction{}}
		} else {
			return err
		}
	}

	(*e).Items = append((*e).Items, ctx)
	k := GZBUnsettledAdvancesKey(r)
	c := appengine.NewContext(r)
	if _, err := datastore.Put(c, k, e); err != nil {
		return err
	}

	return nil
}

func gzbDailyCashEmailApiHandler(w http.ResponseWriter, r *http.Request) {
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
				if err1 := gzbSaveUnsettledAdvanceEntryInDataStore(ct, r); err1 != nil {
					return err1
				}
			}
		}

		if _, err1 := datastore.Put(c, GZBCashRollingCounterKey(r), &cashGAERollingCounter); err1 != nil {
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

		emailSubject := fmt.Sprintf("Rs.%v in hand as on %s evening [SEWPULSE][GZBDC]", closingBalance, logDateDDMMYY)
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
