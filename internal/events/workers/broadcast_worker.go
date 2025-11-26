package workers

import (
	"log"
	"time"

	"events-service/internal/events/models"
	"events-service/internal/events/service"

	"github.com/google/uuid"
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
	log.Printf("BroadcastWorker: processing job id=%d event_id=%s channel=%s attempts=%d\n", job.ID, job.EventID.String(), job.Channel, job.Attempts)

	attempts := job.Attempts + 1
	var lastErr *string

	// Build event payload (you may want to fetch event details if needed)
	payload := job.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	// Placeholder: call real sender functions per channel
	var sendErr error
	switch job.Channel {
	case "fcm":
		// TODO: implement FCM send with Firebase admin SDK
		sendErr = sendFCM(job.EventID, payload)
	case "email":
		// TODO: implement email via AWS SES
		sendErr = sendEmail(job.EventID, payload)
	case "teams":
		// TODO: implement Teams via Microsoft Graph
		sendErr = sendTeams(job.EventID, payload)
	default:
		s := "unknown channel"
		lastErr = &s
		sendErr = nil
	}

	if sendErr != nil {
		msg := sendErr.Error()
		lastErr = &msg
		// mark failed (or retry)
		if attempts >= 3 {
			_ = w.Service.UpdateBroadcastJobStatus(job.ID, "failed", attempts, lastErr)
			// Write audit entry for failed channel
			_ = w.logPublishAudit(job.EventID, job.Channel, "failed", map[string]any{"error": msg})
		} else {
			// set back to pending for retry with incremented attempts
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
func sendFCM(eventID any, payload map[string]any) error {
	log.Printf("sendFCM: event=%v payload=%v (stub)\n", eventID, payload)
	return nil
}
func sendEmail(eventID any, payload map[string]any) error {
	log.Printf("sendEmail: event=%v payload=%v (stub)\n", eventID, payload)
	return nil
}
func sendTeams(eventID any, payload map[string]any) error {
	log.Printf("sendTeams: event=%v payload=%v (stub)\n", eventID, payload)
	return nil
}
