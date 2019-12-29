package mailer

import (
	"git.sr.ht/~humaid/nevernote/modules/settings"
	"net/smtp"
)

func EmailCode(to string, code string) (err error) {
	from := settings.Config.EmailAddress
	message := "From: <" + from + ">\n" +
		"To: <" + to + ">\n" +
		"Subject: Nevernote login code\n\n" +
		"Hello!\nYour login code is " + code + "\n\n" +
		"Ignore this message if you have not requested a login.\n\n\n" +
		"- Nevernote\nThis message is sent from an unmonitored inbox."

	err = smtp.SendMail(settings.Config.EmailSMTPServer,
		smtp.PlainAuth("", from, settings.Config.EmailPassword, "smtp.migadu.com"),
		from, []string{to}, []byte(message))

	return err
}
