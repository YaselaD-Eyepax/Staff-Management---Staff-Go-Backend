package workers

import (
	"context"
	"fmt"
	"log"
	"time"

	"events-service/internal/events/models"
	"events-service/internal/events/service"

	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var sesClient *ses.Client

func InitSES() (*ses.Client, error) {
	if sesClient != nil {
		return sesClient, nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("SES load config error: %w", err)
	}

	sesClient = ses.NewFromConfig(cfg)
	return sesClient, nil
}

type BroadcastWorker struct {
	Service      *service.EventService
	PollInterval time.Duration
	BatchSize    int
	StopCh       chan struct{}
}

func NewBroadcastWorker(svc *service.EventService) *BroadcastWorker {
	return &BroadcastWorker{
		Service:      svc,
		PollInterval: 5 * time.Second,
		BatchSize:    10,
		StopCh:       make(chan struct{}),
	}
}

func (w *BroadcastWorker) Start() {
	go func() {
		log.Println("BroadcastWorker: started")
		for {
			select {
			case <-w.StopCh:
				log.Println("BroadcastWorker: stopping")
				return
			default:
				w.runOnce()
				time.Sleep(w.PollInterval)
			}
		}
	}()
}

func (w *BroadcastWorker) Stop() {
	close(w.StopCh)
}

func (w *BroadcastWorker) runOnce() {
	jobs, err := w.Service.FetchPendingBroadcasts(w.BatchSize)
	if err != nil {
		log.Printf("BroadcastWorker: fetch error: %v\n", err)
		return
	}
	if len(jobs) == 0 {
		return
	}

	for _, job := range jobs {
		w.processJob(job)
	}
}

func (w *BroadcastWorker) processJob(job models.BroadcastQueue) {
	log.Printf("BroadcastWorker: processing job id=%d event_id=%s channel=%s attempts=%d\n",
		job.ID, job.EventID.String(), job.Channel, job.Attempts)

	attempts := job.Attempts + 1
	var lastErr *string

	var sendErr error

	switch job.Channel {

	case "fcm":
		payloadMap := map[string]any(job.Payload)
		sendErr = sendFCM(job.EventID, payloadMap)

	case "email":
		sendErr = w.sendEmailAllStaff(job.EventID)

	case "teams":
		sendErr = sendTeams(job.EventID, nil)

	default:
		msg := "unknown channel"
		lastErr = &msg
		sendErr = nil
	}

	if sendErr != nil {
		msg := sendErr.Error()
		lastErr = &msg

		if attempts >= 3 {
			_ = w.Service.UpdateBroadcastJobStatus(job.ID, "failed", attempts, lastErr)
			_ = w.logPublishAudit(job.EventID, job.Channel, "failed", map[string]any{"error": msg})
		} else {
			_ = w.Service.UpdateBroadcastJobStatus(job.ID, "pending", attempts, lastErr)
		}

		return
	}

	_ = w.Service.UpdateBroadcastJobStatus(job.ID, "sent", attempts, nil)
	_ = w.logPublishAudit(job.EventID, job.Channel, "sent", map[string]any{"note": "delivered"})
}

// ============================================================
//               SES EMAIL SENDER (Broadcast to All Staff)
// ============================================================

func (w *BroadcastWorker) sendEmailAllStaff(eventID uuid.UUID) error {
	client, err := InitSES()
	if err != nil {
		return err
	}

	// Load event content
	event, err := w.Service.GetEvent(eventID)
	if err != nil {
		return fmt.Errorf("cannot load event for SES: %w", err)
	}

	// Load all staff emails (implement in EventService)
	recipients, err := w.Service.GetAllStaffEmails()
	if err != nil {
		return fmt.Errorf("cannot load staff emails: %w", err)
	}

	if len(recipients) == 0 {
		log.Println("SES: No recipients found")
		return nil
	}

	// Build email
	subject := fmt.Sprintf("[Staff Announcement] %s", event.Title)

	bodyHTML := fmt.Sprintf(`
		<h2>%s</h2>
		<p><strong>%s</strong></p>
		<p>%s</p>

		<p>Scheduled at: %v</p>

		<br/><br/>
		<p>Regards,<br/>Eyepax Staff Management System</p>
	`,
		event.Title,
		event.Summary,
		event.Body.Body,
		event.ScheduledAt,
	)

	for _, email := range recipients {
		input := &ses.SendEmailInput{
    Destination: &types.Destination{
        ToAddresses: []string{"yasela2014@gmail.com"},
    },
    Message: &types.Message{
        Body: &types.Body{
            Html: &types.Content{
                Charset: aws.String("UTF-8"),
                Data:    aws.String(bodyHTML),
            },
        },
        Subject: &types.Content{
            Charset: aws.String("UTF-8"),
            Data:    aws.String(subject),
        },
    },
    Source: aws.String("yasela.d@eyepax.com"),
}


		_, err := client.SendEmail(context.Background(), input)
		if err != nil {
			log.Printf("SES: error sending to %s: %v\n", email, err)
		} else {
			log.Printf("SES: email sent to %s\n", email)
		}
	}

	return nil
}

// ============================================================
//                       Audit Logger
// ============================================================

func (w *BroadcastWorker) logPublishAudit(eventID uuid.UUID, channel, status string, details map[string]any) error {
	return w.Service.CreatePublishAudit(eventID, channel, status, details)
}

// ============================================================
//               FCM + TEAMS STUBS (existing)
// ============================================================

func sendFCM(eventID uuid.UUID, payload map[string]any) error {
	if fcmClient == nil {
		_, err := InitFCM("fcm.json")
		if err != nil {
			return err
		}
	}

	msg := &messaging.Message{
		Topic: "events",
		Notification: &messaging.Notification{
			Title: payload["title"].(string),
			Body:  payload["summary"].(string),
		},
	}

	_, err := fcmClient.Send(context.Background(), msg)
	return err
}

func sendTeams(eventID any, payload datatypes.JSONMap) error {
	log.Printf("sendTeams: event=%v payload=%v (stub)\n", eventID, payload)
	return nil
}
