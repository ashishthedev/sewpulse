package sewpulse

import (
	"appengine"
	"appengine/mail"
	"appengine/user"
	"net/http"
)

func SendSEWMail(r *http.Request, subject string, finalHTML string) error {
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
		Subject:  subject,
		HTMLBody: finalHTML,
	}

	return mail.Send(c, msg)
}
