package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"net/http"
	"strconv"
	"time"
)

func rrkDeleteEverythingBeforeAndIncludingDateSlashApiHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		unixTimeAsString := r.URL.Path[len(API_ADMIN_MASS_DELETION_BEFORE_AND_INCLUDING_DATE_SLASH_END):]
		if unixTimeAsString == "" {
			http.Error(w, "Date argument is empty", http.StatusBadRequest)
			return
		}
		unixTime, err := strconv.ParseInt(unixTimeAsString, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		DD_MMM_YY := DDMMMYYFromUnixTime(unixTime)
		dateValue, err := DDMMMYYToGoTime(DD_MMM_YY)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c := appengine.NewContext(r)
		err = datastore.RunInTransaction(c, func(c appengine.Context) error {
			if err1 := rrkDeleteEverythingBeforeAndIncludingDate(r, DD_MMM_YY); err1 != nil {
				return err1
			}
			return RRKIntelligentlySetDD(r, dateValue)
		}, nil)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func rrkDeleteEverythingBeforeAndIncludingDate(r *http.Request, DD_MMM_YY string) error {

	if err := RRKDigestAsMuchStockAsPossibleSinceDirtyDate(r); err != nil {
		return err
	}

	dateValue, err := DDMMMYYToGoTime(DD_MMM_YY)
	if err != nil {
		return err
	}
	dateValue = EOD(dateValue)
	keyRetrieverFuncs := []func(*http.Request, time.Time) ([]*datastore.Key, error){
		RRKGetAllAssembledItemsBeforeThisDateInclusiveKeysOnly,
		RRKGetAllSaleInvoicesBeforeThisDateInclusiveKeysOnly,
		RRKGetAllSTockPosAsStringBeforeThisDateInclusiveKeysOnly,
		RRKGetAllPurchaseItemsBeforeThisDateInclusiveKeysOnly,
		RRKGetAllFPOSTInvoicesBeforeThisDateInclusiveKeysOnly,
		RRKGetAllRMPOSTInvoicesBeforeThisDateInclusiveKeysOnly,
		RRKGetAllFPISTInvoicesBeforeThisDateInclusiveKeysOnly,
		RRKGetAllRMISTInvoicesBeforeThisDateInclusiveKeysOnly,
		RRKGetAllFPAAInvoicesBeforeThisDateInclusiveKeysOnly,
		RRKGetAllRMAAInvoicesBeforeThisDateInclusiveKeysOnly,
	}

	var keysToBeDeleted [][]*datastore.Key
	for _, fn := range keyRetrieverFuncs {
		keys, err := fn(r, dateValue)
		if err != nil {
			return err
		}
		keysToBeDeleted = append(keysToBeDeleted, keys)
	}

	c := appengine.NewContext(r)
	err1 := datastore.RunInTransaction(c, func(c appengine.Context) error {
		for _, keyGroup := range keysToBeDeleted {
			if err := datastore.DeleteMulti(appengine.NewContext(r), keyGroup); err != nil {
				return err
			}
		}
		return nil
	}, nil)
	if err1 != nil {
		return err1
	}
	return nil

}
