# Action: A Go Actor Model Library

**Action** is a lightweight actor model library for Go that simplifies concurrency by encapsulating state and message handling within actors. By enforcing single-threaded execution of actions, it avoids the complexity of manual synchronization (like mutexes), trading fine-grained control for predictability and maintainability — especially in high-concurrency scenarios.

## Features

- **Simple Actor Model:** Encapsulate state and business logic inside a single-threaded actor.
- **Ease of Use:** Intuitive APIs for sending messages and retrieving results.
- **Context-Aware:** Built-in support for context cancellation.
- **Flexible:** Build complex concurrent workflows using actors as building blocks.
- **Benchmarked:** Compare actor-based synchronization with traditional mutex-based approaches.

## Installation

Install via `go get`:

```bash
go get github.com/neonima/action
```

## Quick Start

Below is an example demonstrating how to encapsulate an actor runner within a service struct. The service wraps its methods with actor calls for both sending messages and retrieving values.

```go
package main

import (
	"context"
	"fmt"
	"time"

	. "github.com/neonima/action"
)

// MyService encapsulates an actor runner to safely handle concurrent operations.
type MyService struct {
	actor *Runner // The actor runner instance.
}

// NewMyService creates and starts a new MyService instance.
func NewMyService(ctx context.Context) *MyService {
	// Create a new actor.
	actor := New()
	if err := actor.Start(ctx); err != nil {
		panic(fmt.Sprintf("failed to start actor: %v", err))
	}
	return &MyService{actor: actor}
}

// DoWork enqueues an action to be processed by the actor.
func (s *MyService) DoWork(message string) {
	Act(s.actor, func() {
		// This function is executed by the actor.
		fmt.Println("Processing:", message)
	})
}

// GetCurrentTime safely retrieves data using ActGet.
func (s *MyService) GetCurrentTime() time.Time {
	return ActGet(s.actor, func() time.Time {
		return time.Now()
	})
}

func main() {
	// Instantiate and use the service.
	svc := NewMyService(context.Background())
	svc.DoWork("Hello from the Actor Model!")

	now := svc.GetCurrentTime()
	fmt.Println("Current time:", now)
}
```

## Context-Aware Actions

If you want your action to be cancelable or timeout-aware, you can use the runner's context directly inside the action. For example:

### Basic Pattern: Respect runner shutdown

Use this pattern if your action might block or perform non-trivial work. For fast operations, cancellation checks are typically unnecessary and can be omitted for simplicity.

```go
ActErr(r, func() error {
	select {
	case <-r.Ctx().Done():
		return r.Ctx().Err()
	default:
		// continue execution
		return doSomething()
	}
})
```

You could choose to include this check by default in your actions, especially when you want to short-circuit execution without custom timeouts.

### Custom Timeout Pattern

You can combine a custom timeout with the runner's shutdown context to ensure your action aborts in either case:

```go
ctx, cancel := context.WithTimeout(r.Ctx(), 2*time.Second)
defer cancel()

ActErr(r, func() error {
	select {
	case <-r.Ctx().Done():
		return r.Ctx().Err()
	case <-ctx.Done():
		return ctx.Err()
	default:
		return doSomething()
	}
})
```

### Extended Pattern: Work with periodic checks or long tasks

You can also use the context in long-running loops:

```go
ActErr(r, func() error {
	ctx := r.Ctx()
	for i := 0; i < 100; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
	}
	return nil
})
```

### Alternative: Pass the context into your own functions

If your code is broken into helpers, just pass the runner's context through:

```go
func myOp(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

ActErr(r, func() error {
	return myOp(r.Ctx())
})
```

## Advanced Example: Optimized DataStore

In this example, we demonstrate a more complex data structure that uses a ring buffer (A ring buffer is a fixed-size circular queue that overwrites old data when new data comes in once it’s full). The actor-based solution wraps this data structure for safe concurrent access, while a mutex-based version is provided for comparison.

### DataStore and Services

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/neonima/action"
)

// DataStore represents a ring buffer for integers.
type DataStore struct {
	buf  []int
	head int
	size int
}

// NewDataStore creates a new DataStore with the specified capacity.
func NewDataStore(capacity int) *DataStore {
	return &DataStore{
		buf:  make([]int, capacity),
		head: 0,
		size: capacity,
	}
}

// Add inserts a value into the ring buffer in O(1) time.
func (d *DataStore) Add(val int) {
	d.buf[d.head] = val
	d.head = (d.head + 1) % d.size
}

// GetLast returns the most recently added value in O(1) time.
func (d *DataStore) GetLast() int {
	idx := (d.head - 1 + d.size) % d.size
	return d.buf[idx]
}

// DataStoreService demonstrates an actor-based approach to safely access DataStore.
type DataStoreService struct {
	actor *Runner
	ds    *DataStore
}

// NewDataStoreService creates and starts a new DataStoreService.
func NewDataStoreService(ctx context.Context, capacity int) *DataStoreService {
	actor := New(WithChanSize(10))
	if err := actor.Start(ctx); err != nil {
		panic(fmt.Sprintf("failed to start actor: %v", err))
	}
	return &DataStoreService{
		actor: actor,
		ds:    NewDataStore(capacity),
	}
}

// AddValue enqueues an Add operation to be processed by the actor.
func (s *DataStoreService) AddValue(val int) {
	Act(s.actor, func() {
		s.ds.Add(val)
	})
}

// GetLastValue retrieves the most recent value via the actor.
func (s *DataStoreService) GetLastValue() int {
	return ActGet(s.actor, func() int {
		return s.ds.GetLast()
	})
}

// MutexDataStoreService uses a RWMutex to protect concurrent access to DataStore.
type MutexDataStoreService struct {
	ds  *DataStore
	mtx sync.RWMutex
}

// NewMutexDataStoreService creates a new MutexDataStoreService.
func NewMutexDataStoreService(capacity int) *MutexDataStoreService {
	return &MutexDataStoreService{
		ds: NewDataStore(capacity),
	}
}

// AddValue adds a value using write-lock protection.
func (s *MutexDataStoreService) AddValue(val int) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.ds.Add(val)
}

// GetLastValue retrieves the last value using read-lock protection.
func (s *MutexDataStoreService) GetLastValue() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.ds.GetLast()
}

func main() {
	ctx := context.Background()

	// Actor-based DataStore.
	actorSvc := NewDataStoreService(ctx, 100)
	actorSvc.AddValue(42)
	fmt.Println("Actor-based last value:", actorSvc.GetLastValue())

	// Mutex-based DataStore.
	mutexSvc := NewMutexDataStoreService(100)
	mutexSvc.AddValue(42)
	fmt.Println("Mutex-based last value:", mutexSvc.GetLastValue())
}
```

### Explanation

- **DataStore:**  
  A simple ring buffer that supports O(1) operations for inserting a value and getting the last value.

- **DataStoreService (Actor-based):**  
  Wraps the DataStore inside an actor. This model safely serializes access to the data structure without explicit locks, potentially reducing contention overhead.

- **MutexDataStoreService (Mutex-based):**  
  Uses a `sync.RWMutex` to protect concurrent access. `AddValue` acquires a write lock, while `GetLastValue` uses a read lock. Proper `defer` usage ensures lock release.

## Words of Caution

While **Action** simplifies concurrent programming, it's important to understand the boundaries of what it does (and doesn't) manage for you:

- **No built-in cancellation:**  
  Once an action is enqueued, it will run. If cancellation is important, check `r.Ctx().Done()` inside your action.

- **Timeouts are user-defined:**  
  The library does not impose timeouts or deadlines on action execution. Use `context.WithTimeout(r.Ctx(), ...)` if needed.

- **No retry or panic recovery:**  
  Actions are executed as-is. If you need retries or want to catch panics, you must wrap that logic in your action.

- **Do not send to a stopped runner:**  
  Sending to a runner after its context has been canceled will panic. Use `r.Ctx().Err()` to check if the runner is still alive.

- **Design actions to be idempotent when possible:**  
  Especially when coordinating across multiple actors, using idempotent or safe-to-discard actions can simplify error recovery and retries.

- **Tradeoff between control and simplicity:**  
  This library favors predictable, serialized execution over fine-grained parallel control. It is not a one-size-fits-all tool but excels where reduced mental overhead is preferred over maximizing throughput.

## Documentation

For full API documentation and usage details, please visit the [GoDoc](https://pkg.go.dev/github.com/neonima/action) page.

## Benchmarks

This repository includes benchmarks comparing actor-based synchronization with traditional mutex-based approaches. For example, simple get/set operations tend to be faster with mutexes, while complex workloads involving sophisticated data structures can narrow—and sometimes reverse—the performance gap.

### Sample Benchmark Results

```plaintext
BenchmarkGetSet
BenchmarkGetSet-8               	  450556	      2668 ns/op
BenchmarkMutexGetSet
BenchmarkMutexGetSet-8          	 1078102	      1140 ns/op

BenchmarkComplexGetSet
BenchmarkComplexGetSet-8        	   41164	     24861 ns/op
BenchmarkComplexMutexGetSet
BenchmarkComplexMutexGetSet-8   	   54876	     29796 ns/op
```

Run the benchmarks using:

```bash
go test -bench=.
```

## Contributing

Contributions are welcome! If you have ideas, bug fixes, or improvements:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Commit your changes.
4. Open a pull request with a detailed explanation of your changes.

For major changes, please open an [issue](https://github.com/neonima/action/issues) first to discuss what you would like to modify.

## Potential Improvements

The following improvements are being considered while preserving the library’s current vision: keeping it minimal, intentional, and free from unnecessary complexity.

- **Context-aware wrappers**: Helpers that reduce boilerplate for timeout-aware or cancellation-sensitive actions.
- **Optional panic recovery**: Tools to isolate and log panics within actions without compromising runner stability.
- **Queue drain control**: Graceful shutdown utilities like `Flush()` or `DrainAndStop()` for smoother exits.
- **Observability hooks**: Lightweight hooks or interfaces to integrate queue stats or action timings into existing monitoring systems.

We’re intentionally avoiding features that would increase cognitive overhead or compromise the simplicity of the actor model. Feedback and ideas are welcome — especially if they fit within this philosophy.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

