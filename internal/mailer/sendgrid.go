package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any) error {
	from := mail.NewEmail(fromUser, m.fromEmail)
	to := mail.NewEmail(username, email)

	tmp, err := template.ParseFS(FS, templateFile)
	if err != nil {
		log.Printf("Could not find template")
		return err
	}

	subject := new(bytes.Buffer)
	err = tmp.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmp.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	for i := 0; i < maxRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		if response.StatusCode != 200 {
			fmt.Println(response.Body)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		fmt.Println(i)
		fmt.Println(response.StatusCode)
		fmt.Println("Email sent")
		return nil
	}

	return errors.New("failed to send email to user")
}
