package service

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/go-mail/mail/v2"
)

type MailSender struct {
	smtpHost string
	smtpPort int
	smtpUser string
	smtpPass string
}

func InitMailSender() (*MailSender, error) {
	// Получение параметров из переменных окружения
	host := os.Getenv("SMTP_HOST")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("Не валидный порт SMTP %s", err.Error())
	}
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")

	sender := MailSender{}
	sender.smtpHost = host
	sender.smtpPort = port
	sender.smtpUser = user
	sender.smtpPass = password
	return &sender, nil
}

// Сообщение для отправки по почте
func (sender *MailSender) createMessage(to string, subject string, body string) *mail.Message {
	m := mail.NewMessage()
	m.SetHeader("From", sender.smtpUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	return m
}

// SMTP-диалог
func (sender *MailSender) createDialer() *mail.Dialer {
	d := mail.NewDialer(sender.smtpHost, sender.smtpPort, sender.smtpUser, sender.smtpPass)
	d.TLSConfig = &tls.Config{
		ServerName:         sender.smtpHost,
		InsecureSkipVerify: false, // Не отключать проверку сертификата
	}
	return d
}

// Отправка письма
func (sender *MailSender) sendEmail(m *mail.Message) error {
	// Настройка подключения
	d := sender.createDialer()

	// Отправка
	if err := d.DialAndSend(m); err != nil {
		log.Printf("SMTP error: %v", err)
		return fmt.Errorf("email sending failed")
	}

	return nil
}

// Уведомление о выпуске карты
func (sender *MailSender) SendCVVEmailMessage(emailTo, cvv string) error {
	log.Println("Уведомление о выпуске карты")
	// Создание контента
	content := fmt.Sprintf(`
        <h1>Ваша карта создана!</h1>
        <p>CVV: <strong>%s</strong></p>
        <small>Это автоматическое уведомление</small>
    `, cvv)
	// Подготовка сообщения
	m := sender.createMessage(emailTo, "Ваша карта создана", content)

	// Отправка
	if err := sender.sendEmail(m); err != nil {
		log.Printf("Error send mail %s", err.Error())
		return err
	}

	log.Printf("Email sent to %s", emailTo)
	log.Println()

	return nil
}

// Уведомление о оплате через email
func (sender *MailSender) SendEmailMessage(emailTo string, amount float64) error {
	log.Println("Уведомление о оплате через email")
	// Создание контента
	content := fmt.Sprintf(`
        <h1>Спасибо за оплату!</h1>
        <p>Сумма: <strong>%.2f RUB</strong></p>
        <small>Это автоматическое уведомление</small>
    `, amount)
	// Подготовка сообщения
	m := sender.createMessage(emailTo, "Платеж успешно проведен", content)

	// Отправка
	if err := sender.sendEmail(m); err != nil {
		log.Printf("Error send mail %s", err.Error())
		return err
	}

	log.Printf("Email sent to %s", emailTo)
	log.Println()

	return nil
}
