// Package mailer envoie les emails transactionnels du service (notification
// de commande au vendeur). Suit la meme convention que events.Publisher côté
// catalog-service : une implementation Noop (log uniquement) quand SMTP_HOST
// n'est pas configure, pour ne jamais bloquer le service en dev/CI.
package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/rs/zerolog/log"
)

// Mailer envoie un email texte simple. Les erreurs sont volontairement non
// bloquantes pour l'appelant (log + poursuite) : un email perdu ne doit
// jamais faire echouer le traitement d'un evenement metier.
type Mailer interface {
	Send(to, subject, body string)
}

// NoopMailer journalise l'email au lieu de l'envoyer (SMTP_HOST absent).
type NoopMailer struct{}

func (NoopMailer) Send(to, subject, body string) {
	log.Info().Str("to", to).Str("subject", subject).Msg("email non envoye (SMTP non configure) — voir body en debug")
	log.Debug().Str("to", to).Str("body", body).Msg("contenu email")
}

// SMTPConfig regroupe la configuration serveur SMTP (dev : MailHog, sans TLS
// ni authentification ; prod : renseigner User/Password).
type SMTPConfig struct {
	Host     string
	Port     string
	From     string
	User     string
	Password string
}

type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(to, subject, body string) {
	addr := fmt.Sprintf("%s:%s", m.cfg.Host, m.cfg.Port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.cfg.From, to, subject, body)

	var auth smtp.Auth
	if m.cfg.User != "" {
		auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg)); err != nil {
		log.Error().Err(err).Str("to", to).Msg("envoi email echoue")
		return
	}
	log.Info().Str("to", to).Str("subject", subject).Msg("email envoye")
}

// New choisit l'implementation selon la configuration : SMTPMailer si
// SMTP_HOST est defini, sinon NoopMailer.
func New(cfg SMTPConfig) Mailer {
	if cfg.Host == "" {
		log.Warn().Msg("SMTP_HOST non defini : emails desactives (log uniquement)")
		return NoopMailer{}
	}
	return NewSMTPMailer(cfg)
}
