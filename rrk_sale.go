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
		myDebug(r, fmt.Sprintf("About to return:\n%#v", sis))
		if err := WriteJson(&w, sis); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	case "POST":
		var si SaleInvoice
		if err := json.NewDecoder(r.Body).Decode(&si); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c := appengine.NewContext(r)
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			if err1 := SaveRRKSaleInvoiceInDS(&si, r); err1 != nil {
				return err1
			}
			if err1 := SendMailForRRKSaleInvoice(&si, r); err1 != nil {
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

func RRKSaleInvoiceKey(r *http.Request, siUID string) *datastore.Key {
	return RRK_SEWNewKey("RRKSaleInvoice", siUID, 0, r)
}

func SaveRRKSaleInvoiceInDS(si *SaleInvoice, r *http.Request) error {
	//Have it as a method on saleInvoice? refactor later
	si.DateValue = time.Unix(si.JSDateValueAsSeconds, 0)
	si.DD_MMM_YY = DDMMYYFromGoTime(si.DateValue)
	//BUG: If the customer name is changed or any other id is mutated, the original one should first be deleted.
	si.UID = fmt.Sprintf("%s-%s-%s", si.DD_MMM_YY, si.Number, si.CustomerName)
	for i, item := range si.Items {
		model, err := GetModelWithName(r, item.Name)
		if err != nil {
			return err
		}
		si.Items[i].ModelVal = model
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(model); err != nil {
			return err
		}
		si.Items[i].ModelJSON = b.Bytes()
	}

	myDebug(r, fmt.Sprintf("Decoded si from network:\n%#v", si))
	c := appengine.NewContext(r)
	k := RRKSaleInvoiceKey(r, si.UID)
	e := si
	if _, err := datastore.Put(c, k, e); err != nil {
		myDebug(r, "Datastore put err")
		return err
	}
	return nil
}

func GetAllRRKSaleInvoicesFromDS(r *http.Request) ([]SaleInvoice, error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("RRKSaleInvoice").
		Order("-DateValue")
	//https://cloud.google.com/appengine/docs/go/datastore/reference
	var sis []SaleInvoice
	for t := q.Run(c); ; {
		var si SaleInvoice
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

func GetRRKSaleInvoiceFromDS(siUID string, r *http.Request) (*SaleInvoice, error) {
	//Have it as a method on saleInvoice? refactor later
	c := appengine.NewContext(r)
	k := RRKSaleInvoiceKey(r, siUID)
	e := new(SaleInvoice)
	if _, err := datastore.Put(c, k, e); err != nil {
		return nil, err
	}
	return e, nil
}

func SendMailForRRKSaleInvoice(si *SaleInvoice, r *http.Request) error {
	siDateAsDDMMYYYY := DDMMYYFromGoTime(si.DateValue)

	totalQuantitySold := 0
	for _, item := range si.Items {
		totalQuantitySold += item.Quantity
	}

	goodsValue := 0
	for _, item := range si.Items {
		goodsValue += item.Quantity * int(item.Rate)
	}

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"DDMMYYFromGoTime":         DDMMYYFromGoTime,
		"LogMsgShownForLogTime":    func(x time.Time) string { return LogMsgShownForLogTime(x, time.Now()) },
		"SingleItemGoodsValueFunc": func(i InvoiceItem) float64 { return i.Rate * float64(i.Quantity) },
	}

	emailTemplate := template.Must(template.New("emailRRKDS").Funcs(funcMap).Parse(`
	<html><body>
	<table border=1 cellpadding=5>
	<caption>
	<h4></u>{{.DateValue|LogMsgShownForLogTime }}</u></h4>
	<h4>M/s {{.CustomerName }}</h4>
	<h4>Invoice#: {{.Number }}</h4>
	<h4>{{.DateValue | DDMMYYFromGoTime}}</h4>
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
		Subject:  fmt.Sprintf("%s: Inv#%v | %v | %v pc sold [SEWPULSE][RRKDS]", siDateAsDDMMYYYY, si.Number, si.CustomerName, totalQuantitySold),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		return err
	}
	return nil
}
