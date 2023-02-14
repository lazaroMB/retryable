// The Retrayable package enables retrying of functions with customizable 
// delay between retries, timeout function to treat as an error, and support 
// forcancelling execution. The package returns stat data, including the 
// error of the function (if any) or nil, the number of retries attempted, 
// and the number of timeouts that occurred.
//
// By default we are going to retry once, but you can change that
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
//
// Example with retry:
//  func DoSomething() error {
//   ..... any action
//  }
//
//  stats := retryable.Retry(DoSomething)
//   .SetRetries(4) // we execute function 1 time and retry a max of 4 times
//   .Exec()
//
//  fmt.Println(stats.Err)
//  fmt.Println(stats.Timeout)
//  fmt.Println(stats.Retries)
//
// Example with Sleep and Timeout:
// You can use Sleep and Timeout methods in any order
//  func PollApi() error {
//   ..... Poll an API and return and error if fails
//  }
//
//  stats := retryable.Retry(PollApi)
//    .Sleep(3 * time.Second) // Wait 3 seconds between each retry
//    .Timeout(15 * time.Second) // Each function execution will fail if takes more than 15 seconds then we make a retry or finish
//    .Exec()
// 
//  fmt.Println(stats.Err)
//  fmt.Println(stats.Timeout)
//  fmt.Println(stats.Retries)
//
// Full example:
// Example with Sleep and Timeout:
//  func PollApi() error {
//   ..... Poll an API and return and error if fails
//  }
//
//  rt := retryable.Retry(PollApi)
//    .Timeout(15 * time.Second)
//    .Sleep(3 * time.Second)
//    .SetRetries(10)
//
//  go time.AfterFunc(10 * time.Second, rt.Cancel) // Exec flow cancel fn after 10 seconds
//
//  stats := rt.Exec() 
//  fmt.Println(stats.Err) // Cancellation error
//  fmt.Println(stats.Timeout)
//  fmt.Println(stats.Retries)
package retryable

import (
	"context"
	"errors"
	"time"
)

// Errors String constants
const (
	CANCEL_ERROR  = "Function cancelled"
	TIMEOUT_ERROR = "Function timeout"
)

type RetrayableI interface {
	SetTimeout(timeout time.Duration) RetrayableI
	SetSleep(sleep time.Duration) RetrayableI
	SetRetries(retries int) RetrayableI
	Cancel()
	Exec() Stats
}

// The Err field is an error that represents the result of the function 
// execution. If the function was successful, Err will be nil. Otherwise, 
// Err will contain the error that caused the function to fail.
// The Retries field is an integer that represents the number of times the 
// function was retried before it either succeeded or failed permanently.
// The Timeout field is an integer that represents the number of times the 
// function was timed out before it either succeeded or failed permanently.
type Stats struct {
	Err     error
	Retries int
	Timeout int
}

type Retrayable struct {
	fn            func() error
	retries       int
	sleep         time.Duration
	timeout       time.Duration
	cancelContext context.Context
	cancelFn      context.CancelFunc
}

// The SetTimeout method sets a time duration for the maximum amount of 
// time the function can run before it is considered to have timed out. 
// It returns a RetrayableI instance, allowing method chaining.
func (r *Retrayable) SetTimeout(timeout time.Duration) RetrayableI {
	r.timeout = timeout
	return r
}

// The SetRetries method sets the maximum number of times the function can 
// be retried if it fails. It returns a RetrayableI instance, allowing method 
// chaining.
func (r *Retrayable) SetRetries(retries int) RetrayableI {
	r.retries = retries
	return r
}

// The SetSleep method sets a time duration for the delay between retries. 
// It returns a RetrayableI instance, allowing method chaining.
func (r *Retrayable) SetSleep(sleep time.Duration) RetrayableI {
	r.sleep = sleep
	return r
}

// The Cancel method cancels the execution of the function. It does not return anything.
func (r *Retrayable) Cancel() {
	r.cancelFn()
}

func (r *Retrayable) GetTimeout() <-chan time.Time {
	if r.timeout == 0 {
		return make(<-chan time.Time)
	}

	return time.After(r.timeout)
}

// he Exec method executes the function with the specified settings and returns a 
// Stats struct that contains the error result of the function (if any), the number 
// of retries attempted, and the number of timeouts that occurred.
func (r *Retrayable) Exec() Stats {
	var err error
	stats := Stats{Retries: -1}
	for i := 0; i < r.retries; i++ {
		ch := make(chan error, 1)
		stats.Retries += 1
		go func() {
			ch <- r.fn()
		}()
		select {
		case err = <-ch:
			stats.Err = err
			if err == nil {
				return stats
			}
			time.Sleep(r.sleep)
		case <-r.GetTimeout():
			stats.Err = errors.New(TIMEOUT_ERROR)
			stats.Timeout++
		case <-r.cancelContext.Done():
			close(ch)
			stats.Err = errors.New(CANCEL_ERROR)
			return stats
		}
	}
	return stats
}

// The function Retry is creating and returning an instance of the type RetrayableI.
// The function takes an argument fn, which is a function that returns an error. 
func Retry(fn func() error) RetrayableI {
	ctx, cancel := context.WithCancel(context.Background())
	return &Retrayable{fn: fn, retries: 1, cancelContext: ctx, cancelFn: cancel}
}
