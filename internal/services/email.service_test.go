package services

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sendgrid/sendgrid-go"
	"gopkg.in/gomail.v2"
)

func emailClearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"SENDGRID_API_KEY", "SMTP_FROM", "SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASS"} {
		t.Setenv(k, "")
	}
}

func TestNewEmailService_SendGrid(t *testing.T) {
	emailClearEnv(t)
	t.Setenv("SENDGRID_API_KEY", "SG.x")
	t.Setenv("SMTP_FROM", "me@x.com")
	s := NewEmailService()
	if !s.useSendgrid || s.sendgrid == nil || s.dialer != nil || s.from != "me@x.com" {
		t.Fatalf("unexpected %+v", s)
	}
}

func TestNewEmailService_SMTP(t *testing.T) {
	emailClearEnv(t)
	t.Setenv("SMTP_HOST", "h")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_USER", "u")
	t.Setenv("SMTP_PASS", "p")
	s := NewEmailService()
	if s.useSendgrid || s.dialer == nil || s.from != "noreply@webbuilder.com" {
		t.Fatalf("unexpected %+v", s)
	}
	if s.dialer.Host != "h" || s.dialer.Port != 2525 || s.dialer.Username != "u" {
		t.Fatalf("dialer = %+v", s.dialer)
	}
	t.Setenv("SMTP_PORT", "abc")
	if p := NewEmailService().dialer.Port; p != 587 {
		t.Fatalf("port = %d, want default 587", p)
	}
}

func TestNewEmailService_Unconfigured(t *testing.T) {
	emailClearEnv(t)
	t.Setenv("SMTP_HOST", "h") // incomplete
	s := NewEmailService()
	if s.useSendgrid || s.dialer != nil || s.sendgrid != nil {
		t.Fatalf("unexpected %+v", s)
	}
}

func TestSendInvitationEmail_Mock(t *testing.T) {
	wantNoErr(t, (&EmailService{from: "a@b.c"}).SendInvitationEmail("to@x.com", "P", "http://l"))
}

func emailSendGridService(url string) *EmailService {
	return &EmailService{
		from:        "from@x.com",
		useSendgrid: true,
		sendgrid:    &sendgrid.Client{Request: sendgrid.GetRequest("key", "/v3/mail/send", url)},
	}
}

func TestSendInvitationEmail_SendGrid(t *testing.T) {
	var auth, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.WriteHeader(202)
	}))
	defer srv.Close()
	wantNoErr(t, emailSendGridService(srv.URL).SendInvitationEmail("to@x.com", "MyProj", "http://invite"))
	if auth != "Bearer key" {
		t.Fatalf("auth = %q", auth)
	}
	for _, s := range []string{"to@x.com", "from@x.com", "MyProj", "http://invite", "Invitation to collaborate on MyProj"} {
		if !strings.Contains(body, s) {
			t.Fatalf("body missing %q: %s", s, body)
		}
	}
}

func TestSendInvitationEmail_SendGridStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte("bad key"))
	}))
	defer srv.Close()
	err := emailSendGridService(srv.URL).SendInvitationEmail("t@x.com", "P", "l")
	wantErr(t, err, "SendGrid error: 401 - bad key")
}

func TestSendInvitationEmail_SendGridTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()
	if err := emailSendGridService(url).SendInvitationEmail("t@x.com", "P", "l"); err == nil {
		t.Fatal("expected transport error")
	}
}

func TestSendInvitationEmail_SMTPError(t *testing.T) {
	// Port 1 on loopback refuses connections; no mail leaves the machine.
	s := &EmailService{from: "f@x.com", dialer: gomail.NewDialer("127.0.0.1", 1, "u", "p")}
	if err := s.SendInvitationEmail("t@x.com", "P", "l"); err == nil {
		t.Fatal("expected SMTP dial error")
	}
}
