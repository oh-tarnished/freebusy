package booking

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/shared"
)

// defaultSweepInterval is how often the hold sweeper runs. Holds released at most
// this long after expiry — acceptable, since the capacity check treats an expired
// hold as active only until it is swept.
const defaultSweepInterval = time.Minute

// StartHoldSweeper launches a background loop that periodically expires lapsed
// PENDING_HOLD bookings until ctx is cancelled. It runs one pass immediately, then
// every interval (defaultSweepInterval when interval <= 0). It is non-blocking;
// the goroutine exits when ctx is done.
func (s *Server) StartHoldSweeper(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = defaultSweepInterval
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		s.sweepOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.sweepOnce(ctx)
			}
		}
	}()
}

// sweepOnce runs a single expiry pass, logging the outcome. Errors are logged and
// swallowed so a transient failure does not stop the loop.
func (s *Server) sweepOnce(ctx context.Context) {
	n, err := s.repo.ExpireHolds(ctx)
	switch {
	case err != nil:
		_ = shared.Pulse.Logger.Error("hold sweeper failed", map[string]any{"error": err.Error()})
	case n > 0:
		shared.Pulse.Logger.Debug("hold sweeper expired holds", map[string]any{"count": n})
	}
}
