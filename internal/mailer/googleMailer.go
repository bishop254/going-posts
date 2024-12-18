package mailer

import "net/smtp"

func Send(templateFile, username, email string, data any) error {
	from := "kc@gmail.com"
	pass := "jmwi myie ujxc tnql"
	to := email

	msg := "From: " + from + "To: " + to + "Body: "

	err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, pass, "smtp.gmail.com"), from, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
