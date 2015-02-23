package sewpulse

import (
	"appengine/aetest"
	"appengine/user"
	"fmt"
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

func TestRRKUnsettledAdvance(t *testing.T) {

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

func TestGZBCashSettlement(t *testing.T) {

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

func TestRRkCashSettlement(t *testing.T) {

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
