package mailer

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func New(host string, port int, user, pass, from string) *Mailer {
	d := gomail.NewDialer(host, port, user, pass)
	return &Mailer{dialer: d, from: from}
}

func (m *Mailer) SendWelcome(toEmail, name, verificationLink string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", "Welcome to E-Library")
	body := fmt.Sprintf(`
		<h2>Hi, %s!</h2>
		<p>Welcome to E-Library. Please verify your email by clicking the link below:</p>
		<p><a href="%s">Verify my email</a></p>
		<p>If you did not register, just ignore this message.</p>
	`, name, verificationLink)
	msg.SetBody("text/html", body)
	return m.dialer.DialAndSend(msg)
}
