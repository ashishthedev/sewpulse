package sewpulse

import (
	"appengine"
	"appengine/user"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func YYYYMMMDDFromUTC(utc int64) string {
	return time.Unix(utc/1000, 0).Format("2006-Jan-02")
}

func DDMMYYFromUTC(utc int64) string {
	return time.Unix(utc/1000, 0).Format("02-Jan-06")
}

func myDebug(r *http.Request, s string) {
	c := appengine.NewContext(r)
	c.Debugf(s)
	return
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func IsLocalHostedOrOnDevBranch(r *http.Request) bool {
	if appengine.IsDevAppServer() {
		return true
	}
	if strings.ToLower(r.URL.Host[:4]) == "dev." {
		return true
	}
	return false
}

func IsUserAdminOrDev(r *http.Request) bool {
	if appengine.IsDevAppServer() {
		return true
	}
	c := appengine.NewContext(r)
	if u := user.Current(c); u != nil {
		return u.Admin
	}
	return false
}

var (
	DaysMsg = map[string]string{
		"ON_TIME":     "Submitted on same day",
		"ONE_DAY_OLD": "Submitted after 1 day",
		"X_DAYS_OLD":  "Submitted after %d days",
	}
)

func LogMsgShownForLogTime(logTime time.Time, nowTime time.Time) string {
	l := logTime
	n := nowTime
	newLogTime := time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)
	newNow := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	duration := newNow.Sub(newLogTime)
	hrsDiff := int64(duration.Hours())

	if hrsDiff == 0 {
		return DaysMsg["ON_TIME"]
	} else if hrsDiff == 24 {
		return DaysMsg["ONE_DAY_OLD"]
	} else {
		noOfDays := hrsDiff / 24
		return fmt.Sprintf(DaysMsg["X_DAYS_OLD"], noOfDays)
	}
	panic("Should not reach here")
}
