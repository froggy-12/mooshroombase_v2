package smtpconfigs

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/froggy-12/mooshroombase_v2/configs"
)

type VerificationEmailData struct {
	Code string
}

func SendVerificationEmail(emailTo string, code string) error {
	smtpServer := configs.Configs.SMTPConfigurations.SMTPServerAddress
	smtpPort := configs.Configs.SMTPConfigurations.SMTPServerPORT

	from := configs.Configs.SMTPConfigurations.SMTPEmailAddrss
	password := configs.Configs.SMTPConfigurations.SMTPEmailPassword
	subject := "Email Verification"

	data := VerificationEmailData{
		Code: code, // convert int to string
	}

	tmpl := template.Must(template.New("email").Parse(`	<!DOCTYPE html>
<html lang="en">

<head>

  <meta charset="UTF-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <title>Verification Token</title>
</head>

<body>
  <div>
    <h1>Lets Verify your account 😊😊</h1>
    <h1>Your Token is <span>{{ .Code }}</span></h1>
    <p>Have a nice day</p>
  </div>
</body>

</html>`))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return err
	}

	msg := "To: " + emailTo + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		buf.String()

	auth := smtp.PlainAuth(from, from, password, smtpServer)

	err = smtp.SendMail(smtpServer+":"+smtpPort, auth, from, []string{emailTo}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}

func SendEmailWithAnything(EmailSubject, emailTo, emailBody string) error {
	smtpServer := configs.Configs.SMTPConfigurations.SMTPServerAddress
	smtpPort := configs.Configs.SMTPConfigurations.SMTPServerPORT

	from := configs.Configs.SMTPConfigurations.SMTPEmailAddrss
	password := configs.Configs.SMTPConfigurations.SMTPEmailPassword

	msg := "To: " + emailTo + "\r\n" +
		"Subject: " + EmailSubject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		emailBody

	auth := smtp.PlainAuth(from, from, password, smtpServer)

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, from, []string{emailTo}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
