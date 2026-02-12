package mailer

import (
	"os"

	"github.com/resend/resend-go/v2"
)

var (
	apiKey string
	client *resend.Client
	mailer Mailer
)

func init() {
	apiKey = os.Getenv("RESEND_API_KEY")
	if apiKey == "console" {
		client = nil
		mailer = newConsoleMailer()
	} else {
		client = resend.NewClient(apiKey)
		mailer = client.Emails
	}
}
