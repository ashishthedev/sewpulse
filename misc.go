package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func UnixTimeFromGoTime(goTime time.Time) int64 {
	return goTime.Unix()
}

func GoTimeFromUnixTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}
func YYYYMMMDDFromUnixTime(unixTime int64) string {
	return GoTimeFromUnixTime(unixTime).Format("2006-Jan-02")
}

func DDMMMYYFromUnixTime(unixTime int64) string {
	return GoTimeFromUnixTime(unixTime).Format("02-Jan-06")
}

func YYYYMMMDDFromGoTime(goTime time.Time) string {
	return goTime.Format("2006-Jan-02")
}

func DDMMMYYFromGoTime(goTime time.Time) string {
	return goTime.Format("02-Jan-06")
}

func DDMMMYYToGoTime(DD_MMM_YY string) (time.Time, error) {
	return time.Parse("02-Jan-06", DD_MMM_YY)
}

func ClockTimeFromoGoTime(goTime time.Time) string {
	return goTime.Format("3:04pm (MST)")
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

func WriteJson(w *http.ResponseWriter, i interface{}) error {
	return json.NewEncoder(*w).Encode(i)
}

func fromJson(v []byte, vv interface{}) error {
	return json.Unmarshal(v, vv)
}

func JsonToStruct(data *string, v interface{}, r *http.Request) error {
	b := bytes.NewBuffer([]byte(*data))
	if err := json.NewDecoder(b).Decode(v); err != nil {
		myDebug(r, "Error in JsonToStruct():"+err.Error())
		return err
	}
	return nil
}

func StructToJson(v interface{}, r *http.Request) (*string, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(v); err != nil {
		myDebug(r, "Error in StructToJson():"+err.Error())
		return nil, err
	}

	str := string(b.Bytes())
	return &str, nil
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

func BOD(dt time.Time) time.Time {
	//Strip the time= you get end of the day time.
	return StripTimeKeepDate(dt)
}

func EOD(dt time.Time) time.Time {
	//Strip the time, add a day, subtract a second = you get end of the day time.
	return StripTimeKeepDate(dt).Add(1*24*time.Hour - time.Second)
}

func StripTimeKeepDate(dt time.Time) time.Time {
	return time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
}

func spf(s string, a ...interface{}) string {
	return fmt.Sprintf(s, a...)
}

func myDebug(r *http.Request, s string, a ...interface{}) {
	appengine.NewContext(r).Debugf("\n>>>>|" + spf(s, a...) + "|<<<<")
	return
}

func logErr(r *http.Request, err error, fnName string) error {
	s := "Error returned from " + fnName + "() :\n" + err.Error()
	appengine.NewContext(r).Errorf("\n>>>>|" + s + "|<<<<")
	return err
}

func DecodeBodyToStruct(r *http.Request, s interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		return err
	}
	return nil
}

func ServeError(w http.ResponseWriter, err error, httpStatusCode int, r *http.Request) {
	//Report error through mail.
	c := appengine.NewContext(r)
	ver := appengine.VersionID(c)
	appID := appengine.AppID(c)
	lines := []string{
		DDMMMYYFromGoTime(time.Now().Local()),
		ClockTimeFromoGoTime(time.Now().Local()),
		ver,
		err.Error(),
		"while accessing: ",
		r.URL.Path,
	}
	errMsg := ""
	for _, x := range lines {
		errMsg += "<br><br>" + x
	}
	emailSubject := appID + " faced error in version: " + ver + " | " + time.Now().Local().String()
	SEWReportErrorThroughMail(r, emailSubject, errMsg) //Eat any errors here.

	//Report error in logs.

	appengine.NewContext(r).Errorf("\n>>>>|" + err.Error() + "|<<<<")

	//Report error on ResponseWriter.
	http.Error(w, err.Error(), httpStatusCode)
	return
}
