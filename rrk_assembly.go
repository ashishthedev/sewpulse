package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/user"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func (ai RRKAssembledItem) UID() string {
	const layout = "02Jan2006_304pm_MST"
	//10Nov2009_1100am_UTC
	return fmt.Sprintf("%s-%s", ai.AssemblyLineName, ai.ModelName, ai.DateValue.Format(layout))
}

const RRKAssembledItemKind = "RRKAssembledItem"

func (ai RRKAssembledItem) KeyDS(r *http.Request) *datastore.Key {
	return RRK_SEWNewKey(RRKAssembledItemKind, "RRKAsmbldItem_"+ai.UID(), 0, r)
}

func (ai RRKAssembledItem) SaveInDS(r *http.Request) error {
	if ai.DateValue.IsZero() {
		return errors.New("Trying to save assembled item without Date.")
	}
	model, err := GetModelWithName(r, ai.ModelName)
	if err != nil {
		return err
	}
	data, err := StructToJson(model, r)
	if err != nil {
		return err
	}

	ai.ModelWithFullBOMAsString = *data

	c := appengine.NewContext(r)
	k := ai.KeyDS(r)
	e := &ai
	err1 := datastore.RunInTransaction(c, func(c appengine.Context) error {
		if _, err := datastore.Put(c, k, e); err != nil {
			return err
		}
		return RRKIntelligentlySetDD(r, ai.DateValue)
	}, nil)

	if err1 != nil {
		return err1
	}
	return nil
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
	err1 := datastore.RunInTransaction(c, func(c appengine.Context) error {
		for _, ai := range ais.Items {
			ai.DateValue = ais.DateValue
			if err := ai.SaveInDS(r); err != nil {
				return err
			}
		}

		if err := SendEmailForDailyAssembly(&ais, r); err != nil {
			return err
		}
		return nil
	}, nil)

	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
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
	<td>%v</td>
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

func RRKGetAllAssembledItemsOnSingleDay(r *http.Request, date time.Time) ([]RRKAssembledItem, error) {
	dayBeginning := StripTimeKeepDate(date)
	return RRKGetAllAssembledItemsBetweenTheseDatesInclusive(r, dayBeginning, EOD(dayBeginning))
}

func RRKGetAllAssembledItemsBeforeThisDateInclusiveKeysOnly(r *http.Request, date time.Time) ([]*datastore.Key, error) {
	q := datastore.NewQuery(RRKAssembledItemKind).
		Filter("DateValue <=", EOD(date)).KeysOnly()
	return q.GetAll(appengine.NewContext(r), nil)
}

func RRKGetAllAssembledItemsBetweenTheseDatesInclusive(r *http.Request, starting time.Time, ending time.Time) ([]RRKAssembledItem, error) {
	q := datastore.NewQuery(RRKAssembledItemKind).
		Filter("DateValue >=", starting).
		Filter("DateValue <=", ending).
		Order("-DateValue")

	c := appengine.NewContext(r)
	var ais []RRKAssembledItem
	for t := q.Run(c); ; {
		var ai RRKAssembledItem
		_, err := t.Next(&ai)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if err := JsonToStruct(&(ai.ModelWithFullBOMAsString), &(ai.ModelWithFullBOM), r); err != nil {
			return nil, err
		}
		ais = append(ais, ai)
	}

	return ais, nil
}
