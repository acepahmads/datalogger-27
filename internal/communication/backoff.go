package communication

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

// BackoffPolicy configures bounded exponential backoff with random jitter
type BackoffPolicy struct {
	InitialInterval     time.Duration // Default: 500ms
	MaxInterval         time.Duration // Default: 30s
	Multiplier          float64       // Default: 2.0
	RandomizationFactor float64       // Default: 0.2 (+/- 20% jitter)
	MaxAttempts         int           // Maximum consecutive attempts before error state
	mu                  sync.Mutex
	rng                 *rand.Rand
}

// DefaultBackoffPolicy returns recommended production settings for industrial edge devices
func DefaultBackoffPolicy() *BackoffPolicy {
	return &BackoffPolicy{
		InitialInterval:     500 * time.Millisecond,
		MaxInterval:         30 * time.Second,
		Multiplier:          2.0,
		RandomizationFactor: 0.2,
		MaxAttempts:         10,
		rng:                 rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CalculateDelay calculates the backoff duration for a given attempt index (0-based)
func (b *BackoffPolicy) CalculateDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	initialMs := float64(b.InitialInterval.Milliseconds())
	if initialMs <= 0 {
		initialMs = 500
	}

	maxMs := float64(b.MaxInterval.Milliseconds())
	if maxMs <= 0 {
		maxMs = 30000
	}

	mult := b.Multiplier
	if mult < 1.0 {
		mult = 2.0
	}

	// Exponential delay: Initial * Multiplier^attempt
	delayMs := initialMs * math.Pow(mult, float64(attempt))
	if delayMs > maxMs {
		delayMs = maxMs
	}

	// Apply Jitter: +/- RandomizationFactor
	factor := b.RandomizationFactor
	if factor > 0 && factor < 1.0 {
		b.mu.Lock()
		if b.rng == nil {
			b.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		}
		// Random float in [-factor, +factor]
		delta := (b.rng.Float64()*2.0 - 1.0) * factor * delayMs
		b.mu.Unlock()

		delayMs += delta
		if delayMs < 0 {
			delayMs = initialMs
		}
	}

	return time.Duration(delayMs) * time.Millisecond
}

// Sleep waits for the calculated delay for the given attempt, respecting context cancellation
func (b *BackoffPolicy) Sleep(ctx context.Context, attempt int) error {
	delay := b.CalculateDelay(attempt)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
