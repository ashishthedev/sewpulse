package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/user"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

type ModelSet struct {
	ModelNames []string
}

func GetModelSetKey(r *http.Request) *datastore.Key {
	//Delete this once the models are being picked up from bom
	//TODO: Once the implementation matures, remove this data from datastore
	// and read from BOM
	//TODO: Change it to COMMON_SEWNewkey later as it belongs to both the comps.
	return RRK_SEWNewKey("ModelSet", "modelSetKey", 0, r)
}

func GetModelsFromServer(r *http.Request) (*ModelSet, error) {
	//Delete this once the models are being picked up from bom
	c := appengine.NewContext(r)
	k := GetModelSetKey(r)
	e := new(ModelSet)
	if err := datastore.Get(c, k, e); err != nil {
		return e, err
	}
	return e, nil
}

func SaveModelsToServer(r *http.Request, modelSet *ModelSet) error {
	//Delete this once the models are being picked up from bom
	if _, err := datastore.Put(appengine.NewContext(r), GetModelSetKey(r), modelSet); err != nil {
		return err
	}
	return nil
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
	switch r.Method {
	case "POST":
		modelSet, err := GetModelsFromServer(r)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				if err := json.NewEncoder(w).Encode(ModelSet{ModelNames: []string{}}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := json.NewEncoder(w).Encode(modelSet); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	default:
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

type PolishedItem struct {
	ModelName        string
	Quantity         int
	Unit             string
	AssemblyLineName string
	Remarks          string
}

type PolishedItems struct {
	JSDateValueAsSeconds int64
	Items                []PolishedItem
}

func rrkDailyPolishEmailSendApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var polishedItems PolishedItems
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&polishedItems); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	submissionDateTimeAsUnixTime := polishedItems.JSDateValueAsSeconds
	logTime := time.Unix(submissionDateTimeAsUnixTime, 0)
	logDateDDMMMYY := DDMMMYYFromUnixTime(submissionDateTimeAsUnixTime)
	logMsg := LogMsgShownForLogTime(logTime, time.Now())

	totalQuantityProduced := 0
	for _, pi := range polishedItems.Items {
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
		logDateDDMMMYY,
		logMsg,
		totalQuantityProduced,
	)

	for _, pi := range polishedItems.Items {
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
		Subject:  fmt.Sprintf("%s: %v pc [SEWPULSE][RRKDP]", logDateDDMMMYY, totalQuantityProduced),
		HTMLBody: finalHTML,
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
