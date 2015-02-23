package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"net/http"
	"time"
)

type ZoneCIHKeyFunc func(r *http.Request) *datastore.Key
type ZoneGAECashFunc func(r *http.Request) (*CashGAERollingCounter, error)
type ZoneSaveUnsettledAdvInDataStoreFunc func(ctx CashTransaction, r *http.Request) error

func HTMLforDailyCash(cashTxs *CashTxsCluster, closingBalance int64, openingBalance int64, xxxEmailTag string) (emailSubject string, html string, err error) {

	submissionTimeAsUnixTime := cashTxs.DateOfTransactionAsUnixTime
	logDateDDMMYY := DDMMYYFromUnixTime(submissionTimeAsUnixTime)

	logTime := time.Unix(submissionTimeAsUnixTime, 0)
	logMsg := LogMsgShownForLogTime(logTime, time.Now())

	html = fmt.Sprintf(`
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

	html += fmt.Sprintf(`
			<tr>
			<td colspan=2>%v</td>
			<td>%v</td>
			<td></td>
			<td></td>
			</tr>`, "Opening Balance", openingBalance)

	dateOfTxAsDDMMYY := DDMMYYFromUnixTime(cashTxs.DateOfTransactionAsUnixTime)

	for _, ct := range cashTxs.Items {
		html += fmt.Sprintf(`
		<tr>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		</tr>`, dateOfTxAsDDMMYY, ct.Nature, ct.Amount, ct.BillNumber, ct.Description)
	}

	html += "</table>"

	html = fmt.Sprintf("<html><head></head><body>%s</body></html>", html)
	emailSubject = fmt.Sprintf("Rs.%v in hand as on %s evening [SEWPULSE][%s]", closingBalance, logDateDDMMYY, xxxEmailTag)
	return

}

func FindAndSaveUnsettledAdvInDataStore(r *http.Request, cashTxs *CashTxsCluster, zoneSaveUnsettledAdvInDataStoreFunc ZoneSaveUnsettledAdvInDataStoreFunc) error {
	for _, ct := range cashTxs.Items {
		//Save any unsettled amount in the datastore
		if ct.Nature == "Unsettled Advance" {
			if err1 := zoneSaveUnsettledAdvInDataStoreFunc(ct, r); err1 != nil {
				return err1
			}
		}
	}
	return nil
}

func CashBookStoreAndEmailApi(cashTxs *CashTxsCluster, r *http.Request, zoneCIHKeyFunc ZoneCIHKeyFunc, zoneGetGAECashGFunc ZoneGAECashFunc, zoneSaveUnsettledAdvInDataStoreFunc ZoneSaveUnsettledAdvInDataStoreFunc, xxxEmailTag string) error {

	c := appengine.NewContext(r)

	cashRollingCounter, err := zoneGetGAECashGFunc(r)
	var openingBalance int64
	if err != nil && err != datastore.ErrNoSuchEntity {
		return err
	} else if err == datastore.ErrNoSuchEntity {
		//If no such entity then cash is 0 as the zone has just started out
		openingBalance = cashTxs.OpeningBalance
	} else {
		openingBalance = cashRollingCounter.Amount
	}

	closingBalance := openingBalance
	for _, ct := range cashTxs.Items {
		closingBalance += ct.Amount
	}

	//if closingBalance != cashTxs.ClosingBalance {
	//	return errors.New(fmt.Sprintf("Application error: Closing Balance is not consistent on client and server. %v != %v", closingBalance, cashTxs.ClosingBalance))
	//}

	cashGAERollingCounter := CashGAERollingCounter{
		Amount: closingBalance,
		DateOfTransactionAsUnixTime: cashTxs.DateOfTransactionAsUnixTime,
		SetByUserName:               user.Current(c).String(),
	}

	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		if err1 := FindAndSaveUnsettledAdvInDataStore(r, cashTxs, zoneSaveUnsettledAdvInDataStoreFunc); err1 != nil {
			return err1
		}

		if _, err1 := datastore.Put(c, zoneCIHKeyFunc(r), &cashGAERollingCounter); err1 != nil {
			return err1
		}

		emailSubject, finalHTML, err1 := HTMLforDailyCash(cashTxs, closingBalance, openingBalance, xxxEmailTag)
		if err1 != nil {
			return err1
		}
		if err1 := SendMailToDesignatedPeopleNow(r, emailSubject, finalHTML); err1 != nil {
			return err1
		}
		return nil
	}, nil)

	if err != nil {
		return err
	}

	return nil
}
