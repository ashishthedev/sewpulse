package sewpulse

import (
	"appengine"
	"appengine/mail"
	"appengine/user"
	"net/http"
)

var BCC_ADDR = Reverse("moc.liamg@vedehthsihsa")
var ATD = Reverse("moc.liamg@vedehthsihsa")

func SendSEWMail(r *http.Request, subject string, finalHTML string) error {
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
		Bcc:      []string{BCC_ADDR},
		Subject:  subject + "[SEW]",
		HTMLBody: finalHTML,
	}

	return mail.Send(c, msg)
}

func SEWReportErrorThroughMail(r *http.Request, subject string, finalHTML string) error {
	c := appengine.NewContext(r)
	msg := &mail.Message{
		Sender:   ATD,
		To:       []string{ATD},
		Subject:  subject + " [SEW][SEWErr]",
		HTMLBody: finalHTML,
	}

	return mail.Send(c, msg)
}
