
# Action: A Go Actor Model Library

**Action** is a lightweight actor model library for Go that simplifies concurrency by encapsulating state and message handling within actors. This approach improves code maintainability and safety in high-concurrency scenarios.

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

// MutexDataStoreService uses a mutex to protect concurrent access to DataStore.
type MutexDataStoreService struct {
	ds  *DataStore
	mtx sync.Mutex
}

// NewMutexDataStoreService creates a new MutexDataStoreService.
func NewMutexDataStoreService(capacity int) *MutexDataStoreService {
	return &MutexDataStoreService{
		ds: NewDataStore(capacity),
	}
}

// AddValue adds a value using mutex protection.
func (s *MutexDataStoreService) AddValue(val int) {
	s.mtx.Lock()
	s.ds.Add(val)
	s.mtx.Unlock()
}

// GetLastValue retrieves the last value using mutex protection.
func (s *MutexDataStoreService) GetLastValue() int {
	s.mtx.Lock()
	defer s.mtx.Unlock()
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
  Uses a standard mutex to protect DataStore. Although mutex operations are very fast for simple tasks, the overhead may become more noticeable under high contention or complex operations.

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

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.