package ratelimit

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	rl *rate.Limiter
}

func New(rps int) *Limiter {
	if rps <= 0 {
		rps = 1
	}
	lim := rate.NewLimiter(rate.Limit(rps), rps)
	return &Limiter{rl: lim}
}

func (l *Limiter) Wait(ctx context.Context) error {
	return l.rl.Wait(ctx)
}

func (l *Limiter) Stop() {
	// nothing to stop; kept for symmetry / future extensions
	_ = time.Now()
}
