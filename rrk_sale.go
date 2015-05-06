package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/user"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func rrkSaleInvoiceWithSalshApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		siUID := r.URL.Path[len(API_RRK_SALE_INVOICE_SALSH_END):]
		si, err := GetRRKSaleInvoiceFromDS(siUID, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := WriteJson(&w, si); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	case "DELETE":
		siUID := r.URL.Path[len(API_RRK_SALE_INVOICE_SALSH_END):]
		if err := DeleteRRKSaleInvoiceFromDS(siUID, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	default:
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func rrkSaleInvoiceApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		sis, err := GetAllRRKSaleInvoicesFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		myDebug(r, "Fetched sale invoices: %#v", sis)
		if err := WriteJson(&w, struct{ RRKSaleInvoices []RRKSaleInvoice }{sis}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	case "POST":
		var si RRKSaleInvoice
		if err := json.NewDecoder(r.Body).Decode(&si); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		myDebug(r, "Just decoded new SI from Web: \n%#v", si)

		c := appengine.NewContext(r)
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			if err1 := si.SaveRRKSaleInvoiceInDS(r); err1 != nil {
				return err1
			}
			if err1 := si.SendMailForRRKSaleInvoice(r); err1 != nil {
				return err1
			}
			return nil
		}, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return

	}
}

const RRKSaleInvoiceKind = "RRKSaleInvoice"

func RRKSaleInvoiceKey(r *http.Request, siUID string) *datastore.Key {
	return RRK_SEWNewKey(RRKSaleInvoiceKind, siUID, 0, r)
}

func (si *RRKSaleInvoice) GetUID() string {
	return fmt.Sprintf("%s-%s-%s", si.DD_MMM_YY, si.Number, si.CustomerName)
}
func (si *RRKSaleInvoice) SaveRRKSaleInvoiceInDS(r *http.Request) error {
	si.DateValue = GoTimeFromUnixTime(si.JSDateValueAsSeconds)
	si.DD_MMM_YY = DDMMMYYFromGoTime(si.DateValue)
	si.UID = si.GetUID()
	for i, item := range si.Items {
		model, err := GetModelWithName(r, item.Name)
		if err != nil {
			return err
		}
		data, err := StructToJson(model, r)
		if err != nil {
			return err
		}

		si.Items[i].ModelWithFullBOMAsString = *data
		//TODO: This may not be necessary. Because we are saving the ArticleAndQty in assembled items.
	}

	c := appengine.NewContext(r)
	err1 := datastore.RunInTransaction(c, func(c appengine.Context) error {
		k := RRKSaleInvoiceKey(r, si.UID)
		if _, err := datastore.Put(c, k, si); err != nil {
			return err
		}
		return RRKIntelligentlySetDD(r, si.DateValue)
	}, nil)
	return err1
}

func DeleteRRKSaleInvoiceFromDS(siUID string, r *http.Request) error {
	c := appengine.NewContext(r)
	k := RRKSaleInvoiceKey(r, siUID)
	return datastore.Delete(c, k)
}

func GetRRKSaleInvoiceFromDS(siUID string, r *http.Request) (*RRKSaleInvoice, error) {
	c := appengine.NewContext(r)
	k := RRKSaleInvoiceKey(r, siUID)
	e := new(RRKSaleInvoice)

	if err := datastore.Get(c, k, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (si *RRKSaleInvoice) SendMailForRRKSaleInvoice(r *http.Request) error {
	siDateAsDDMMMYYYY := DDMMMYYFromGoTime(si.DateValue)

	totalQuantitySold := float64(0)
	for _, item := range si.Items {
		totalQuantitySold += item.Quantity
	}

	goodsValue := 0.0
	for _, item := range si.Items {
		goodsValue += item.Quantity * item.Rate
	}

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"DDMMMYYFromGoTime":        DDMMMYYFromGoTime,
		"LogMsgShownForLogTime":    func(x time.Time) string { return LogMsgShownForLogTime(x, time.Now()) },
		"SingleItemGoodsValueFunc": func(i SoldItem) float64 { return i.Rate * float64(i.Quantity) },
	}

	emailTemplate := template.Must(template.New("emailRRKDS").Funcs(funcMap).Parse(`
	<html><body>
	<table border=1 cellpadding=5>
	<caption>
	<h4></u>{{.DateValue|LogMsgShownForLogTime }}</u></h4>
	<h4>M/s {{.CustomerName }}</h4>
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
	<font color="grey">SEW RRK</font>
	</body></html>
	`))

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, si); err != nil {
		return err
	}
	finalHTML := buf.String()

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
		Subject:  fmt.Sprintf("%s: Inv#%v | %v | %v pc sold [SEWPULSE][RRKDS]", siDateAsDDMMMYYYY, si.Number, si.CustomerName, totalQuantitySold),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		return err
	}
	return nil
}

func HTTPSingleSaleInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t := template.Must(template.ParseFiles("templates/admin/rrk_sale_invoice_single.html"))
		if err := t.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	default:
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return

	}
}
func RRKGetAllSaleInvoicesOnSingleDay(r *http.Request, date time.Time) ([]RRKSaleInvoice, error) {
	singleDate := StripTimeKeepDate(date)
	justBeforeNextDay := singleDate.Add(1*24*time.Hour - time.Second)
	return RRKGetAllSaleInvoicesBetweenTheseDatesInclusive(r, singleDate, justBeforeNextDay)
}

func GetAllRRKSaleInvoicesFromDS(r *http.Request) ([]RRKSaleInvoice, error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery(RRKSaleInvoiceKind)
	//https://cloud.google.com/appengine/docs/go/datastore/reference
	var sis []RRKSaleInvoice
	for t := q.Run(c); ; {
		var si RRKSaleInvoice
		_, err := t.Next(&si)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		sis = append(sis, si)
	}

	return sis, nil
}

func RRKGetAllSaleInvoicesBetweenTheseDatesInclusive(r *http.Request, starting time.Time, ending time.Time) ([]RRKSaleInvoice, error) {
	q := datastore.NewQuery(RRKSaleInvoiceKind).
		Filter("DateValue >=", starting).
		Filter("DateValue <=", ending)

	c := appengine.NewContext(r)
	var sis []RRKSaleInvoice
	for t := q.Run(c); ; {
		var si RRKSaleInvoice
		_, err := t.Next(&si)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		sis = append(sis, si)
	}

	return sis, nil
}
