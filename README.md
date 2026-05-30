# retryable

The `retryable` Go package enables retrying of functions with customizable delay between retries, timeouts, and cancellation support. The package returns statistics including the resulting error, the number of retries attempted, and the number of timeouts that occurred.

By default, it executes the function once and does not retry. You can customize this behavior with the fluent API.

## Features

* Fluent API for configuration
* Constant sleep time between retries
* **New**: Exponential Backoff strategy support
* Cooperative cancellation support
* Per-attempt execution timeout
* 100% backward compatibility for legacy types

---

## Installation

```bash
go get github.com/lazaroMB/retryable
```

---

## Basic Example

```go
package main

import (
	"fmt"
	"github.com/lazaroMB/retryable"
)

func DoSomething() error {
	// ... perform action ...
	return nil
}

func main() {
	stats := retryable.Retry(DoSomething).Exec()

	fmt.Println("Error:", stats.Err)
	fmt.Println("Timeouts:", stats.Timeout)
	fmt.Println("Retries:", stats.Retries)
}
```

---

## Example with Retries and Delay

```go
package main

import (
	"fmt"
	"time"
	"github.com/lazaroMB/retryable"
)

func DoSomething() error {
	// ... perform action ...
	return nil
}

func main() {
	stats := retryable.Retry(DoSomething).
		SetRetries(4).                    // Try up to 4 times (1 initial attempt + 3 retries)
		SetSleep(500 * time.Millisecond). // Wait 500ms between attempts
		Exec()

	fmt.Println("Error:", stats.Err)
	fmt.Println("Timeouts:", stats.Timeout)
	fmt.Println("Retries:", stats.Retries)
}
```

---

## Example with Timeout and Cancellation

```go
package main

import (
	"fmt"
	"time"
	"github.com/lazaroMB/retryable"
)

func PollApi() error {
	// ... perform network request ...
	return nil
}

func main() {
	rt := retryable.Retry(PollApi).
		SetTimeout(5 * time.Second).     // Each attempt times out after 5 seconds
		SetSleep(1 * time.Second).       // Sleep 1 second between attempts
		SetRetries(10)

	// Cancel execution asynchronously after 10 seconds
	time.AfterFunc(10 * time.Second, rt.Cancel)

	stats := rt.Exec()
	fmt.Println("Error:", stats.Err) // Will contain "Function cancelled" if cancelled
	fmt.Println("Timeouts:", stats.Timeout)
	fmt.Println("Retries:", stats.Retries)
}
```

---

## Advanced: Exponential Backoff

You can configure exponential backoff using the `SetBackoff` method:

```go
package main

import (
	"fmt"
	"time"
	"github.com/lazaroMB/retryable"
)

func PollApi() error {
	// ... perform network request ...
	return nil
}

func main() {
	backoff := retryable.ExponentialBackoff{
		Min:    100 * time.Millisecond, // Initial wait time
		Max:    5 * time.Second,        // Maximum wait time limit
		Factor: 2.0,                    // Multiplier factor
	}

	stats := retryable.Retry(PollApi).
		SetRetries(5).
		SetBackoff(backoff).
		Exec()

	fmt.Println("Error:", stats.Err)
	fmt.Println("Retries:", stats.Retries)
}
```

---

## Backward Compatibility Note

To maintain backward compatibility with previous versions of this library, the typo-containing types are kept as deprecated aliases and can still compile:

* `RetrayableI` is an alias for `Retryable` interface.
* `Retrayable` is an alias for `Retrier` struct.
