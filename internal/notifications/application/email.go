package application

import (
	"context"
	"log/slog"
	"newsletter/config"
	"newsletter/internal/notifications/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// EmailService is responsible for sending emails using AWS SES.
type EmailService struct {
	client *ses.Client
}

func NewEmailService(client *ses.Client) *EmailService {
	return &EmailService{client: client}
}

// Send sends an email to a recipient.
//
// Parameters:
//   - email: A pointer to domain.Email containing recipient info, subject, and body.
//
// Behavior:
//   - Constructs both HTML and plain text versions of the email.
//   - Sends the email via AWS SES.
//
// Notes:
//   - The "from" address must be verified in AWS SES (sandbox or production).
//   - In the SES sandbox, recipient addresses must also be verified.
//
// Returns:
//   - An error if sending the email fails; otherwise nil.
func (es *EmailService) Send(email *domain.Email) error {
	// Construct the SES SendEmailInput
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{email.To},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(email.HTML),
				},
				Text: &types.Content{
					Data: aws.String(email.Text),
				},
			},
			Subject: &types.Content{
				Data: aws.String(email.Subject),
			},
		},
		Source: aws.String(config.GetEnv("AWS_FROM", "")),
	}

	// Send the email
	response, err := es.client.SendEmail(context.TODO(), input)
	if err != nil {
		slog.Warn("Message was not delivered to recipient", "error", err)
		return err
	}

	slog.Info("Message was delivered successfully", "message", response.MessageId)

	return nil
}
