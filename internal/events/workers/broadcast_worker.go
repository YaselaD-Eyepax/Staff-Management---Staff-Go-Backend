package workers

import (
	"context"
	"log"
	"time"

	"events-service/internal/events/models"
	"events-service/internal/events/service"

	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type BroadcastWorker struct {
	Service      *service.EventService
	PollInterval time.Duration
	BatchSize    int
	StopCh       chan struct{}
}

// NewBroadcastWorker constructs worker
func NewBroadcastWorker(svc *service.EventService) *BroadcastWorker {
	return &BroadcastWorker{
		Service:      svc,
		PollInterval: 5 * time.Second,
		BatchSize:    10,
		StopCh:       make(chan struct{}),
	}
}

// Start the worker (non-blocking)
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

// Stop worker
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
		// process each job (sync in this example)
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
        // Convert datatypes.JSONMap → map[string]any
        payloadMap := map[string]any(job.Payload)

        sendErr = sendFCM(job.EventID, payloadMap)

    case "email":
        sendErr = sendEmail(job.EventID, nil)

    case "teams":
        sendErr = sendTeams(job.EventID, nil)

    default:
        msg := "unknown channel"
        lastErr = &msg
        sendErr = nil
    }

    // Handle error
    if sendErr != nil {
        msg := sendErr.Error()
        lastErr = &msg

        // retry or fail
        if attempts >= 3 {
            _ = w.Service.UpdateBroadcastJobStatus(job.ID, "failed", attempts, lastErr)
            _ = w.logPublishAudit(job.EventID, job.Channel, "failed", map[string]any{"error": msg})
        } else {
            _ = w.Service.UpdateBroadcastJobStatus(job.ID, "pending", attempts, lastErr)
        }

        return
    }

    // success
    _ = w.Service.UpdateBroadcastJobStatus(job.ID, "sent", attempts, nil)
    _ = w.logPublishAudit(job.EventID, job.Channel, "sent", map[string]any{"note": "delivered"})
}


func (w *BroadcastWorker) logPublishAudit(eventID uuid.UUID, channel, status string, details map[string]any) error {
    return w.Service.CreatePublishAudit(eventID, channel, status, details)
}


// --- Placeholder sender stubs (replace with real SDK calls) ---
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



func sendEmail(eventID any, payload datatypes.JSONMap) error {
	log.Printf("sendEmail: event=%v payload=%v (stub)\n", eventID, payload)
	return nil
}
func sendTeams(eventID any, payload datatypes.JSONMap) error {
	log.Printf("sendTeams: event=%v payload=%v (stub)\n", eventID, payload)
	return nil
}
