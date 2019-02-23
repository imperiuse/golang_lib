package email

import (
	"crypto/tls"
	"fmt"
	"github.com/imperiuse/golib/concat"
	"log"
	"net"
	"net/mail"
	"net/smtp"
)

// MailBean - all settings for email package in Bean-like struct
type MailBean struct {
	from mail.Address   // mail.Address{"FromMe", "from@yandex.ru"}
	to   []mail.Address // mail.Address{"ToYou", "to@yandex.ru"}
	Subj string
	Credentials
	EnableNotify bool
}

// Credentials - credentials
type Credentials struct {
	Username   string // "from@yandex.ru"
	Password   string // "token app"
	ServerName string // "smtp.yandex.ru:465"
	Identity   string // ""
}

// SetFromAndToEmailAddresses - set from and to email addresses
func (m *MailBean) SetFromAndToEmailAddresses(from mail.Address, to []mail.Address) {
	m.from = from
	m.to = to
}

// SendEmailByDefaultTemplate -  send email with default template @see email.emailTemplate const
func (m *MailBean) SendEmails(body string) error {
	if m.EnableNotify {
		for _, to := range m.to {
			if err := sendEmail(m.from, to, m.Subj, body, m.Credentials); err != nil {
				return err
			}
		}
	}
	return nil
}

// sendEmail - send one email
func sendEmail(from, to mail.Address, subj, body string, c Credentials) (err error) {
	// Setup headers
	headers := make(map[string]string)
	headers["from"] = from.String()
	headers["to"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		concat.Strings(message, fmt.Sprintf("%s: %s\r\n", k, v))
	}
	concat.StringsMulti(message, "\r\n", body)

	// Connect to the SMTP Server
	host, _, _ := net.SplitHostPort(c.ServerName)

	auth := smtp.PlainAuth(c.Identity, c.Username, c.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", c.ServerName, tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = client.Auth(auth); err != nil {
		log.Panic(err)
	}

	// to && from
	if err = client.Mail(from.Address); err != nil {
		log.Panic(err)
	}

	if err = client.Rcpt(to.Address); err != nil {
		log.Panic(err)
	}

	// Data
	w, err := client.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	err = client.Quit()

	return err
}
