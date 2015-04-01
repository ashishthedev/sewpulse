package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func YYYYMMMDDFromUnixTime(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("2006-Jan-02")
}

func DDMMYYFromUnixTime(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("02-Jan-06")
}

func YYYYMMMDDFromGoTime(goTime time.Time) string {
	return goTime.Format("2006-Jan-02")
}

func DDMMYYFromGoTime(goTime time.Time) string {
	return goTime.Format("02-Jan-06")
}

func myDebug(r *http.Request, s string) {
	//line := "________________________"
	//s1 := "\n" + line + "\n" + s + "\n" + line + "\n\n"
	s1 := "\n|||||||||||||||||" + s
	c := appengine.NewContext(r)
	c.Debugf(s1)
	return
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func BranchName(r *http.Request) string {
	if appengine.IsDevAppServer() {
		return "localhost"
	}
	host := r.URL.Host
	if host == "" {
		panic("Host is empty. A lot depends upon the host of the URL")
	}
	s := strings.Split(strings.ToLower(host), ".")[0]
	if len(s) > 0 {
		return s
	}
	return host
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

func _SEWNewKey(kind string, stringId string, numericID int64, r *http.Request) *datastore.Key {
	c := appengine.NewContext(r)
	//This is the only place datastore.NewKey should appear as we are creating a silo for
	//each hosted version. Any operation done in this app should only be limited
	//to only that silo. For ex/- demo.sew.appspot.com should only effect "demo"
	//silo and not the live version data.

	ancestorKey := datastore.NewKey(c, "ANCESTOR_KEY", BranchName(r), 0, nil)
	return datastore.NewKey(c, kind, stringId, numericID, ancestorKey)
}

func GZB_SEWNewKey(kind string, stringId string, numericID int64, r *http.Request) *datastore.Key {
	GZB_PREFIX := "GZB_"
	return _SEWNewKey(kind, GZB_PREFIX+stringId, numericID, r)
}

func RRK_SEWNewKey(kind string, stringId string, numericID int64, r *http.Request) *datastore.Key {
	RRK_PREFIX := ""
	return _SEWNewKey(kind, RRK_PREFIX+stringId, numericID, r)
}

func CMN_SEWNewKey(kind string, stringId string, numericID int64, r *http.Request) *datastore.Key {
	COMMON_PREFIX := "CMN_"
	return _SEWNewKey(kind, COMMON_PREFIX+stringId, numericID, r)
}

func toJson(i interface{}) []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("ToJson() %v ==> %v", i, data))
	return data
}

func fromJson(v []byte, vv interface{}) error {
	err := json.Unmarshal(v, vv)
	fmt.Println(fmt.Sprintf("FromJson() %v ==> %v", v, vv))
	return err
}
