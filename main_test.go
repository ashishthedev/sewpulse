package sewpulse

import (
	"appengine"
	"appengine/aetest"
	"appengine/datastore"
	"appengine/user"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var newTests = []struct {
	id             int
	minutesBefore  int64
	expectedLogMsg string
}{
	{1, -.5 * 60, DaysMsg["ON_TIME"]},
	{2, -1 * 60, DaysMsg["ON_TIME"]},
	{3, -2 * 60, DaysMsg["ON_TIME"]},
	{4, -20 * 60, DaysMsg["ON_TIME"]},
	{5, -23 * 60, DaysMsg["ONE_DAY_OLD"]},
	{6, -23*60 - 58, DaysMsg["ONE_DAY_OLD"]},
	{7, -24 * 60, DaysMsg["ONE_DAY_OLD"]},
	{8, -25 * 60, DaysMsg["ONE_DAY_OLD"]},
	{9, -47 * 60, fmt.Sprintf(DaysMsg["X_DAYS_OLD"], 2)},
	{10, -2 * 24 * 60, fmt.Sprintf(DaysMsg["X_DAYS_OLD"], 2)},
	{11, -45 * 24 * 60, fmt.Sprintf(DaysMsg["X_DAYS_OLD"], 45)},
}

func TestReportedDays(t *testing.T) {
	for _, n := range newTests {
		tenPM := time.Date(2014, 12, 16, 22, 0, 0, 0, time.UTC) //10:00pm
		newTime := tenPM.Add(time.Duration(n.minutesBefore) * time.Minute)
		resultMsg := LogMsgShownForLogTime(newTime, tenPM)
		if resultMsg != n.expectedLogMsg {
			t.Fatalf("LogMsgShownForLogTime for %d minutes earlier\nTest#%d:\ntenPM=%v\nnewTime=%v\nGot = %s\nExpected = %s", n.minutesBefore, n.id, tenPM, newTime, resultMsg, n.expectedLogMsg)
		}
	}
}

func addNameToList(r *http.Request, name string) error {
	type MasterCopy struct {
		Names []string
	}
	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "MasterCopy", "MasterCopyKey", 0, nil)
	var x MasterCopy
	if err := datastore.Get(c, k, &x); err != nil && err != datastore.ErrNoSuchEntity {
		return err
	}
	fmt.Printf("\nRetrived master copy from datastore: %v", x)

	x.Names = append(x.Names, name)
	fmt.Printf("\nSaving the master copy in datastore: %v", x)
	if _, err := datastore.Put(c, k, &x); err != nil {
		return err
	}
	return nil
}

func noTestSimpleDSTest(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	for _, name := range []string{"alpha", "beta", "gamma"} {
		r, err := inst.NewRequest("POST", "/addName", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := addNameToList(r, name); err != nil {
			t.Error(err)
		}

	}
	return
}

func expectedStockAtThisStage(r *http.Request, t *testing.T, expectedStock *RRKStockPos, stage string) {
	rrkStock, err := GetRRKstockForThisDDMMMYY(r, DDMMMYYFromGoTime(time.Now()))
	if err != nil {
		t.Errorf(err.Error())
	}
	if !AreTwoStocksSame(rrkStock, expectedStock, t, stage) {
		t.Errorf(stage + fmt.Sprintf(":\nExpectedStock:\n%v\nGot Stock:\n%v", expectedStock, rrkStock))
		return
	}

	return
}

func expectedBomAndArticlesAtThisStage(r *http.Request, t *testing.T, expectedBOM *BOM, stage string) {
	bomFromDS, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		t.Error(err)
		return
	}

	if IsBOMHavingSameArticleListAcrossModels(expectedBOM, t, stage) != true {
		t.Errorf(stage + ": " + "Wrong hardcoded values")
		return
	}

	if IsBOMHavingSameArticleListAcrossModels(bomFromDS, t, stage) != true {
		t.Errorf(stage + ": " + "BOM from DS have models which have different article list. This is a major assumption in rest of the code.")
		return
	}

	if AreTwoBOMsEqual(expectedBOM, bomFromDS, t, stage) != true {
		t.Errorf(stage + fmt.Sprintf(":\nExpectedBOM:\n%v\nGot:\n%v", expectedBOM, bomFromDS))
		return
	}

	if !AreTwoArticleListsSame(bomFromDS.AML, expectedBOM.AML, t, stage) {
		t.Errorf(fmt.Sprintf(stage+": "+"\nWanted: %v,\n Got %v", bomFromDS.AML, expectedBOM.AML))
		return
	}

	return

}

func TestEndToEndCaseForBOMManipulation(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	u := &user.User{Email: "test@example.com"}
	//======================================================
	// Create new articles
	//======================================================

	BURNER := Article{
		Name:              "BURNER",
		Unit:              "pc",
		ReorderPoint:      1500,
		EOQ:               2000,
		MinimumAlarmPoint: 500,
		MaximumAlarmPoint: 3500,
	}

	KNOB := Article{
		Name:              "KNOB",
		Unit:              "pc",
		ReorderPoint:      1500,
		EOQ:               2000,
		MinimumAlarmPoint: 500,
		MaximumAlarmPoint: 3500,
	}

	for _, article := range []Article{BURNER, KNOB} {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(article); err != nil {
			t.Fatal(err)
		}

		r, err := inst.NewRequest("POST", API_BOM_ARTICLE_END, &b)
		if err != nil {
			t.Fatal(err)
		}

		aetest.Login(u, r)
		w := httptest.NewRecorder()
		bomArticleWithoutSalshAPIHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Body:%v", w.Body.String())
		} else {
			fmt.Println("Created article: ", article.Name)
		}

	}

	r, err := inst.NewRequest("POST", API_BOM_ARTICLE_SLASH_END, nil)
	if err != nil {
		t.Fatal(err)
	}

	//======================================================
	// Compare with expected list
	//======================================================
	expectedBOM := NewBOM()
	expectedBOM.AML = ArticleMap{KNOB.Name: KNOB, BURNER.Name: BURNER}

	expectedBomAndArticlesAtThisStage(r, t, expectedBOM, "Just created two articles")

	//======================================================
	// Create new models
	//======================================================

	PREMIUM_PLUS := Model{Name: "PremiumPlus", Unit: "pc", ArticleAndQty: QtyMap{BURNER.Name: 0, KNOB.Name: 0}}
	SAPPHIRE := Model{Name: "Sapphire", Unit: "pc", ArticleAndQty: QtyMap{BURNER.Name: 0, KNOB.Name: 0}}

	for _, singleModel := range []Model{PREMIUM_PLUS, SAPPHIRE} {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(singleModel); err != nil {
			t.Fatal(err)
		}

		r, err := inst.NewRequest("POST", API_BOM_MODEL_END, &b)
		if err != nil {
			t.Fatal(err)
		}

		aetest.Login(u, r)
		w := httptest.NewRecorder()
		bomModelWithoutSlashAPIHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("\nError occured:\n HTML Body:%v", w.Body.String())
		}
	}
	//======================================================
	// Compare with expected list
	//======================================================
	expectedBOM = NewBOM()
	expectedBOM.AML = ArticleMap{KNOB.Name: KNOB, BURNER.Name: BURNER}
	expectedBOM.Models = map[string]Model{
		"PremiumPlus": {
			Name:          "PremiumPlus",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0, BURNER.Name: 0},
		},
		"Sapphire": {
			Name:          "Sapphire",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0, BURNER.Name: 0},
		},
	}
	expectedBomAndArticlesAtThisStage(r, t, expectedBOM, "stage2: models created after articles")

	//======================================================
	// Create new articles again and see if they are added to existing models
	//======================================================
	GUARD := Article{
		Name:              "GUARD",
		Unit:              "gm",
		ReorderPoint:      1501,
		EOQ:               2001,
		MinimumAlarmPoint: 501,
		MaximumAlarmPoint: 3501,
	}

	for _, singleArticle := range []Article{GUARD} {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(singleArticle); err != nil {
			t.Fatal(err)
		}

		r, err := inst.NewRequest("POST", API_BOM_ARTICLE_END, &b)
		if err != nil {
			t.Fatal(err)
		}

		aetest.Login(u, r)
		//if err := CreateArticle(r); err != nil {
		//	t.Error(err)
		//}
		w := httptest.NewRecorder()
		bomArticleWithoutSalshAPIHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Body:%v", w.Body.String())
		}
	}

	expectedBOM = NewBOM()
	expectedBOM.Models = map[string]Model{
		"PremiumPlus": {
			Name:          "PremiumPlus",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0, GUARD.Name: 0.0},
		},
		"Sapphire": {
			Name:          "Sapphire",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0, GUARD.Name: 0.0},
		},
	}
	expectedBOM.AML = ArticleMap{KNOB.Name: KNOB, BURNER.Name: BURNER, GUARD.Name: GUARD}
	expectedBomAndArticlesAtThisStage(r, t, expectedBOM, "stage3: Creating new articles again after models")

	//======================================================
	// Create new models again
	//======================================================

	RUBY := Model{Name: "Ruby", Unit: "pc"}

	for _, singleModel := range []Model{RUBY} {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(singleModel); err != nil {
			t.Fatal(err)
		}

		r, err := inst.NewRequest("POST", API_BOM_MODEL_END, &b)
		if err != nil {
			t.Fatal(err)
		}

		aetest.Login(u, r)
		//if err := CreateNewModel(r); err != nil {
		//	t.Error(err)
		//}
		w := httptest.NewRecorder()
		bomModelWithoutSlashAPIHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("\nError occured:\n HTML Body:%v", w.Body.String())
		}
	}
	//======================================================
	// Compare with expected list
	//======================================================
	expectedBOM = NewBOM()
	expectedBOM.Models = map[string]Model{
		"PremiumPlus": {
			Name:          "PremiumPlus",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0, GUARD.Name: 0.0},
		},
		"Sapphire": {
			Name:          "Sapphire",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0, GUARD.Name: 0.0},
		},
		"Ruby": {
			Name:          "Ruby",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0, GUARD.Name: 0.0},
		},
	}
	expectedBOM.AML = ArticleMap{KNOB.Name: KNOB, BURNER.Name: BURNER, GUARD.Name: GUARD}
	expectedBomAndArticlesAtThisStage(r, t, expectedBOM, "stage4: Created a new model ruby")

	//======================================================
	//TODO:Delete an article and test
	//======================================================

	for _, singleArticle := range []Article{GUARD} {

		r, err := inst.NewRequest("DELETE", API_BOM_ARTICLE_SLASH_END+singleArticle.Name, nil)
		if err != nil {
			t.Fatal(err)
		}

		aetest.Login(u, r)
		w := httptest.NewRecorder()
		bomArticleWithSlashAPIHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Body:%v", w.Body.String())
		}
	}

	expectedBOM = NewBOM()
	expectedBOM.Models = map[string]Model{
		"PremiumPlus": {
			Name:          "PremiumPlus",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0},
		},
		"Sapphire": {
			Name:          "Sapphire",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0},
		},
		"Ruby": {
			Name:          "Ruby",
			Unit:          "pc",
			ArticleAndQty: QtyMap{KNOB.Name: 0.0, BURNER.Name: 0.0},
		},
	}
	expectedBOM.AML = ArticleMap{KNOB.Name: KNOB, BURNER.Name: BURNER}
	expectedBomAndArticlesAtThisStage(r, t, expectedBOM, "stage5: Deleted GUARD")

	//======================================================
	//TODO:Delete a model and test
	//======================================================

	//======================================================
	//TODO:Rename a model and test
	//======================================================

	//======================================================
	//TODO:Rename an article and test
	//======================================================

	//======================================================
	//TODO:Create an article with specific name inside get request
	//======================================================

	//======================================================
	//TODO:Create a model with specific name inside get request
	//======================================================

	//======================================================
	//TODO:Create a sale invoice
	//======================================================
	ITEM_PP_2PC := SoldItem{NameRateQuantity{Name: "PremiumPlus", Rate: 22, Quantity: 2}, Model{}, ""}
	ITEM_RUBY_4PC := SoldItem{NameRateQuantity{Name: "Ruby", Rate: 16, Quantity: 4}, Model{}, ""}
	ITEMS_PP_2PC_RUBY_4PC := []SoldItem{ITEM_PP_2PC, ITEM_RUBY_4PC}
	RRKSI_PP_2PC_RUBY_4PC := RRKSaleInvoice{
		_SaleInvoice: _SaleInvoice{
			_BillFields: _BillFields{
				Number:               "111",
				GoodsValue:           44,
				GrandTotal:           55,
				TotalTax:             5,
				TotalFreight:         6,
				Remarks:              "Some remarks",
				DateValue:            time.Now(),
				JSDateValueAsSeconds: time.Now().Unix(),
			},
			CustomerName: "Django",
			Items:        ITEMS_PP_2PC_RUBY_4PC,
		},
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(RRKSI_PP_2PC_RUBY_4PC); err != nil {
		t.Fatal(err)
	}
	r, err = inst.NewRequest("POST", API_RRK_SALE_INVOICE_END, &b)
	if err != nil {
		t.Fatal(err)
	}

	aetest.Login(u, r)
	w := httptest.NewRecorder()
	rrkSaleInvoiceApiHandler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Body:%v", w.Body.String())
	}

	expectedStock := &RRKStockPos{
		_StockPos: _StockPos{
			ModelQtyMap: QtyMap{
				PREMIUM_PLUS.Name: -2,
				RUBY.Name:         -4,
				SAPPHIRE.Name:     0,
			},
			ArticleQtyMap: QtyMap{
				KNOB.Name:   0,
				BURNER.Name: 0,
			},
			DateValue: StripTimeKeepDate(time.Now()),
			DD_MMM_YY: DDMMMYYFromGoTime(time.Now()),
		},
	}

	expectedStockAtThisStage(r, t, expectedStock, "Just created a sale invoice for 2 pc PremiumPlus and 4 pc ruby")

	//======================================================
	//TODO:Delete a sale invoice
	//======================================================
	RRKSI_PP_2PC_RUBY_4PC.DD_MMM_YY = DDMMMYYFromGoTime(time.Now()) //This is done internally in the code before saving the entity, so just simulating this.
	rrksiUID := RRKSI_PP_2PC_RUBY_4PC.GetUID()
	fmt.Println("rrksiUID = ", rrksiUID)
	r, err = inst.NewRequest("DELETE", API_RRK_SALE_INVOICE_SALSH_END+rrksiUID, nil)
	if err != nil {
		t.Fatal(err)
	}

	aetest.Login(u, r)
	w = httptest.NewRecorder()
	rrkSaleInvoiceWithSalshApiHandler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Body:%v", w.Body.String())
	}

	expectedStock = &RRKStockPos{
		_StockPos: _StockPos{
			ModelQtyMap: QtyMap{
				PREMIUM_PLUS.Name: 0,
				RUBY.Name:         0,
				SAPPHIRE.Name:     0,
			},
			ArticleQtyMap: QtyMap{
				KNOB.Name:   0,
				BURNER.Name: 0,
			},
			DateValue: StripTimeKeepDate(time.Now()),
			DD_MMM_YY: DDMMMYYFromGoTime(time.Now()),
		},
	}

	expectedStockAtThisStage(r, t, expectedStock, "Just the sale invoice for 2 pc PremiumPlus and 4 pc ruby")

	//======================================================
	//TODO:Create a purchase invoice
	//======================================================

	//======================================================
	//TODO:Delete a sale invoice
	//======================================================

	//======================================================
	//TODO:Create a assembly invoice
	//======================================================

	//======================================================
	//TODO:Create a polish invoice
	//======================================================

	//======================================================
	//TODO:Create a raw-material ost invoice
	//======================================================

	//======================================================
	//TODO:Create a raw-material ist invoice
	//======================================================

	//======================================================
	//TODO:Create a finished stock ost invoice
	//======================================================

	//======================================================
	//TODO:Create a finished stock ist invoice
	//======================================================

	return
}

func IsBOMHavingSameArticleListAcrossModels(bom *BOM, t *testing.T, stage string) bool {
	if len(bom.Models) == 0 {
		return true
	}

	firstName := ""
	for k, _ := range bom.Models {
		firstName = k
	}

	firstModel := bom.Models[firstName]
	for _, singleModel := range bom.Models {
		if AreTwoArticleListsSame(singleModel, firstModel, t, stage) == false {
			return false
		}
	}
	return true
}

func AreTwoModelsEqual(m1 *Model, m2 *Model, t *testing.T, stage string) bool {

	if m1.Name != m2.Name {
		t.Errorf(fmt.Sprintf(stage+":\nModel name %v != Model name %v", m1.Name, m2.Name))
		return false
	}
	if !AreTwoArticleListsSame(m1, m2, t, stage) {
		return false
	}
	if m1.Unit != m2.Unit {
		t.Errorf(fmt.Sprintf(stage+":\nUnit %v != Unit %v", m1.Unit, m2.Unit))
		return false
	}
	if !AreTwoQtyMapsSame(m1.ArticleAndQty, m2.ArticleAndQty, t, stage) {
		return false
	}
	return true
}

func AreTwoBOMsEqual(bom1 *BOM, bom2 *BOM, t *testing.T, stage string) bool {
	for k1, m1 := range bom1.Models {
		//Compare to see if all the model names in bom1 are there in second bom2
		m2, found := bom2.Models[k1]
		if !found {
			t.Errorf(fmt.Sprintf(stage+":\nModel %v is not present in actual BOM from DS", m1.Name))
			return false
		}
		if !AreTwoModelsEqual(&m1, &m2, t, stage) {
			return false
		}
	}
	if len(bom1.Models) != len(bom2.Models) {
		t.Errorf(fmt.Sprintf(stage+":\nModel length %v is not same as model2 length: %v", len(bom1.Models), len(bom2.Models)))
		return false
	}
	if !AreTwoArticleListsSame(bom1.AML, bom2.AML, t, stage) {
		return false
	}
	return true
}

func AreTwoQtyMapsSame(map1 QtyMap, map2 QtyMap, t *testing.T, stage string) bool {
	for an, qty := range map1 {
		_, present := map2[an]
		if !present {
			t.Errorf(fmt.Sprintf(stage+":\n Article Name %v not present in other map", an))
			return false
		}
		qty2 := map2[an]
		if qty != qty2 {
			t.Errorf(fmt.Sprintf(stage+":\n Article Name %v quantity %v does not match other quantity %v present in other map\nMap1:%#v\nMap2:%#v", an, qty, qty2, map1, map2))
			return false
		}
	}
	return true
}

func TestEnableRest(t *testing.T) {
	fmt.Println("_________________________________________")
	fmt.Println("Dont forget to enable rest of the tests")
	fmt.Println("_________________________________________")
}

func AreTwoArticlesSame(a1 *Article, a2 *Article) bool {
	if a1.Name != a2.Name {
		return false
	}
	if a1.ReorderPoint != a2.ReorderPoint {
		return false
	}
	if a1.MinimumAlarmPoint != a2.MinimumAlarmPoint {
		return false
	}
	if a1.MaximumAlarmPoint != a2.MaximumAlarmPoint {
		return false
	}
	if a1.EOQ != a2.EOQ {
		return false
	}
	if a1.Unit != a2.Unit {
		return false
	}
	return true
}

func AreTwoArticleListsSame(al1 ArticleLister, al2 ArticleLister, t *testing.T, stage string) bool {
	aml1 := al1.ArticleList()
	aml2 := al2.ArticleList()

	if len(aml1) != len(aml2) {
		t.Errorf(fmt.Sprintf(stage+":\nlength of aml1 %v and aml2 %v are not same", len(aml1), len(aml2)))
		return false
	}
	for _, a1 := range aml1 {
		found := false
		for _, a2 := range aml2 {
			if a1 == a2 {
				found = true
			}
		}
		if found == false {
			return false
		}
	}
	return true
}

func AreTwoStocksSame(s1 *RRKStockPos, s2 *RRKStockPos, t *testing.T, stage string) bool {
	if len(s1.ArticleQtyMap) != len(s2.ArticleQtyMap) {
		t.Errorf(fmt.Sprintf(stage+":\nlength of s1 %v and s2 %v are not same", len(s1.ArticleQtyMap), len(s2.ArticleQtyMap)))
		return false
	}
	for a1, q1 := range s1.ArticleQtyMap {
		q2, ok := s2.ArticleQtyMap[a1]
		if !ok {
			t.Errorf(fmt.Sprintf(stage+":\n%s article not present in s2", a1))
			return false
		}
		if q1 != q2 {
			t.Errorf(fmt.Sprintf(stage+":\n%s article's quantity is not same in two stocks", a1))
			return false
		}
	}

	if len(s1.ModelQtyMap) != len(s2.ModelQtyMap) {
		t.Errorf(fmt.Sprintf(stage+":\nlength of s1 %v and s2 %v are not same", len(s1.ModelQtyMap), len(s2.ModelQtyMap)))
		return false
	}
	for a1, q1 := range s1.ModelQtyMap {
		q2, ok := s2.ModelQtyMap[a1]
		if !ok {
			t.Errorf(fmt.Sprintf(stage+":\n%s model is not present in s2", a1))
			return false
		}
		if q1 != q2 {
			t.Errorf(fmt.Sprintf(stage+":\n%s model's quantity is not same in two stocks", a1))
			return false
		}
	}
	if s1.DD_MMM_YY != s2.DD_MMM_YY {
		t.Errorf(fmt.Sprintf(stage+":\nStock dates are not same %s != %s", s1.DD_MMM_YY, s2.DD_MMM_YY))
		return false
	}
	return true
}

var (
	INITIAL_CASH_TXS = CashTxsCluster{
		DateOfTransactionAsUnixTime: time.Now().Unix(),
		OpeningBalance:              2111,
		Items: []CashTransaction{
			{
				Nature:         "Unsettled Advance",
				BillNumber:     "1",
				Amount:         -111,
				Description:    "my description",
				DateAsUnixTime: time.Now().Unix(),
			},
		},
	}

	DEBIT_2000 = CashTxsCluster{
		DateOfTransactionAsUnixTime: time.Now().Unix(),
		Items: []CashTransaction{
			{
				Nature:         "Spent",
				BillNumber:     "10",
				Amount:         -2000,
				Description:    "my description",
				DateAsUnixTime: time.Now().Unix(),
			},
		},
	}
	DEBIT_100 = CashTxsCluster{
		DateOfTransactionAsUnixTime: time.Now().Unix(),
		Items: []CashTransaction{
			{
				Nature:         "Spent",
				BillNumber:     "10",
				Amount:         -100,
				Description:    "my description",
				DateAsUnixTime: time.Now().Unix(),
			},
		},
	}
	DEBIT_100_100 = CashTxsCluster{
		DateOfTransactionAsUnixTime: time.Now().Unix(),
		Items: []CashTransaction{
			{
				Nature:         "Spent",
				BillNumber:     "10",
				Amount:         -100,
				Description:    "my description",
				DateAsUnixTime: time.Now().Unix(),
			},
			{
				Nature:         "Spent",
				BillNumber:     "11",
				Amount:         -100,
				Description:    "my description for another 100",
				DateAsUnixTime: time.Now().Unix(),
			},
		},
	}

	RECEIVED_100 = CashTxsCluster{
		DateOfTransactionAsUnixTime: time.Now().Unix(),
		Items: []CashTransaction{
			{
				Nature:         "Received",
				BillNumber:     "10",
				Amount:         100,
				Description:    "my description",
				DateAsUnixTime: time.Now().Unix(),
			},
		},
	}
	RECEIVED_100_100 = CashTxsCluster{
		DateOfTransactionAsUnixTime: time.Now().Unix(),
		Items: []CashTransaction{
			{
				Nature:         "Received",
				BillNumber:     "10",
				Amount:         100,
				Description:    "my description",
				DateAsUnixTime: time.Now().Unix(),
			},
			{
				Nature:         "Received",
				BillNumber:     "10",
				Amount:         100,
				Description:    "my description for another 100",
				DateAsUnixTime: time.Now().Unix(),
			},
		},
	}

	CASH_TEST = []struct {
		cashTxs                CashTxsCluster
		expectedClosingBalance int64
	}{
		{INITIAL_CASH_TXS, 2000},
		{DEBIT_100, 1900},
		{DEBIT_100, 1800},
		{DEBIT_100_100, 1600},
		{DEBIT_100_100, 1400},
		{DEBIT_100, 1300},
		{RECEIVED_100_100, 1500},
		{RECEIVED_100_100, 1700},
		{DEBIT_2000, -300},
	}
)

func noTestRRKUnsettledAdvance(t *testing.T) {

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	r, err := inst.NewRequest("POST", "/api/rrkCashBookStoreAndEmailApi", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	u := &user.User{Email: "CurrentAETestUser@test.com"}
	aetest.Login(u, r)

	//TODO
	//======================================================
	// Initialize RRK and store its value
	//======================================================
	//if err := CashBookStoreAndEmailApi(&INITIAL_CASH_TXS, r, RRKCashRollingCounterKey, RRKGetPreviousCashRollingCounter, rrkSaveUnsettledAdvanceEntryInDataStore, "RRKDC"); err != nil {
	//	t.Errorf("Error: %v", err)
	//}
	//oldRrkCashRollingCounter, err := RRKGetPreviousCashRollingCounter(r)
	//if err != nil {
	//	t.Errorf("Error: %v", err)
	//}

}

func noTestGZBCashSettlement(t *testing.T) {

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	r, err := inst.NewRequest("POST", "/api/gzbCashBookStoreAndEmailApi", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	u := &user.User{Email: "CurrentAETestUser@test.com"}
	aetest.Login(u, r)

	//======================================================
	// Initialize RRK and store its value
	//======================================================
	if err := CashBookStoreAndEmailApi(&INITIAL_CASH_TXS, r, RRKCashRollingCounterKey, RRKGetPreviousCashRollingCounter, rrkSaveUnsettledAdvanceEntryInDataStore, "RRKDC"); err != nil {
		t.Errorf("Error: %v", err)
	}
	oldRrkCashRollingCounter, err := RRKGetPreviousCashRollingCounter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, cashTestCase := range CASH_TEST {
		if err := CashBookStoreAndEmailApi(&cashTestCase.cashTxs, r, GZBCashRollingCounterKey, GZBGetPreviousCashRollingCounter, gzbSaveUnsettledAdvanceEntryInDataStore, "GZBDC"); err != nil {
			t.Errorf("Error: %v", err)
		}

		cashRollingCounter, err := GZBGetPreviousCashRollingCounter(r)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if cashTestCase.expectedClosingBalance != cashRollingCounter.Amount {
			t.Errorf("Expected: %v; Got: %v", cashTestCase.expectedClosingBalance, cashRollingCounter.Amount)
		}
	}

	newRrkCashRollingCounter, err := RRKGetPreviousCashRollingCounter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if newRrkCashRollingCounter.Amount != oldRrkCashRollingCounter.Amount {
		t.Errorf("GZB Cash settlement is affecting RRK cash settlement. This is fatal!")
	}
}

func noTestRRkCashSettlement(t *testing.T) {

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	r, err := inst.NewRequest("POST", "/api/rrkCashBookStoreAndEmailApi", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	u := &user.User{Email: "CurrentAETestUser@test.com"}
	aetest.Login(u, r)

	//======================================================
	// Initialize Gzb and store its value
	//======================================================

	if err := CashBookStoreAndEmailApi(&INITIAL_CASH_TXS, r, GZBCashRollingCounterKey, GZBGetPreviousCashRollingCounter, gzbSaveUnsettledAdvanceEntryInDataStore, "GZBDC"); err != nil {
		t.Errorf("Error: %v", err)
	}

	oldGZBCashRollingCounter, err := GZBGetPreviousCashRollingCounter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, cashTestCase := range CASH_TEST {
		if err := CashBookStoreAndEmailApi(&cashTestCase.cashTxs, r, RRKCashRollingCounterKey, RRKGetPreviousCashRollingCounter, rrkSaveUnsettledAdvanceEntryInDataStore, "RRKDC"); err != nil {
			t.Errorf("Error: %v", err)
		}

		cashRollingCounter, err := RRKGetPreviousCashRollingCounter(r)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if cashTestCase.expectedClosingBalance != cashRollingCounter.Amount {
			t.Errorf("Expected: %v; Got: %v", cashTestCase.expectedClosingBalance, cashRollingCounter.Amount)
		}
	}
	newGZBCashRollingCounter, err := GZBGetPreviousCashRollingCounter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if newGZBCashRollingCounter.Amount != oldGZBCashRollingCounter.Amount {
		t.Errorf("Rrk cash settlement is affecting gzb cash")
	}

}
