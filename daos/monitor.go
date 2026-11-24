package daos

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
)

const (
	slowQueryThreshold = 500 * time.Millisecond
	maxLoggedCommand   = 4096
)

// newSlowQueryMonitor logs every mongo command that fails or takes longer than slowQueryThreshold.
func newSlowQueryMonitor() *event.CommandMonitor {
	// Command bytes are copied because the driver may reuse the buffer after Started returns.
	var commands sync.Map

	take := func(requestID int64) string {
		v, ok := commands.LoadAndDelete(requestID)
		if !ok {
			return ""
		}
		return bson.Raw(v.([]byte)).String()
	}

	return &event.CommandMonitor{
		Started: func(_ context.Context, e *event.CommandStartedEvent) {
			if len(e.Command) <= maxLoggedCommand {
				commands.Store(e.RequestID, append([]byte(nil), e.Command...))
			}
		},
		Succeeded: func(_ context.Context, e *event.CommandSucceededEvent) {
			cmd := take(e.RequestID)
			if e.Duration >= slowQueryThreshold {
				logger.Warningf("slow mongo query: %s on %s took %s: %s", e.CommandName, e.DatabaseName, e.Duration, cmd)
			}
		},
		Failed: func(_ context.Context, e *event.CommandFailedEvent) {
			cmd := take(e.RequestID)
			logger.Errorf("mongo query failed: %s on %s after %s: %v: %s", e.CommandName, e.DatabaseName, e.Duration, e.Failure, cmd)
		},
	}
}
