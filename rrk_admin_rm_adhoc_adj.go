package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

const RRKRawMatAdhocAdjInvoiceKind = "RRKRawMatAdhocAdjInvoice"

func RRKRMAAInvoiceKey(r *http.Request, UID string) *datastore.Key {
	return RRK_SEWNewKey(RRKRawMatAdhocAdjInvoiceKind, UID, 0, r)
}

func (x *RRKRMAAInvoice) SaveInDS(r *http.Request) error {
	c := appengine.NewContext(r)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		x.DateValue = time.Unix(x.JSDateValueAsSeconds, 0)
		x.DD_MMM_YY = DDMMMYYFromGoTime(x.DateValue)
		x.UID = fmt.Sprintf("%s-%v", x.DD_MMM_YY, x.JSDateValueAsSeconds)

		k := RRKRMAAInvoiceKey(r, x.UID)
		e := x
		if _, err1 := datastore.Put(c, k, e); err1 != nil {
			return err1
		}
		return RRKIntelligentlySetDD(r, x.DateValue)
	}, nil)
	if err != nil {
		return logErr(r, err, "Inside (x *RRKRMAAInvoice) SaveInDS()")
	}
	return nil
}

func RRKRMAAInvoiceNoSalshApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var rmaa = new(RRKRMAAInvoice)
		if err := json.NewDecoder(r.Body).Decode(rmaa); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rmaa.DateValue = time.Unix(rmaa.JSDateValueAsSeconds, 0)

		c := appengine.NewContext(r)
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			if err1 := rmaa.SaveInDS(r); err1 != nil {
				return err1
			}
			if err1 := rmaa.SendMailForRMAAInvoice(r); err1 != nil {
				return err1
			}
			return nil
		}, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return

	default:
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (x *RRKRMAAInvoice) SendMailForRMAAInvoice(r *http.Request) error {
	DDMMMYYYY := DDMMMYYFromGoTime(x.DateValue)

	totalQuantity := 0.0
	for _, item := range x.Items {
		if item.Quantity < 0 {
			totalQuantity -= item.Quantity
		} else {
			totalQuantity += item.Quantity
		}
	}

	goodsValue := 0.0
	for _, item := range x.Items {
		goodsValue += item.Quantity * item.Rate
	}

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"DDMMMYYFromGoTime":        DDMMMYYFromGoTime,
		"LogMsgShownForLogTime":    func(x time.Time) string { return LogMsgShownForLogTime(x, time.Now()) },
		"SingleItemGoodsValueFunc": func(i NameRateQuantity) float64 { return i.Rate * i.Quantity },
	}

	emailTemplate := template.Must(template.New("emailRRKrmaa").Funcs(funcMap).Parse(`
	<html><body>
	<table border=1 cellpadding=5>
	<caption>
	<h4></u>{{.DateValue|LogMsgShownForLogTime }}</u></h4>
	<h4>AD HOC STOCK ADJUSTMENT FOR RAW MATERIAL</h4>
	<h4>{{.DateValue | DDMMMYYFromGoTime}}</h4>
	</caption>
	<thead>
	<tr bgcolor=#838468>
	<th><font color='#000000'> Item </font></th>
	<th><font color='#000000'> Quantity </font></th>
	<th><font color='#000000'> Rate </font></th>
	<th><font color='#000000'> Amount </font></th>
	</tr>
	</thead>
	<tfoot>
	<tr>
	<td colspan=3>Goods Value:</td>
	<td colspan=1>&#8377; {{ .GoodsValue}}</td>
	</tr>
	<tr>
	<td colspan=3>Total Tax:</td>
	<td colspan=1>&#8377; {{.TotalTax }}</td>
	</tr>
	<tr>
	<td colspan=3>Total Freight:</td>
	<td colspan=1>&#8377; {{.TotalFreight }}</td>
	</tr>
	<tr>
	<td colspan=3>Grand Total:</td>
	<td colspan=1><font color="#DD472F"><b>&#8377; {{.GrandTotal }}</b></font></td>
	</tr>
	</tfoot>
	{{range .Items}}
		<tr>
		<td>{{.Name }}</td>
		<td>{{.Quantity }}</td>
		<td>&#8377; {{.Rate }}</td>
		<td>&#8377; {{.|SingleItemGoodsValueFunc }}</td>
		</tr>
	{{end}}
	</table>
	<h4>Remarks:{{if .Remarks}} <font color="#DD472F">{{.Remarks }}</font>{{else}} No remarks. {{end}}</h4>
	<font color="grey">SEW RRK RAW MATERIAL AD HOC ADJUSTMENT</font>
	</body></html>
	`))

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, x); err != nil {
		return err
	}
	finalHTML := buf.String()

	subject := fmt.Sprintf("%s: %v | %v pc adjustment [SEWPULSE][RRK-RM-ADHOCADJ]", DDMMMYYYY, x.DateValue.Local(), totalQuantity)
	if err := SendSEWMail(r, subject, finalHTML); err != nil {
		return err
	}
	return nil
}

func RRKGetAllRMAAInvoicesOnSingleDay(r *http.Request, date time.Time) ([]RRKRMAAInvoice, error) {
	return RRKGetAllRMAAInvoicesBetweenTheseDatesInclusive(r, BOD(date), EOD(date))
}

func RRKGetAllRMAAInvoicesBetweenTheseDatesInclusive(r *http.Request, starting time.Time, ending time.Time) ([]RRKRMAAInvoice, error) {

	q := datastore.NewQuery(RRKRawMatAdhocAdjInvoiceKind).
		Filter("DateValue >=", starting).
		Filter("DateValue <=", ending).
		Order("-DateValue")

	c := appengine.NewContext(r)
	var rmaas []RRKRMAAInvoice
	for t := q.Run(c); ; {
		var rmaa RRKRMAAInvoice
		_, err := t.Next(&rmaa)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		rmaas = append(rmaas, rmaa)
	}

	return rmaas, nil
}

func RRKGetAllRMAAInvoicesBeforeThisDateInclusiveKeysOnly(r *http.Request, date time.Time) ([]*datastore.Key, error) {
	q := datastore.NewQuery(RRKRawMatAdhocAdjInvoiceKind).
		Filter("DateValue <=", EOD(date)).KeysOnly()
	return q.GetAll(appengine.NewContext(r), nil)
}
