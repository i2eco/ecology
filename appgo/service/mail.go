package service

import (
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

var Mailer *mailer

type mailer struct {
	*gomail.Dialer
}

func InitMailer() {
	Mailer = &mailer{
		Dialer: gomail.NewDialer(viper.GetString("email.host"), viper.GetInt("email.port"), viper.GetString("email.username"), viper.GetString("email.password")),
	}
}

func (m *mailer) Send(subject, to string, html string, attachment string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", viper.GetString("email.from"))
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", html)
	if attachment != "" {
		msg.Attach(attachment)
	}
	// Send the email to Bob, Cora and Dan.
	if err := m.Dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

//
//
//
////@param            conf            邮箱配置
////@param            subject         邮件主题
////@param            email           收件人
////@param            body            邮件内容
//func SendMail(conf *conf.SmtpConf, subject, email string, body string) error {
//	msg := &mail.Message{
//		Header: mail.Header{
//			"From":         {conf.FormUserName},
//			"To":           {email},
//			"Reply-To":     {conf.ReplyUserName},
//			"Subject":      {subject},
//			"Content-Type": {"text/html"},
//		},
//		Body: strings.NewReader(body),
//	}
//	port := conf.SmtpPort
//	host := conf.SmtpHost
//	username := conf.FormUserName
//	password := conf.SmtpPassword
//	m := mailer.NewMailer(host, username, password, port)
//	return m.Send(msg)
//}
