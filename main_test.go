package sewpulse

import (
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
