package main

import (
	"bytes"
	"html/template"
	"log"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) sendToSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	formattedHTMLMmessage, err := m.buildHTMLMessage(msg)

	if err != nil {
		log.Panic("Error while parsing html file!")
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)

	if err != nil {
		log.Panic("Error while creating Plain Text Message")
		return err
	}

	// define mail server
	srv := mail.NewSMTPClient()
	srv.Host = m.Host
	srv.Port = m.Port
	srv.Username = m.Username
	srv.Encryption = m.getEncription(m.Encryption)
	srv.KeepAlive = false
	srv.ConnectTimeout = 10 * time.Second
	srv.SendTimeout = 10 * time.Second

	smtpClient, err := srv.Connect()

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMessage).
		AddAlternative(mail.TextHTML, formattedHTMLMmessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}
	err = email.Send(smtpClient)

	if err != nil {
		log.Panic("Error while sending Message")
		return err
	}
	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {

	templateToRender := "./templates/mail.html.gohtml"
	t, error := template.New("email-html").ParseFiles(templateToRender)

	if error != nil {
		log.Panic("Error while parsing html file!")
		return "", error
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Panic("Error While Executing Template of message body")
		return "", err
	}

	htmlMessage := tpl.String()

	return htmlMessage, nil
}
func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {

	templateToRender := "./templates/mail.plain.gohtml"
	t, error := template.New("email-plain").ParseFiles(templateToRender)

	if error != nil {
		log.Panic("Error while parsing html file!")
		return "", error
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Panic("Error While Executing Template of message body")
		return "", err
	}

	plainMessage := tpl.String()
	return plainMessage, nil
}

func (m *Mail) inlineCSS(msg string) (string, error) {

	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(msg, &options)

	if err != nil {
		log.Panic("Error while executing premailer")
		return "", err
	}

	html, err := prem.Transform()

	if err != nil {
		log.Panic("Error while creating html")
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncription(s string) mail.Encryption {

	switch s {
	case "ssl":
		return mail.EncryptionSTARTTLS
	case "tls":
		return mail.EncryptionSSLTLS
	case "", "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}

}
