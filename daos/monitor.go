package daos

import (
	"context"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
)

const (
	slowQueryThreshold = 500 * time.Millisecond
	maxLoggedCommand   = 4096
)

// loadAvg returns the 1/5/15 minute load averages and CPU count; load is "n/a" where /proc/loadavg is missing.
func loadAvg() (string, int) {
	load := "n/a"
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		if f := strings.Fields(string(b)); len(f) >= 3 {
			load = strings.Join(f[:3], " ")
		}
	}
	return load, runtime.NumCPU()
}

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
				load, cpus := loadAvg()
				logger.Warningf("slow mongo query: %s on %s took %s (load avg %s, %d cpus): %s", e.CommandName, e.DatabaseName, e.Duration, load, cpus, cmd)
			}
		},
		Failed: func(_ context.Context, e *event.CommandFailedEvent) {
			cmd := take(e.RequestID)
			load, cpus := loadAvg()
			logger.Errorf("mongo query failed: %s on %s after %s (load avg %s, %d cpus): %v: %s", e.CommandName, e.DatabaseName, e.Duration, load, cpus, e.Failure, cmd)
		},
	}
}
