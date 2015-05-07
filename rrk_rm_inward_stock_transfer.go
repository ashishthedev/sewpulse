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

const RRKRawMatInwardStockTransferInvoiceKind = "RRKRawMatInwardStockTransferInvoice"

func RRKRMISTInvoiceKey(r *http.Request, UID string) *datastore.Key {
	return RRK_SEWNewKey(RRKRawMatInwardStockTransferInvoiceKind, UID, 0, r)
}

func (x *RRKRMISTInvoice) SaveInDS(r *http.Request) error {
	c := appengine.NewContext(r)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		x.DateValue = time.Unix(x.JSDateValueAsSeconds, 0)
		x.DD_MMM_YY = DDMMMYYFromGoTime(x.DateValue)
		x.UID = fmt.Sprintf("%s-%s-%s", x.DD_MMM_YY, x.PartyName, x.Number)

		k := RRKRMISTInvoiceKey(r, x.UID)
		e := x
		if _, err1 := datastore.Put(c, k, e); err1 != nil {
			return err1
		}
		return RRKIntelligentlySetDD(r, x.DateValue)
	}, nil)
	if err != nil {
		return logErr(r, err, "Inside (x *RRKRMISTInvoice) SaveInDS()")
	}
	return nil
}

func RRKRMISTInvoiceNoSalshApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var rmist = new(RRKRMISTInvoice)
		if err := json.NewDecoder(r.Body).Decode(rmist); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rmist.DateValue = time.Unix(rmist.JSDateValueAsSeconds, 0)

		c := appengine.NewContext(r)
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			if err1 := rmist.SaveInDS(r); err1 != nil {
				return err1
			}
			if err1 := rmist.SendMailForRMISTInvoice(r); err1 != nil {
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

func (x *RRKRMISTInvoice) SendMailForRMISTInvoice(r *http.Request) error {
	DDMMMYYYY := DDMMMYYFromGoTime(x.DateValue)

	totalQuantity := 0.0
	for _, item := range x.Items {
		totalQuantity += item.Quantity
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

	emailTemplate := template.Must(template.New("emailRRKRMIST").Funcs(funcMap).Parse(`
	<html><body>
	<table border=1 cellpadding=5>
	<caption>
	<h4></u>{{.DateValue|LogMsgShownForLogTime }}</u></h4>
	<h4>M/s {{.PartyName }}</h4>
	<h4>Invoice#: {{.Number }}</h4>
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
	<font color="grey">SEW RRK RAW MATERIAL INWARD STOCK TRANSFER</font>
	</body></html>
	`))

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, x); err != nil {
		return err
	}
	finalHTML := buf.String()

	subject := fmt.Sprintf("%s: Inv#%v | %v | %v pc sold [SEWPULSE][RRK-RM-OST]", DDMMMYYYY, x.Number, x.PartyName, totalQuantity)
	if err := SendSEWMail(r, subject, finalHTML); err != nil {
		return err
	}
	return nil
}

func RRKGetAllRMISTInvoicesOnSingleDay(r *http.Request, date time.Time) ([]RRKRMISTInvoice, error) {
	singleDate := StripTimeKeepDate(date)
	justBeforeNextDay := singleDate.Add(1*24*time.Hour - time.Second)
	return RRKGetAllRMISTInvoicesBetweenTheseDatesInclusive(r, singleDate, justBeforeNextDay)
}

func RRKGetAllRMISTInvoicesBetweenTheseDatesInclusive(r *http.Request, starting time.Time, ending time.Time) ([]RRKRMISTInvoice, error) {

	q := datastore.NewQuery(RRKRawMatInwardStockTransferInvoiceKind).
		Filter("DateValue >=", starting).
		Filter("DateValue <=", ending).
		Order("-DateValue")

	c := appengine.NewContext(r)
	var rmists []RRKRMISTInvoice
	for t := q.Run(c); ; {
		var rmist RRKRMISTInvoice
		_, err := t.Next(&rmist)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		rmists = append(rmists, rmist)
	}

	return rmists, nil
}
