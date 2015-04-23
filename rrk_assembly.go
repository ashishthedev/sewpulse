package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/user"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RRKAssembledItem struct {
	ModelName        string
	Quantity         float64
	Unit             string
	AssemblyLineName string
	Remarks          string
}

type RRKAssembledItems struct {
	JSDateValueAsSeconds int64 `datastore:"-"`
	Items                []RRKAssembledItem
	DateValue            time.Time
}

func (ais RRKAssembledItems) UID() string {
	const layout = "02Jan2006_304pm_MST"
	//10Nov2009_1100am
	t := time.Unix(ais.JSDateValueAsSeconds, 0)
	return fmt.Sprintf(t.UTC().Format(layout))
}

func (ais RRKAssembledItems) GetOrCreateInDS(r *http.Request) error {
	c := appengine.NewContext(r)
	k := ais.KeyDS(r)
	e := &ais
	return datastore.Get(c, k, e)
}

const RRKAssembledItemsKind = "RRKAssembledItems"

func (ais RRKAssembledItems) KeyDS(r *http.Request) *datastore.Key {
	return RRK_SEWNewKey(RRKAssembledItemsKind, "RRKAssItems_"+ais.UID(), 0, r)
}

func (ais RRKAssembledItems) SaveInDS(r *http.Request) error {
	//TODO: All save in DS should be done in transactions. Find pattern and do it on mass scale.
	c := appengine.NewContext(r)
	k := ais.KeyDS(r)
	e := &ais
	if _, err := datastore.Put(c, k, e); err != nil {
		return err
	}
	return RRKIntelligentlySetDD(r, ais.DateValue)
}

func rrkDailyAssemblySubmissionApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var ais RRKAssembledItems
	if err := json.NewDecoder(r.Body).Decode(&ais); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ais.DateValue = time.Unix(ais.JSDateValueAsSeconds, 0)

	c := appengine.NewContext(r)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		if err1 := ais.SaveInDS(r); err1 != nil {
			return err1
		}

		if err1 := SendEmailForDailyAssembly(&ais, r); err1 != nil {
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

func SendEmailForDailyAssembly(ais *RRKAssembledItems, r *http.Request) error {

	aisDateAsUnixTime := ais.JSDateValueAsSeconds
	aisGoDate := time.Unix(aisDateAsUnixTime, 0)
	assembledOnDateDDMMMYY := DDMMMYYFromUnixTime(aisDateAsUnixTime)
	logMsg := LogMsgShownForLogTime(aisGoDate, time.Now())

	totalQuantityProduced := float64(0)
	for _, pi := range ais.Items {
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
		assembledOnDateDDMMMYY,
		logMsg,
		totalQuantityProduced,
	)

	for _, pi := range ais.Items {
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
		Subject:  fmt.Sprintf("%s: %v pc [SEWPULSE][RRKDA]", assembledOnDateDDMMMYY, totalQuantityProduced),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		return err
	}

	return nil
}

func RRKGetAllAssembledItemsOnSingleDay(r *http.Request, date time.Time) ([]RRKAssembledItems, error) {
	singleDate := StripTimeKeepDate(date)
	justBeforeNextDay := singleDate.Add(1*24*time.Hour - time.Second)
	return RRKGetAllAssembledItemsBetweenTheseDatesInclusive(r, singleDate, justBeforeNextDay)
}
func RRKGetAllAssembledItemsBetweenTheseDatesInclusive(r *http.Request, starting time.Time, ending time.Time) ([]RRKAssembledItems, error) {
	q := datastore.NewQuery(RRKAssembledItemsKind).
		Filter("DateValue >=", starting).
		Filter("DateValue <=", ending).
		Order("-DateValue")

	c := appengine.NewContext(r)
	var ais []RRKAssembledItems
	for t := q.Run(c); ; {
		var ai RRKAssembledItems
		_, err := t.Next(&ai)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		ais = append(ais, ai)
	}

	return ais, nil
}
