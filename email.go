package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	gomail "gopkg.in/gomail.v2"
)

const welcomeEmailTemplate = `Hi,

Thanks for signing up for Pijun.

To verify your email, click the link below:
https://e.pijun.io/verify-email?t={{.Token}}

I hope you enjoy using Pijun as much I enjoyed creating it.

Best,
Arash

If you didn't sign up for Pijun, sorry for the inconvenience. Somebody signed up and mistakenly used your email address. Click the link below to remove your email address from this account:
https://e.pijun.io/disavow-email?t={{.Token}}
`

var emailConfiguration = struct {
	smtpUser     string
	smtpPassword string
	smtpServer   string
	smtpPort     int
}{}

// const gmailUser = `noreply@pijun.io`
// const gmailPassword = `whatisagoodpasswordfor229nvbzzaaldfppqq`
// const gmailSMTPServer = `smtp.gmail.com`
// const gmailSMTPPort = 587

func sendEmail(to, from, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	dialer := gomail.NewDialer(emailConfiguration.smtpServer,
		emailConfiguration.smtpPort,
		emailConfiguration.smtpUser,
		emailConfiguration.smtpPassword)

	return dialer.DialAndSend(msg)
}

func sendVerificationEmail(token, email string) error {
	tmpl, err := template.New("").Parse(welcomeEmailTemplate)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	tmpl.Execute(buf, struct{ Token string }{Token: token})
	return sendEmail(email, "Pijun <noreply@pijun.io>", "Pijun: Email Verification", buf.String())
}

// verifyEmailHandler handles POST /email-verifications
func verifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Token string `json:"token"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendBadReq(w, "unable to parse POST body")
		return
	}

	if body.Token == "" {
		sendBadReqCode(w, "Missing verification token", errorBadRequest)
	}

	evtr, err := rs.EmailVerificationTokenRecord(body.Token)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	if evtr == nil {
		sendBadReqCode(w, "Invalid token", errorMissingVerificationToken)
		return
	}

	// add the email to the user, then delete the verification
	err = rs.VerifyEmail(evtr.Email, evtr.UserID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, nil)
}

// disavowEmailHandler handles DELETE /email-verifications
func disavowEmailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	if token == "" {
		sendBadReqCode(w, "Invalid token", errorMissingVerificationToken)
		return
	}

	err := rs.DisavowEmail(token)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, nil)
}
