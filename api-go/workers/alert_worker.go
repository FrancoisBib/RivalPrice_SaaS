package workers

import (
	"log"
	"time"

	"github.com/rivalprice/api-go/services"
	"gorm.io/gorm"
)

// AlertWorker polls detected_changes every 30s and processes new ones
type AlertWorker struct {
	alertSvc *services.AlertService
	interval time.Duration
	stopCh   chan struct{}
}

func NewAlertWorker(db *gorm.DB) *AlertWorker {
	return &AlertWorker{
		alertSvc: services.NewAlertService(db),
		interval: 30 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

// Start runs the alert worker loop in the background
func (w *AlertWorker) Start() {
	log.Println("🔔 AlertWorker started (interval: 30s)")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run once immediately on start
	w.run()

	for {
		select {
		case <-ticker.C:
			w.run()
		case <-w.stopCh:
			log.Println("🔔 AlertWorker stopped")
			return
		}
	}
}

// Stop gracefully stops the worker
func (w *AlertWorker) Stop() {
	close(w.stopCh)
}

// run processes all unprocessed detected_changes
func (w *AlertWorker) run() {
	changes, err := w.alertSvc.GetUnprocessedChanges()
	if err != nil {
		log.Printf("❌ AlertWorker: failed to fetch changes: %v", err)
		return
	}

	if len(changes) == 0 {
		return
	}

	log.Printf("🔔 AlertWorker: processing %d new change(s)", len(changes))

	for i := range changes {
		change := &changes[i]
		if err := w.alertSvc.ProcessChange(change); err != nil {
			log.Printf("❌ AlertWorker: error processing change %d: %v", change.ID, err)
		}
	}
}
