package sewpulse

//import "fmt"
import (
	"appengine"
	"appengine/datastore"
	"errors"
	"net/http"
	"strconv"
	"time"
)

//===================================================================
//       CMN
//===================================================================

type _StockPos struct {
	Models    QtyMap
	Articles  QtyMap
	DateValue time.Time
	DD_MMM_YY string
}

//===================================================================
//       GZB
//===================================================================
type GZBStockPos struct {
	_StockPos
}

//===================================================================
//       RRKStockDirtyDate
//===================================================================

var GLOBAL_RRK_STOCK_DIRTY_DATE RRKStockDirtyDate

type RRKStockDirtyDate struct {
	DateValue time.Time
}

func GetRRKDD(r *http.Request) (time.Time, error) {
	d := new(RRKStockDirtyDate)
	if err := d._GetOrCreateInDS(r); err != nil {
		return time.Time{}, err
	}
	return d.DateValue, nil
}

func RRKIntelligentlySetDD(r *http.Request, dv time.Time) error {
	//Just call this function with time from various places that can invalidate the stock and you are done. Next time stock is asked for, it will be recalculated from that date onwards.

	d := new(RRKStockDirtyDate)
	dv = StripTimeKeepDate(dv)

	oldRRKDD, err := GetRRKDD(r)
	if err != nil {
		return err
	}

	if oldRRKDD.Before(dv) {
		return nil
	}

	d.DateValue = dv
	myDebug(r, "Setting dirty date to: %s", DDMMMYYFromGoTime(dv))
	return d._SaveInDS(r)
}

func (d *RRKStockDirtyDate) _GetOrCreateInDS(r *http.Request) error {
	if !GLOBAL_RRK_STOCK_DIRTY_DATE.DateValue.IsZero() {
		*d = GLOBAL_RRK_STOCK_DIRTY_DATE
		return nil
	}
	c := appengine.NewContext(r)
	k := d._KeyDS(r)
	e := d
	if err := datastore.Get(c, k, e); err != nil && err != datastore.ErrNoSuchEntity {
		return err
	} else if err == datastore.ErrNoSuchEntity {
		d.DateValue = time.Now()
		return d._SaveInDS(r)
	}
	GLOBAL_RRK_STOCK_DIRTY_DATE = *d
	return nil
}

func (d *RRKStockDirtyDate) _KeyDS(r *http.Request) *datastore.Key {
	return RRK_SEWNewKey("RRKStockDirtyDate", "RRKStockDD", 0, r)
}

func (d *RRKStockDirtyDate) _SaveInDS(r *http.Request) error {
	d.DateValue = StripTimeKeepDate(d.DateValue)
	c := appengine.NewContext(r)
	k := d._KeyDS(r)
	e := d

	_, err := datastore.Put(c, k, e)
	if err != nil {
		return err
	}

	GLOBAL_RRK_STOCK_DIRTY_DATE = *d
	return nil
}
func RRKUnconditionalSaveDD(r *http.Request, dv time.Time) error {
	d := new(RRKStockDirtyDate)
	d.DateValue = dv
	return d._SaveInDS(r)
}

//===================================================================
//  RRKStockAsString
//===================================================================
type RRKStockAsString struct {
	StringData string `datastore:"noindex"`
}

func RRKBlankStockStruct(r *http.Request) (*RRKStockPos, error) {

	rrksp := new(RRKStockPos)
	models, err := GetAllModelsFromBOM(r)
	if err != nil {
		return nil, err
	}
	rrksp.Models = make(QtyMap)
	for _, model := range models {
		rrksp.Models[model.Name] = 0
	}

	rrksp.Articles = make(QtyMap)
	articles, err := GetAllArticlesFromDS(r)
	if err != nil {
		return nil, err
	}

	for _, article := range articles {
		rrksp.Articles[article.Name] = 0
	}

	return rrksp, nil

}

func (rrksas *RRKStockAsString) _KeyDS(r *http.Request, uid string) *datastore.Key {
	return RRK_SEWNewKey("RRKStockAsString", "RRKStock_"+uid, 0, r)
}

func (rrksas *RRKStockAsString) _GetOrCreateInDS(r *http.Request, uid string) (*RRKStockPos, error) {
	c := appengine.NewContext(r)
	k := rrksas._KeyDS(r, uid)
	e := rrksas

	err := datastore.Get(c, k, e)

	if err == datastore.ErrNoSuchEntity {
		rrksp, err := RRKBlankStockStruct(r)
		rrksp.DD_MMM_YY = uid
		rrksp.DateValue, err = DDMMMYYToGoTime(uid)
		if err != nil {
			return nil, err
		}
		if err := rrksas._SaveInDS(r, rrksp); err != nil {
			myDebug(r, "Error from: rrksas._SaveInDS():"+err.Error())
			return nil, err
		}
		return rrksp, nil
	}
	if err != nil {
		myDebug(r, "Error from: datastore.Get():"+err.Error())
		return nil, err
	}

	rrksp := new(RRKStockPos)
	data := rrksas.StringData
	myDebug(r, "\nString Data:\n%v", data)
	if err := JsonToStruct(&data, &rrksp, r); err != nil {
		myDebug(r, "Error from: JsonToStruct():"+err.Error())
		return nil, err
	}
	myDebug(r, "\nConverted Json:\n%#v", rrksp)

	return rrksp, nil
}

func (rrksas *RRKStockAsString) _SaveInDS(r *http.Request, rrksp *RRKStockPos) error {
	//Design flaw: we really dont need RRKStockAsString. we need only RRKStockPos
	data, err := StructToJson(rrksp, r)
	if err != nil {
		return err
	}
	c := appengine.NewContext(r)
	k := rrksas._KeyDS(r, rrksp.UID())
	e := &RRKStockAsString{StringData: *data}
	//myDebug(r, "\nSaving RRKStockPos for %v \n %#v \nas RRKStockAsString \n %v", rrksp.DD_MMM_YY, rrksp, *data)
	if _, err := datastore.Put(c, k, e); err != nil {
		return err
	}

	return nil
}

//===================================================================
//       RRKStockPos
//===================================================================
type RRKStockPos struct {
	_StockPos
}

func (rrksp *RRKStockPos) UID() string {
	return rrksp.DD_MMM_YY
}

func (rrksp *RRKStockPos) _GetOrCreateInDS(r *http.Request) error {
	//DD_MMM_YY should be present if you want to get anything from DS
	//If present in Cache, return from there.
	//Else load bytes buffer from DS
	//Convert it to struct
	//Save it in Global cache and return
	if rrksp.DD_MMM_YY == "" {
		return errors.New("Requesting the stock without a specific date.")
	}

	rrksas := new(RRKStockAsString)
	tsp, err := rrksas._GetOrCreateInDS(r, rrksp.UID())
	if err != nil {
		myDebug(r, "Error from: rrksas._GetOrCreateInDS():"+err.Error())
		return err
	}

	*rrksp = *tsp

	return nil
}

func (x *RRKStockPos) _KeyDS(r *http.Request) *datastore.Key {
	//Daily stock is saved with the DD_MMM_YY format so that it can be retrieved directly without querying.
	//return RRK_SEWNewKey("[]byte", "RRKStock_"+x.UID(), 0, r)
	return RRK_SEWNewKey("RRKStockPos", "RRKStock_"+x.UID(), 0, r)
}

func (rrksp *RRKStockPos) _SaveInDS(r *http.Request) error {
	//DD_MMM_YY must be set when trying to save anything in DS
	myDebug(r, "Got a stock save request for date "+rrksp.DD_MMM_YY)
	//1. Convert to bytes
	//2. Store in DS
	if rrksp.DD_MMM_YY == "" {
		return errors.New("Trying to save stock without Date.")
	}
	t, err := DDMMMYYToGoTime(rrksp.DD_MMM_YY)
	if err != nil {
		return err
	}
	//The whole datastore operates on DateValue. Hence storing it.
	rrksp.DateValue = StripTimeKeepDate(t) //Might be unnecessary but being defensive here.

	rrksas := new(RRKStockAsString)
	if err := rrksas._SaveInDS(r, rrksp); err != nil {
		return err
	}

	return nil
}

func rrkStockPositionForDateSlashApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		unixTimeAsString := r.URL.Path[len(API_RRK_STOCK_POSITION_FOR_DATE_SLASH_END):]
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
		stock, err := GetRRKstockForThisDDMMMYY(r, DD_MMM_YY)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := WriteJson(&w, stock); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return

	default:
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func GetRRKstockForThisDDMMMYY(r *http.Request, DD_MMM_YY string) (*RRKStockPos, error) {
	if err := RRKRecalculateStockSinceDirtyDate(r); err != nil {
		myDebug(r, "Error from: RRKRecalculateStockSinceDirtyDate():"+err.Error())
		return nil, err
	}
	rrkStockPosition := new(RRKStockPos)
	rrkStockPosition.DD_MMM_YY = DD_MMM_YY
	err := rrkStockPosition._GetOrCreateInDS(r)
	if err != nil {
		myDebug(r, "Error from: rrkStockPosition._GetOrCreateInDS():"+err.Error())
		return nil, err
	}
	return rrkStockPosition, nil
}

//===================================================================
//       RRKUtility Functions
//===================================================================
func RRKRecalculateStockSinceDirtyDate(r *http.Request) error {
	dd, err := GetRRKDD(r)
	if err != nil {
		return err
	}
	dd = StripTimeKeepDate(dd)
	today := StripTimeKeepDate(time.Now())
	for ; !dd.After(today); dd = dd.Add(24 * time.Hour) {
		//TODO: Do it in a transaction.
		if err := _CalculateAndSaveRRKStockForDate(r, dd); err != nil {
			return err
		}
	}
	return nil
}

func _CalculateAndSaveRRKStockForDate(r *http.Request, dirtyDate time.Time) error {
	//Thing about how task ques can be used.

	/////////////////////////////////////////////////////////////////////////
	//Logic:
	// It starts with dirty date.
	// Find good condition stock from yesterday(yesterday)
	// Apply operations
	// Save the new stock for dirty date
	// Increment the dirty date stock by 1.
	/////////////////////////////////////////////////////////////////////////

	dirtyDate = StripTimeKeepDate(dirtyDate)
	yesterday := dirtyDate.Add(-1 * 24 * time.Hour)
	tomorrow := dirtyDate.Add(1 * 24 * time.Hour)
	/////////////////////////////////////////////////////////////////////////
	//1. Get the yesterday's stock pos as reference for today's stock
	/////////////////////////////////////////////////////////////////////////
	c := appengine.NewContext(r)
	err1 := datastore.RunInTransaction(c, func(c appengine.Context) error {
		//TODO: See if everything inside a transaction is slowing things up?
		todaysStock, err := RRKBlankStockStruct(r)
		if err != nil {
			return logErr(r, err, "RRKBlankStockStruct()")
		}
		todaysStock.DD_MMM_YY = DDMMMYYFromGoTime(yesterday)

		if err := todaysStock._GetOrCreateInDS(r); err != nil {
			return logErr(r, err, "todaysStock._GetOrCreateInDS")
		}

		/////////////////////////////////////////////////////////////////////////
		//2. Calculate the effect of all transactions on this reference Stock
		/////////////////////////////////////////////////////////////////////////
		todaysStock.DD_MMM_YY = DDMMMYYFromGoTime(dirtyDate)
		todaysStock.DateValue, err = DDMMMYYToGoTime(todaysStock.DD_MMM_YY)
		if err != nil {
			return logErr(r, err, " DDMMMYYToGoTime(todaysStock.DD_MMM_YY)")
		}

		/////////////////////////////////////////////////////////////////////////
		//2a. Add daily assembled items to models
		/////////////////////////////////////////////////////////////////////////
		ais, err := RRKGetAllAssembledItemsOnSingleDay(r, dirtyDate)
		if err != nil {
			return err
		}
		for _, ai := range ais {
			for _, item := range ai.Items {
				todaysStock.Models[item.ModelName] += item.Quantity
			}
		}

		/////////////////////////////////////////////////////////////////////////
		//2b. Subtract daily sale from models
		/////////////////////////////////////////////////////////////////////////
		asi, err := RRKGetAllSaleInvoicesOnSingleDay(r, dirtyDate)
		if err != nil {
			return err
		}
		for _, si := range asi {
			for _, item := range si.Items {
				todaysStock.Models[item.Name] -= item.Quantity
			}
		}

		/////////////////////////////////////////////////////////////////////////
		//2c. Subtract daily production consumed articles from articles
		/////////////////////////////////////////////////////////////////////////

		/////////////////////////////////////////////////////////////////////////
		//2d. Subtract outward stock transfer
		/////////////////////////////////////////////////////////////////////////

		/////////////////////////////////////////////////////////////////////////
		//2e. Add inward stock transfer
		/////////////////////////////////////////////////////////////////////////

		/////////////////////////////////////////////////////////////////////////
		//2f. Add daily purchase to articles
		/////////////////////////////////////////////////////////////////////////
		rrkpis, err := RRKGetAllPurchaseInvoicesOnSingleDay(r, dirtyDate)
		if err != nil {
			return err
		}

		for _, pi := range rrkpis {
			for _, item := range pi.Items {
				todaysStock.Articles[item.Name] += item.Quantity
			}
		}

		err = todaysStock._SaveInDS(r)
		if err != nil {
			return logErr(r, err, "todaysStock._SaveInDS(r)")
		}

		//myDebug(r, "After calculation Stock on %s is:\n%#v", todaysStock.DD_MMM_YY, todaysStock)
		if err := RRKUnconditionalSaveDD(r, tomorrow); err != nil {
			return logErr(r, err, "RRKUnconditionalSaveDD(r, tomorrow)")
		}

		return nil
	}, nil) //Transaction ends
	return err1
}
