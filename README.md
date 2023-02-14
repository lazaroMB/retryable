# retryable
The Retrayable package enables retrying of functions with customizable 
delay between retries, timeout function to treat as an error, and support 
forcancelling execution. The package returns stat data, including the 
error of the function (if any) or nil, the number of retries attempted, 
and the number of timeouts that occurred.

By default we are going to retry once, but you can change that

## Features
* Sleep time between retries
* Max time function of execution
* Set the retries number
* Cacel execution

## Basic example
```
func DoSomething() error {
  ..... any action
 }

 stats := retryable.Retry(DoSomething).Exec()

 fmt.Println(stats.Err)
 fmt.Println(stats.Timeout)
 fmt.Println(stats.Retries)
```

## Example with retry
```
 func DoSomething() error {
  ..... any action
 }

 stats := retryable.Retry(DoSomething)
  .SetRetries(4) // we execute function 1 time and retry a max of 4 times
  .Exec()

 fmt.Println(stats.Err)
 fmt.Println(stats.Timeout)
 fmt.Println(stats.Retries)
```

## Example with Sleep and Timeout
You can use Sleep and Timeout methods in any order
```
 func PollApi() error {
  ..... Poll an API and return and error if fails
 }

 stats := retryable.Retry(PollApi)
   .Sleep(3 * time.Second) // Wait 3 seconds between each retry
   .Timeout(15 * time.Second) // Each function execution will fail if takes more than 15 seconds then we make a retry or finish
   .Exec()

 fmt.Println(stats.Err)
 fmt.Println(stats.Timeout)
 fmt.Println(stats.Retries)
```

## Full example
Example with Sleep and Timeout:
```
 func PollApi() error {
  ..... Poll an API and return and error if fails
 }

 rt := retryable.Retry(PollApi)
   .Timeout(15 * time.Second)
   .Sleep(3 * time.Second)
   .SetRetries(10)

 go time.AfterFunc(10 * time.Second, rt.Cancel) // Exec flow cancel fn after 10 seconds

 stats := rt.Exec() 
 fmt.Println(stats.Err) // Cancellation error
 fmt.Println(stats.Timeout)
 fmt.Println(stats.Retries)
```
