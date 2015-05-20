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

func rrkPurchaseInvoiceWithSalshApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		piUID := r.URL.Path[len(API_RRK_PURCHASE_INVOICE_SALSH_END):]
		pi, err := GetRRKPurchaseInvoiceFromDS(piUID, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := WriteJson(&w, pi); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	case "DELETE":
		piUID := r.URL.Path[len(API_RRK_PURCHASE_INVOICE_SALSH_END):]
		if err := DeleteRRKPurchaseInvoiceFromDS(piUID, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	default:
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func rrkPurchaseInvoiceApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		pis, err := GetAllRRKPurchaseInvoicesFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := WriteJson(&w, struct{ RRKPurchaseInvoices []RRKPurchaseInvoice }{pis}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	case "POST":

		var pi RRKPurchaseInvoice
		if err := json.NewDecoder(r.Body).Decode(&pi); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pi.DateValue = time.Unix(pi.JSDateValueAsSeconds, 0)

		c := appengine.NewContext(r)
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			if err1 := pi.SaveInDS(r); err1 != nil {
				return err1
			}
			if err1 := SendMailForRRKPurchaseInvoice(&pi, r); err1 != nil {
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

const RRKPurchaseInvoiceKind = "RRKPurchaseInvoice"

func RRKPurchaseInvoiceKey(r *http.Request, piUID string) *datastore.Key {
	return RRK_SEWNewKey(RRKPurchaseInvoiceKind, piUID, 0, r)
}

func (pi *RRKPurchaseInvoice) GetUID() string {
	return fmt.Sprintf("%s-%s-%s", pi.DD_MMM_YY, pi.SupplierName, pi.Number)
}

func (pi *RRKPurchaseInvoice) SaveInDS(r *http.Request) error {
	c := appengine.NewContext(r)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		pi.DateValue = time.Unix(pi.JSDateValueAsSeconds, 0)
		pi.DD_MMM_YY = DDMMMYYFromGoTime(pi.DateValue)
		pi.UID = pi.GetUID()

		k := RRKPurchaseInvoiceKey(r, pi.GetUID())
		e := pi
		if _, err1 := datastore.Put(c, k, e); err1 != nil {
			return err1
		}
		return RRKIntelligentlySetDD(r, pi.DateValue)
	}, nil)
	if err != nil {
		return logErr(r, err, "Inside (pi *RRKPurchaseInvoice) SaveInDS()")
	}
	return nil
}

func GetAllRRKPurchaseInvoicesFromDS(r *http.Request) ([]RRKPurchaseInvoice, error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery(RRKPurchaseInvoiceKind).
		Order("-DateValue")
	//https://cloud.google.com/appengine/docs/go/datastore/reference
	var pis []RRKPurchaseInvoice
	for t := q.Run(c); ; {
		var pi RRKPurchaseInvoice
		_, err := t.Next(&pi)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		pis = append(pis, pi)
	}

	return pis, nil
}

func DeleteRRKPurchaseInvoiceFromDS(piUID string, r *http.Request) error {
	c := appengine.NewContext(r)
	k := RRKPurchaseInvoiceKey(r, piUID)
	return datastore.Delete(c, k)
}

func GetRRKPurchaseInvoiceFromDS(piUID string, r *http.Request) (*RRKPurchaseInvoice, error) {
	c := appengine.NewContext(r)
	k := RRKPurchaseInvoiceKey(r, piUID)
	e := new(RRKPurchaseInvoice)
	if err := datastore.Get(c, k, e); err != nil {
		return nil, err
	}
	return e, nil
}

func SendMailForRRKPurchaseInvoice(pi *RRKPurchaseInvoice, r *http.Request) error {
	piDateAsDDMMMYYYY := DDMMMYYFromGoTime(pi.DateValue)

	totalQuantitySold := float64(0)
	for _, item := range pi.Items {
		totalQuantitySold += item.Quantity
	}

	goodsValue := float64(0)
	for _, item := range pi.Items {
		goodsValue += item.Quantity * item.Rate
	}

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"DDMMMYYFromGoTime":        DDMMMYYFromGoTime,
		"LogMsgShownForLogTime":    func(x time.Time) string { return LogMsgShownForLogTime(x, time.Now()) },
		"SingleItemGoodsValueFunc": func(i NameRateQuantity) float64 { return i.Rate * i.Quantity },
	}

	emailTemplate := template.Must(template.New("emailRRKDPUR").Funcs(funcMap).Parse(`
	<html><body>
	<table border=1 cellpadding=5>
	<caption>
	<h4></u>{{.DateValue|LogMsgShownForLogTime }}</u></h4>
	<h4>M/s {{.SupplierName }}</h4>
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
	<font color="grey">SEW RRK PURCHASE</font>
	</body></html>
	`))

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, pi); err != nil {
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
		Subject:  fmt.Sprintf("%s: Inv#%v | %v | %v pc sold [SEWPULSE][RRKPUR]", piDateAsDDMMMYYYY, pi.Number, pi.SupplierName, totalQuantitySold),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		return err
	}
	return nil
}

func HTTPSinglePurchaseInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t := template.Must(template.ParseFiles("templates/admin/rrk_purchase_invoice_single.html"))
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

func RRKGetAllPurchaseInvoicesOnSingleDay(r *http.Request, date time.Time) ([]RRKPurchaseInvoice, error) {
	return RRKGetAllPurchaseInvoicesBetweenTheseDatesInclusive(r, BOD(date), EOD(date))
}
func RRKGetAllPurchaseInvoicesBetweenTheseDatesInclusive(r *http.Request, starting time.Time, ending time.Time) ([]RRKPurchaseInvoice, error) {

	q := datastore.NewQuery(RRKPurchaseInvoiceKind).
		Filter("DateValue >=", starting).
		Filter("DateValue <=", ending).
		Order("-DateValue")

	c := appengine.NewContext(r)
	var pis []RRKPurchaseInvoice
	for t := q.Run(c); ; {
		var pi RRKPurchaseInvoice
		_, err := t.Next(&pi)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		pis = append(pis, pi)
	}

	return pis, nil
}

func RRKGetAllPurchaseItemsBeforeThisDateInclusiveKeysOnly(r *http.Request, date time.Time) ([]*datastore.Key, error) {
	q := datastore.NewQuery(RRKPurchaseInvoiceKind).
		Filter("DateValue <=", EOD(date)).KeysOnly()
	return q.GetAll(appengine.NewContext(r), nil)
}
