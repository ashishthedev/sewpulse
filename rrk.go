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
	"sort"
	"time"
)

type SoldItem struct {
	BillNumber   string
	ModelName    string
	Quantity     int
	Amount       int
	CustomerName string
	Rate         int
}

type SoldItemsJSONValues struct {
	DateTimeAsUnixTime int64
	Items              []SoldItem
}

type ProducedItem struct {
	ModelName        string
	Quantity         int
	Unit             string
	AssemblyLineName string
	Remarks          string
}

type ProducedItemsJSONValues struct {
	DateTimeAsUnixTime int64
	Items              []ProducedItem
}

type ModelSet struct {
	ModelNames []string
}

func GetModelSetKey(r *http.Request) *datastore.Key {
	//TODO: Once the implementation matures, remove this data from datastore
	// and read from BOM
	//TODO: Change it to COMMON_SEWNewkey later as it belongs to both the comps.
	return RRK_SEWNewKey("ModelSet", "modelSetKey", 0, r)
}

func GetModelsFromServer(r *http.Request) (*ModelSet, error) {
	c := appengine.NewContext(r)
	k := GetModelSetKey(r)
	e := new(ModelSet)
	if err := datastore.Get(c, k, e); err != nil {
		return e, err
	}
	return e, nil
}

func SaveModelsToServer(r *http.Request, modelSet *ModelSet) error {
	if _, err := datastore.Put(appengine.NewContext(r), GetModelSetKey(r), modelSet); err != nil {
		return err
	}
	return nil
}

func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func rrkAddModelNameApiHandler(w http.ResponseWriter, r *http.Request) {
	type NewModel struct {
		NewModelName string
	}
	newModel := new(NewModel)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(newModel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	modelSet, err := GetModelsFromServer(r)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			modelSet = &ModelSet{ModelNames: []string{}}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	newModelName := newModel.NewModelName

	//Check for pre-existence of model with this name.
	for _, x := range modelSet.ModelNames {
		if x == newModelName {
			myDebug(r, fmt.Sprintf("Model already exists. Not creating a new one with name %s ", newModelName))
			return
		}
	}

	myDebug(r, fmt.Sprintf("Adding a new model with name %s ", newModelName))

	modelSet.ModelNames = append(modelSet.ModelNames, newModelName)
	RemoveDuplicates(&modelSet.ModelNames)

	var ss sort.StringSlice = modelSet.ModelNames
	ss.Sort()

	if err := SaveModelsToServer(r, modelSet); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rrkGetModelApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	modelSet, err := GetModelsFromServer(r)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			json.NewEncoder(w).Encode(ModelSet{ModelNames: []string{}})
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	json.NewEncoder(w).Encode(modelSet)
}

func rrkDailySaleEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var si SaleInvoice
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&si); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	si.DateValue = time.Unix(si.JSDateValueAsSeconds, 0)
	myDebug(r, fmt.Sprintf("%#", si))

	if err := SendMailForRRKSaleInvoice(&si, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SendMailForRRKSaleInvoice(si *SaleInvoice, r *http.Request) (err error) {
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

func rrkDailyAssemblyEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var producedItemsAsJson ProducedItemsJSONValues
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&producedItemsAsJson); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	submissionDateTimeAsUnixTime := producedItemsAsJson.DateTimeAsUnixTime
	logTime := time.Unix(submissionDateTimeAsUnixTime, 0)
	logDateDDMMYY := DDMMYYFromUnixTime(submissionDateTimeAsUnixTime)
	logMsg := LogMsgShownForLogTime(logTime, time.Now())

	totalQuantityProduced := 0
	for _, pi := range producedItemsAsJson.Items {
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
		logDateDDMMYY,
		logMsg,
		totalQuantityProduced,
	)

	for _, pi := range producedItemsAsJson.Items {
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
		Subject:  fmt.Sprintf("%s: %v pc [SEWPULSE][RRKDA]", logDateDDMMYY, totalQuantityProduced),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func rrkDailyPolishEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var producedItemsAsJson ProducedItemsJSONValues
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&producedItemsAsJson); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	submissionDateTimeAsUnixTime := producedItemsAsJson.DateTimeAsUnixTime
	logTime := time.Unix(submissionDateTimeAsUnixTime, 0)
	logDateDDMMYY := DDMMYYFromUnixTime(submissionDateTimeAsUnixTime)
	logMsg := LogMsgShownForLogTime(logTime, time.Now())

	totalQuantityProduced := 0
	for _, pi := range producedItemsAsJson.Items {
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
		logDateDDMMYY,
		logMsg,
		totalQuantityProduced,
	)

	for _, pi := range producedItemsAsJson.Items {
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
		Subject:  fmt.Sprintf("%s: %v pc [SEWPULSE][RRKDP]", logDateDDMMYY, totalQuantityProduced),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
