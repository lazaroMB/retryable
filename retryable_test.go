package retryable

import (
	"errors"
	"testing"
	"time"
)

func TestBasicRetrySuccess(t *testing.T) {
	calls := 0
	fn := func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	rt := Retry(fn).SetRetries(3).SetSleep(1 * time.Millisecond)
	stats := rt.Exec()

	if stats.Err != nil {
		t.Errorf("expected no error, got %v", stats.Err)
	}
	if stats.Retries != 2 {
		t.Errorf("expected 2 retries, got %d", stats.Retries)
	}
}

func TestRetryCancel(t *testing.T) {
	fn := func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	rt := Retry(fn).SetRetries(5).SetSleep(1 * time.Millisecond)
	
	// Cancel after a short delay
	time.AfterFunc(10*time.Millisecond, func() {
		rt.Cancel()
	})

	stats := rt.Exec()

	if stats.Err == nil || stats.Err.Error() != CANCEL_ERROR {
		t.Errorf("expected cancel error, got %v", stats.Err)
	}

	// Wait to see if any background goroutine panics when writing to closed channel
	time.Sleep(100 * time.Millisecond)
}

func TestRetryTimeout(t *testing.T) {
	fn := func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	rt := Retry(fn).SetRetries(2).SetTimeout(10 * time.Millisecond)
	stats := rt.Exec()

	if stats.Err == nil || stats.Err.Error() != TIMEOUT_ERROR {
		t.Errorf("expected timeout error, got %v", stats.Err)
	}
	if stats.Timeout != 2 {
		t.Errorf("expected 2 timeouts, got %d", stats.Timeout)
	}
}

func TestExponentialBackoff(t *testing.T) {
	eb := ExponentialBackoff{
		Min:    10 * time.Millisecond,
		Max:    100 * time.Millisecond,
		Factor: 2.0,
	}

	// Test Next calculations
	if eb.Next(0) != 10*time.Millisecond {
		t.Errorf("expected 10ms for attempt 0, got %v", eb.Next(0))
	}
	if eb.Next(1) != 20*time.Millisecond {
		t.Errorf("expected 20ms for attempt 1, got %v", eb.Next(1))
	}
	if eb.Next(2) != 40*time.Millisecond {
		t.Errorf("expected 40ms for attempt 2, got %v", eb.Next(2))
	}
	if eb.Next(5) != 100*time.Millisecond { // Max check
		t.Errorf("expected 100ms max limit, got %v", eb.Next(5))
	}
}

func TestDeprecatedAliases(t *testing.T) {
	var _ RetrayableI = &Retrayable{}
}
