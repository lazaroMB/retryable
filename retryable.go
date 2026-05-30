// Package retryable enables retrying of functions with customizable 
// delay between retries, timeout function to treat as an error, and support 
// for cancelling execution. The package returns stat data, including the 
// error of the function (if any) or nil, the number of retries attempted, 
// and the number of timeouts that occurred.
//
// By default we are going to retry once, but you can change that.
//
// Basic example:
//  func DoSomething() error {
//   ..... any action
//  }
//
//  stats := retryable.Retry(DoSomething).Exec()
//
//  fmt.Println(stats.Err)
//  fmt.Println(stats.Timeout)
//  fmt.Println(stats.Retries)
package retryable

import (
	"context"
	"errors"
	"math"
	"time"
)

// Errors String constants
const (
	CANCEL_ERROR  = "Function cancelled"
	TIMEOUT_ERROR = "Function timeout"
)

var (
	ErrCancelled = errors.New(CANCEL_ERROR)
	ErrTimeout   = errors.New(TIMEOUT_ERROR)
)

// Retryable defines the interface for configuring and executing a retriable operation.
type Retryable interface {
	SetTimeout(timeout time.Duration) Retryable
	SetSleep(sleep time.Duration) Retryable
	SetRetries(retries int) Retryable
	SetBackoff(backoff Backoff) Retryable
	Cancel()
	Exec() Stats
}

// Deprecated: RetrayableI is a typo-retaining alias for Retryable. Use Retryable instead.
type RetrayableI = Retryable

// Backoff defines the interface for custom backoff strategies.
type Backoff interface {
	Next(attempt int) time.Duration
}

// ConstantBackoff implements a fixed delay between retries.
type ConstantBackoff struct {
	Duration time.Duration
}

func (c ConstantBackoff) Next(attempt int) time.Duration {
	return c.Duration
}

// ExponentialBackoff implements a backoff strategy where delay increases exponentially.
type ExponentialBackoff struct {
	Min    time.Duration
	Max    time.Duration
	Factor float64
}

func (e ExponentialBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return e.Min
	}
	d := time.Duration(float64(e.Min) * math.Pow(e.Factor, float64(attempt)))
	if d > e.Max || d < 0 { // check for overflow
		return e.Max
	}
	return d
}

// Stats represents the execution results.
type Stats struct {
	Err     error
	Retries int
	Timeout int
}

// Retrier implements the Retryable interface.
type Retrier struct {
	fn            func() error
	retries       int
	backoff       Backoff
	timeout       time.Duration
	cancelContext context.Context
	cancelFn      context.CancelFunc
}

// Deprecated: Retrayable is a typo-retaining alias for Retrier. Use Retrier instead.
type Retrayable = Retrier

// SetTimeout sets a time duration for the maximum amount of time each single 
// attempt can run before it is considered to have timed out.
func (r *Retrier) SetTimeout(timeout time.Duration) Retryable {
	r.timeout = timeout
	return r
}

// SetRetries sets the maximum number of attempts.
// If retries <= 0, the function will not be executed.
func (r *Retrier) SetRetries(retries int) Retryable {
	r.retries = retries
	return r
}

// SetSleep sets a fixed time duration delay between retries.
func (r *Retrier) SetSleep(sleep time.Duration) Retryable {
	r.backoff = ConstantBackoff{Duration: sleep}
	return r
}

// SetBackoff configures a custom backoff strategy.
func (r *Retrier) SetBackoff(backoff Backoff) Retryable {
	r.backoff = backoff
	return r
}

// Cancel cancels the ongoing and subsequent executions.
func (r *Retrier) Cancel() {
	r.cancelFn()
}

// Exec executes the function with the configured retry settings.
func (r *Retrier) Exec() Stats {
	var err error
	stats := Stats{Retries: -1}

	if r.retries <= 0 {
		return stats
	}

	for i := 0; i < r.retries; i++ {
		ch := make(chan error, 1)
		stats.Retries += 1
		go func() {
			ch <- r.fn()
		}()

		var timeoutChan <-chan time.Time
		var timer *time.Timer
		if r.timeout > 0 {
			timer = time.NewTimer(r.timeout)
			timeoutChan = timer.C
		}

		select {
		case err = <-ch:
			if timer != nil {
				timer.Stop()
			}
			stats.Err = err
			if err == nil {
				return stats
			}

			// If it's the last attempt, don't sleep
			if i < r.retries-1 && r.backoff != nil {
				sleepDur := r.backoff.Next(i)
				if sleepDur > 0 {
					select {
					case <-time.After(sleepDur):
					case <-r.cancelContext.Done():
						stats.Err = ErrCancelled
						return stats
					}
				}
			}
		case <-timeoutChan:
			stats.Err = ErrTimeout
			stats.Timeout++
		case <-r.cancelContext.Done():
			if timer != nil {
				timer.Stop()
			}
			stats.Err = ErrCancelled
			return stats
		}
	}
	return stats
}

// Retry creates and returns a Retryable instance.
func Retry(fn func() error) Retryable {
	ctx, cancel := context.WithCancel(context.Background())
	return &Retrier{
		fn:            fn,
		retries:       1,
		backoff:       ConstantBackoff{Duration: 0},
		cancelContext: ctx,
		cancelFn:      cancel,
	}
}
