package mailer

import (
	"fmt"
	"time"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func New(host string, port int, user, pass, from string) *Mailer {
	return &Mailer{
		dialer: gomail.NewDialer(host, port, user, pass),
		from:   from,
	}
}

func (m *Mailer) SendBorrowConfirmation(toEmail, name, bookName string, dueDate time.Time) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", "Book borrowed: "+bookName)
	msg.SetBody("text/html", fmt.Sprintf(
		"<h2>Hi, %s!</h2><p>You borrowed <b>%s</b>.</p><p>Return before <b>%s</b>.</p>",
		name, bookName, dueDate.Format("2006-01-02"),
	))
	return m.dialer.DialAndSend(msg)
}

func (m *Mailer) SendReturnConfirmation(toEmail, name, bookName string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", "Book returned: "+bookName)
	msg.SetBody("text/html", fmt.Sprintf(
		"<h2>Hi, %s!</h2><p>Thank you for returning <b>%s</b>.</p>",
		name, bookName,
	))
	return m.dialer.DialAndSend(msg)
}

func (m *Mailer) SendReservationConfirmation(toEmail, name, bookName string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", "Reservation confirmed: "+bookName)
	msg.SetBody("text/html", fmt.Sprintf(
		"<h2>Hi, %s!</h2><p>You reserved <b>%s</b>.</p>",
		name, bookName,
	))
	return m.dialer.DialAndSend(msg)
}