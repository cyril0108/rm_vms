package events

import (
	"context"
	"fmt"

	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/email"
)

// EmailNotifier implements the Notifier interface.
// It receives events from the EventHub and sends email notifications
// to the groups subscribed to each event type.
type EmailNotifier struct {
	repo repository.EmailRepository
}

func NewEmailNotifier(repo repository.EmailRepository) *EmailNotifier {
	return &EmailNotifier{repo: repo}
}

func (n *EmailNotifier) ID() string {
	return "EmailNotifier"
}

// Handle is called by the EventHub for every published event.
// It checks if email is enabled, finds subscribed groups, and sends the notification.
func (n *EmailNotifier) Handle(ctx context.Context, e models.Event) {
	// 1. Get SMTP settings and check if enabled
	settings, err := n.repo.GetSMTPSettings(ctx)
	if err != nil {
		LOG.Error("[EmailNotifier] Failed to get SMTP settings", "error", err)
		return
	}

	if !settings.Enabled || settings.Host == "" {
		return
	}

	// 2. Find groups subscribed to this event type
	groups, err := n.repo.GetGroupsByEventType(ctx, string(e.Type))
	if err != nil {
		LOG.Error("[EmailNotifier] Failed to get groups for event", "type", e.Type, "error", err)
		return
	}

	if len(groups) == 0 {
		return
	}

	// 3. Collect unique recipients across all matched groups
	recipientSet := make(map[string]struct{})
	for _, g := range groups {
		for _, r := range g.Recipients {
			recipientSet[r] = struct{}{}
		}
	}

	recipients := make([]string, 0, len(recipientSet))
	for r := range recipientSet {
		recipients = append(recipients, r)
	}

	if len(recipients) == 0 {
		return
	}

	// 4. Build sender and send
	sender := &email.Sender{
		Host:        settings.Host,
		Port:        settings.Port,
		Username:    settings.Username,
		Password:    settings.Password,
		SenderEmail: settings.SenderEmail,
		SenderName:  settings.SenderName,
		UseTLS:      settings.UseTLS,
	}

	subject := fmt.Sprintf("[NVR Alert] %s", e.Type)
	timestamp := e.CreatedAt.Format("2006-01-02 15:04:05")

	bodyText := fmt.Sprintf("Event: %s\nMessage: %s\nTime: %s", e.Type, e.Message, timestamp)

	bodyHTML := fmt.Sprintf(
		`<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
	<div style="background: #d32f2f; color: white; padding: 16px 24px; border-radius: 8px 8px 0 0;">
		<h2 style="margin: 0;">⚠ NVR Alert</h2>
	</div>
	<div style="background: #f5f5f5; padding: 24px; border-radius: 0 0 8px 8px;">
		<table style="border-collapse: collapse;">
			<tr><td style="padding: 8px 16px 8px 0; font-weight: bold;">Event Type</td><td style="padding: 8px 0;">%s</td></tr>
			<tr><td style="padding: 8px 16px 8px 0; font-weight: bold;">Message</td><td style="padding: 8px 0;">%s</td></tr>
			<tr><td style="padding: 8px 16px 8px 0; font-weight: bold;">Time</td><td style="padding: 8px 0;">%s</td></tr>
		</table>
	</div>
</div>`,
		e.Type, e.Message, timestamp,
	)

	err = sender.Send(&email.Message{
		To:      recipients,
		Subject: subject,
		Body:    bodyText,
		HTML:    bodyHTML,
	})

	if err != nil {
		LOG.Error("[EmailNotifier] Failed to send email", "type", e.Type, "error", err)
	} else {
		LOG.Info("[EmailNotifier] Email sent", "type", e.Type, "recipients", len(recipients))
	}
}
