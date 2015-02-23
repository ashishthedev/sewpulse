package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
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
		"/api/gzbCashBookStoreAndEmailApi": apiStruct{
			handler: gzbCashBookStoreAndEmailApiHandler,
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
		"/gzb/update": apiStruct{
			handler: gzbDailyCashUpdateModelApiHandler,
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

func gzbDailyCashUpdateModelApiHandler(w http.ResponseWriter, r *http.Request) {
	type oldCashGAERollingCounter struct {
		Amount                 int64
		DateOfTransactionAsUTC int64
		SetByUserName          string
	}

	c := appengine.NewContext(r)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		cashk := GZBCashRollingCounterKey(r)
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
			BillNumber  string
			Amount      int64
			Nature      string
			Description string
			DateUTC     int64
		}
		type OldGZBUnsettledAdvances struct {
			Items []OldCashTransaction
		}

		unsAdvk := GZBUnsettledAdvancesKey(r)
		olde, err := func(r *http.Request) (*OldGZBUnsettledAdvances, error) {
			olde := new(OldGZBUnsettledAdvances)
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

		newe := new(GZBUnsettledAdvances)
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
		if math.Abs(float64(x.Amount)) == math.Abs(float64(ctx.Amount)) && x.Description == ctx.Description && x.BillNumber == ctx.BillNumber && x.DateAsUnixTime == ctx.DateAsUnixTime {
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

func gzbCashBookStoreAndEmailApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var cashTxs CashTxsCluster
	if err := json.NewDecoder(r.Body).Decode(&cashTxs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := CashBookStoreAndEmailApi(&cashTxs, r, GZBCashRollingCounterKey, GZBGetPreviousCashRollingCounter, gzbSaveUnsettledAdvanceEntryInDataStore, "GZBDC"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
