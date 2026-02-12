package mailer

import (
	"log/slog"

	"github.com/resend/resend-go/v2"
)

type consoleMailer struct{}

func newConsoleMailer() consoleMailer {
	return consoleMailer{}
}

func (consoleMailer) Send(opts *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
	slog.Info("mail", "opts", opts)
	return nil, nil
}
