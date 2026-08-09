package handlers

import (
	"context"
	"log"
	"time"
)

const clickTrackingTimeout = 2 * time.Second

// trackClickAsync keeps the read response independent of analytics writes.
// The detached, bounded context lets the update finish after the HTTP request
// completes without allowing an unavailable database to leave a goroutine stuck.
func (h *URLHandler) trackClickAsync(shortCode string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), clickTrackingTimeout)
		defer cancel()

		if err := h.service.IncrementClickCount(ctx, shortCode); err != nil {
			log.Printf("failed to track click for short code %q: %v", shortCode, err)
		}
	}()
}
