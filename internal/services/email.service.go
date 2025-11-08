package services

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	dialer     *gomail.Dialer
	from       string
	sendgrid   *sendgrid.Client
	useSendgrid bool
}

func NewEmailService() *EmailService {
	sendgridKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "noreply@webbuilder.com" // default
	}

	if sendgridKey != "" {
		client := sendgrid.NewSendClient(sendgridKey)
		fmt.Println("Email service initialized with SendGrid")
		return &EmailService{
			from:        from,
			sendgrid:    client,
			useSendgrid: true,
		}
	}

	// Fall back to SMTP
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	if host == "" || portStr == "" || user == "" || pass == "" {
		fmt.Println("Neither SendGrid nor SMTP configuration complete, emails will not be sent")
		return &EmailService{from: from}
	}

	port := 587 // default
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	dialer := gomail.NewDialer(host, port, user, pass)
	fmt.Println("Email service initialized with SMTP")

	return &EmailService{
		dialer: dialer,
		from:   from,
	}
}

func (e *EmailService) SendInvitationEmail(to, projectName, inviteLink string) error {
	fmt.Printf("Attempting to send invitation email to %s for project %s\n", to, projectName)
	if e.useSendgrid && e.sendgrid != nil {
		return e.sendInvitationEmailSendGrid(to, projectName, inviteLink)
	} else if e.dialer != nil {
		return e.sendInvitationEmailSMTP(to, projectName, inviteLink)
	} else {
		fmt.Printf("Mock email: Invitation to %s for project %s: %s\n", to, projectName, inviteLink)
		return nil
	}
}

func (e *EmailService) sendInvitationEmailSendGrid(to, projectName, inviteLink string) error {
	from := mail.NewEmail("WebBuilder", e.from)
	subject := fmt.Sprintf("Invitation to collaborate on %s", projectName)
	toEmail := mail.NewEmail("", to)

	htmlContent := fmt.Sprintf(`
<html>
<body>
<h2>You've been invited to collaborate!</h2>
<p>You have been invited to join the project <strong>%s</strong> as a collaborator.</p>
<p>Click the link below to accept the invitation:</p>
<a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Accept Invitation</a>
<p>This link will expire in 7 days.</p>
<p>If you did not expect this invitation, please ignore this email.</p>
</body>
</html>
`, projectName, inviteLink)

	message := mail.NewSingleEmail(from, subject, toEmail, "", htmlContent)

	fmt.Printf("Sending invitation email via SendGrid to %s\n", to)
	response, err := e.sendgrid.Send(message)
	if err != nil {
		fmt.Printf("Failed to send email via SendGrid: %v\n", err)
		return err
	}

	if response.StatusCode >= 400 {
		err := fmt.Errorf("SendGrid error: %d - %s", response.StatusCode, response.Body)
		fmt.Printf("Failed to send email via SendGrid: %v\n", err)
		return err
	}

	fmt.Printf("Successfully sent invitation email via SendGrid to %s\n", to)
	return nil
}

func (e *EmailService) sendInvitationEmailSMTP(to, projectName, inviteLink string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("Invitation to collaborate on %s", projectName))
	m.SetBody("text/html", fmt.Sprintf(`
<html>
<body>
<h2>You've been invited to collaborate!</h2>
<p>You have been invited to join the project <strong>%s</strong> as a collaborator.</p>
<p>Click the link below to accept the invitation:</p>
<a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Accept Invitation</a>
<p>This link will expire in 7 days.</p>
<p>If you did not expect this invitation, please ignore this email.</p>
</body>
</html>
`, projectName, inviteLink))

	fmt.Printf("Sending invitation email via SMTP to %s\n", to)
	err := e.dialer.DialAndSend(m)
	if err != nil {
		fmt.Printf("Failed to send email via SMTP: %v\n", err)
		return err
	}
	fmt.Printf("Successfully sent invitation email via SMTP to %s\n", to)
	return nil
}
