package mailer

import "testing"

func TestNewReturnsNoopWhenHostEmpty(t *testing.T) {
	m := New(SMTPConfig{})
	if _, ok := m.(NoopMailer); !ok {
		t.Fatalf("sans SMTP_HOST, New() devrait renvoyer NoopMailer, obtenu %T", m)
	}
}

func TestNewReturnsSMTPMailerWhenHostSet(t *testing.T) {
	m := New(SMTPConfig{Host: "mailhog", Port: "1025", From: "test@collector.shop"})
	smtpMailer, ok := m.(*SMTPMailer)
	if !ok {
		t.Fatalf("avec SMTP_HOST defini, New() devrait renvoyer *SMTPMailer, obtenu %T", m)
	}
	if smtpMailer.cfg.Host != "mailhog" {
		t.Errorf("host attendu 'mailhog', obtenu %q", smtpMailer.cfg.Host)
	}
}

func TestNoopMailerSendDoesNotPanic(t *testing.T) {
	NoopMailer{}.Send("dest@example.com", "sujet", "corps")
}

func TestSMTPMailerSendUnreachableHostDoesNotPanic(t *testing.T) {
	// Aucun serveur SMTP sur ce port : Send() doit logger l'echec et revenir
	// sans jamais paniquer ni bloquer indefiniment (pas de timeout configure
	// cote net/smtp mais le refus de connexion est immediat en local).
	m := NewSMTPMailer(SMTPConfig{Host: "127.0.0.1", Port: "1", From: "test@collector.shop"})
	m.Send("dest@example.com", "sujet", "corps")
}
