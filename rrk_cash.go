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
		"/api/rrkCashBookStoreAndEmailApi": apiStruct{
			handler: rrkCashBookStoreAndEmailApiHandler,
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
		"/rrk/update": apiStruct{
			handler: rrkDailyCashUpdateModelApiHandler,
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
	BillNumber     string
	Amount         int64
	Nature         string
	Description    string
	DateAsUnixTime int64
}

type CashTxsCluster struct {
	DateOfTransactionAsUnixTime int64
	OpeningBalance              int64
	ClosingBalance              int64
	Items                       []CashTransaction
}

type CashGAERollingCounter struct {
	Amount                      int64
	DateOfTransactionAsUnixTime int64
	SetByUserName               string
}

type RRKUnsettledAdvances struct {
	Items []CashTransaction
}

func RRKUnsettledAdvancesKey(r *http.Request) *datastore.Key {
	return RRK_SEWNewKey("RRKUnsettledAdvances", "UnsettledAdvances", 0, r)
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

func RRKCashRollingCounterKey(r *http.Request) *datastore.Key {
	return RRK_SEWNewKey("CashGAERollingCounter", "cashCounter", 0, r)
}

func RRKGetPreviousCashRollingCounter(r *http.Request) (*CashGAERollingCounter, error) {
	c := appengine.NewContext(r)
	k := RRKCashRollingCounterKey(r)
	e := new(CashGAERollingCounter)
	err := datastore.Get(c, k, e)
	return e, err
}

func rrkDailyCashUpdateModelApiHandler(w http.ResponseWriter, r *http.Request) {
	type oldCashGAERollingCounter struct {
		Amount                 int64
		DateOfTransactionAsUTC int64
		SetByUserName          string
	}

	c := appengine.NewContext(r)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		cashk := RRKCashRollingCounterKey(r)
		oldE := new(oldCashGAERollingCounter)

		if err := datastore.Get(c, cashk, oldE); err != nil {
			return err
		}

		e := new(CashGAERollingCounter)
		e.Amount = oldE.Amount
		e.DateOfTransactionAsUnixTime = oldE.DateOfTransactionAsUTC
		e.SetByUserName = oldE.SetByUserName

		if _, err := datastore.Put(c, cashk, e); err != nil {
			return err
		}

		type OldCashTransaction struct {
			BillNumber      string
			Amount          int64
			RealAmountSpent int64
			Nature          string
			Description     string
			DateUTC         int64
		}
		type OldRRKUnsettledAdvances struct {
			Items []OldCashTransaction
		}

		unsAdvk := RRKUnsettledAdvancesKey(r)
		olde, err := func(r *http.Request) (*OldRRKUnsettledAdvances, error) {
			olde := new(OldRRKUnsettledAdvances)
			if err := datastore.Get(c, unsAdvk, olde); err != nil {
				if err == datastore.ErrNoSuchEntity {
					olde.Items = []OldCashTransaction{}
					return olde, nil
				}
				return olde, err
			}
			return olde, nil
		}(r)

		if err != nil {
			return err
		}

		newe := new(RRKUnsettledAdvances)
		for _, item := range olde.Items {
			newe.Items = append(newe.Items, CashTransaction{
				BillNumber:     item.BillNumber,
				Amount:         item.Amount,
				Nature:         item.Nature,
				Description:    item.Description,
				DateAsUnixTime: item.DateUTC / 1000,
			})
		}
		if _, err := datastore.Put(c, unsAdvk, newe); err != nil {
			return err
		}
		return nil

	}, nil)

	if err == nil {
		fmt.Fprintf(w, "Updated")
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return

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
		if math.Abs(float64(x.Amount)) == math.Abs(float64(ctx.Amount)) && x.Description == ctx.Description && x.BillNumber == ctx.BillNumber && x.DateAsUnixTime == ctx.DateAsUnixTime {
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
				cashRollingCounter, err1 := RRKGetPreviousCashRollingCounter(r)
				if err1 != nil {
					return err1
				}
				cashRollingCounter.Amount += int64(math.Abs(float64(x.Amount)))
				cashGAERollingCounterKey := RRKCashRollingCounterKey(r)

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

	cashRollingCounter, err := RRKGetPreviousCashRollingCounter(r)
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

func rrkCashBookStoreAndEmailApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var cashTxs CashTxsCluster
	if err := json.NewDecoder(r.Body).Decode(&cashTxs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := CashBookStoreAndEmailApi(&cashTxs, r, RRKCashRollingCounterKey, RRKGetPreviousCashRollingCounter, rrkSaveUnsettledAdvanceEntryInDataStore, "RRKDC"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
