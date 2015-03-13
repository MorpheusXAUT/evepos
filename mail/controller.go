package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/morpheusxaut/evepos/database"
	"github.com/morpheusxaut/evepos/misc"
	"github.com/morpheusxaut/evepos/models"

	"github.com/dustin/go-humanize"
)

// Controller handles sending mails via a given SMTP server as required by the app
type Controller struct {
	config   *misc.Configuration
	database database.Connection
}

// SetupMailController initialises a new mail controller
func SetupMailController(conf *misc.Configuration, db database.Connection) *Controller {
	controller := &Controller{
		config:   conf,
		database: db,
	}

	return controller
}

// SendPasswordReset sends a verification email with a password reset link to the user's given email address
func (controller *Controller) SendPasswordReset(username string, email string, verification string) error {
	templates := template.Must(template.New("").Funcs(controller.TemplateFunctions()).ParseFiles("app/templates/passwordreset.html"))

	data := make(map[string]interface{})
	data["username"] = username
	data["verificationLink"] = fmt.Sprintf("%s/login/reset/verify?email=%s&username=%s&verification=%s", controller.config.HTTPPublicURL, email, username, verification)

	var buf bytes.Buffer

	err := templates.ExecuteTemplate(&buf, "passwordreset", data)
	if err != nil {
		return err
	}

	return controller.SendEmail(email, "evepos - Password reset", buf.String(), fmt.Sprintf("Please use the following link to reset your password: %s/login/reset/verify?email=%s&username=%s&verification=%s", controller.config.HTTPPublicURL, email, username, verification))
}

func (controller *Controller) SendFuelReminder(username string, email string, poses []*models.POS) error {
	templates := template.Must(template.New("").Funcs(controller.TemplateFunctions()).ParseFiles("app/templates/fuelreminder.html"))

	data := make(map[string]interface{})
	data["username"] = username
	data["poses"] = poses

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "fuelreminder", data)
	if err != nil {
		return err
	}

	return controller.SendEmail(email, "evepos - POS fuel reminder", buf.String(), fmt.Sprintf("POS fuel reminder. Check %s/poses", controller.config.HTTPPublicURL))
}

// SendEmail properly formats an email with the given data and sends it via a SMTP client
func (controller *Controller) SendEmail(email string, subject string, message string, plainMessage string) error {
	smtpHostname, _, err := net.SplitHostPort(controller.config.SMTPHost)
	if err != nil {
		return err
	}

	appURL, err := url.Parse(controller.config.HTTPPublicURL)
	if err != nil {
		return err
	}

	appHostname := appURL.Host
	if strings.Contains(appHostname, ":") {
		appHostname, _, err = net.SplitHostPort(appHostname)
		if err != nil {
			return err
		}
	}

	auth := smtp.PlainAuth("", controller.config.SMTPUser, controller.config.SMTPPassword, smtpHostname)

	smtpClient, err := smtp.Dial(controller.config.SMTPHost)
	if err != nil {
		return err
	}

	err = smtpClient.Hello(appHostname)
	if err != nil {
		return err
	}

	if controller.config.SMTPStartTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         smtpHostname,
		}

		err = smtpClient.StartTLS(tlsConfig)
		if err != nil {
			return err
		}
	}

	err = smtpClient.Auth(auth)
	if err != nil {
		return err
	}

	err = smtpClient.Mail(controller.config.SMTPSender)
	if err != nil {
		return err
	}

	err = smtpClient.Rcpt(email)
	if err != nil {
		return err
	}

	wc, err := smtpClient.Data()
	if err != nil {
		return err
	}

	messageBuffer := controller.CreateMessageBuffer(controller.config.SMTPSender, email, subject, message, plainMessage)

	_, err = messageBuffer.WriteTo(wc)
	if err != nil {
		return err
	}

	err = wc.Close()
	if err != nil {
		return err

	}

	err = smtpClient.Quit()
	if err != nil {
		return err
	}

	return nil
}

// CreateMessageBuffer creates a byte-buffer containing the properly formatted mail message content
func (controller *Controller) CreateMessageBuffer(from string, to string, subject string, message string, plainMessage string) bytes.Buffer {
	var buffer bytes.Buffer

	boundaryString := misc.GenerateRandomString(32)

	buffer.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buffer.WriteString("MIME-Version: 1.0\r\n")
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"", boundaryString))
	buffer.WriteString("\r\n")

	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundaryString))
	buffer.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(fmt.Sprintf("%s\r\n", plainMessage))

	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundaryString))
	buffer.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(fmt.Sprintf("%s\r\n", message))

	buffer.WriteString(fmt.Sprintf("--%s--\r\n", boundaryString))

	return buffer
}

// TemplateFunctions prepares a map of functions to be used within templates
func (controller *Controller) TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"FormatType":              func(t int64) string { return controller.FormatType(t) },
		"FormatLocation":          func(m int64) string { return controller.FormatLocation(m) },
		"FormatRemainingFuelTime": func(u int64, q int64) string { return controller.FormatRemainingFuelTime(u, q) },
	}
}

func (controller *Controller) FormatType(typeID int64) string {
	typeName, err := controller.database.QueryTypeName(typeID)
	if err != nil {
		return strconv.FormatInt(typeID, 10)
	}

	return typeName
}

func (controller *Controller) FormatLocation(moonID int64) string {
	location, err := controller.database.QueryLocationName(moonID)
	if err != nil {
		return strconv.FormatInt(moonID, 10)
	}

	return location
}

func (controller *Controller) FormatRemainingFuelTime(usage int64, quantity int64) string {
	return humanize.Time(time.Now().Add(time.Hour * time.Duration(quantity/usage)))
}
