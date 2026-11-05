package repository

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithAggregationDBRetry_RecoversTransientDeadlockAndLockTimeout(t *testing.T) {
	ctx := context.Background()

	t.Run("recovers MariaDB Error 1213 deadlock on retry", func(t *testing.T) {
		attempts := 0
		err := withAggregationDBRetry(ctx, func() error {
			attempts++
			if attempts < 3 {
				return errors.New("Error 1213 (40001): Deadlock found when trying to get lock; try restarting transaction")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("expected nil error after transient deadlock recovery, got %v", err)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("recovers MariaDB Error 1205 lock wait timeout and SQLite busy", func(t *testing.T) {
		attempts := 0
		err := withAggregationDBRetry(ctx, func() error {
			attempts++
			if attempts == 1 {
				return errors.New("Error 1205 (HY000): Lock wait timeout exceeded; try restarting transaction")
			}
			if attempts == 2 {
				return errors.New("database is locked (5) (SQLITE_BUSY)")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("expected nil error after transient lock recovery, got %v", err)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("does not retry permanent syntax or constraint errors", func(t *testing.T) {
		attempts := 0
		permErr := errors.New("Error 1054 (42S22): Unknown column 'foo' in 'field list'")
		err := withAggregationDBRetry(ctx, func() error {
			attempts++
			return permErr
		})
		if !errors.Is(err, permErr) {
			t.Fatalf("expected permanent error returned immediately, got %v", err)
		}
		if attempts != 1 {
			t.Fatalf("expected 1 attempt for non-transient error, got %d", attempts)
		}
	})

	t.Run("aborts promptly when context is cancelled", func(t *testing.T) {
		cctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		err := withAggregationDBRetry(cctx, func() error {
			return errors.New("Error 1213: Deadlock found when trying to get lock")
		})
		if err == nil {
			t.Fatalf("expected context error when cancelled during backoff")
		}
	})
}

