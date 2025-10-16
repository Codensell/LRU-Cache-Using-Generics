# LRU-Cache-Using-Generics

## Implement a Least Recently Used (LRU) cache using Go generics.

#### Focus: synchronization primitives (sync.Mutex, sync.Map), generic types, concurrent access safety.

Expected Result:
- Cache automatically evicts the least recently used entries after reaching capacity.
- Concurrent read/write operations are handled safely.
- The structure is reusable for any key/value types via generics.

#### Overview

A thread-safe Least Recently Used (LRU) cache implemented in Go using generics.
The cache automatically evicts the least recently used entries once the capacity limit is reached.
All operations (Set, Get, Peek, Delete) have O(1) amortized time complexity and are safe for concurrent access.

#### Installation

go get github.com/Codensell/LRU-Cache-Using-Generics/lru

#### API Reference

NewCache[K comparable, V any](capacity int) *Cache[K, V]

(*Cache[K,V]).Set(key K, value V)
(*Cache[K,V]).Get(key K) (V, bool)
(*Cache[K,V]).Peek(key K) (V, bool)
(*Cache[K,V]).Delete(key K) bool
(*Cache[K,V]).Clear()
(*Cache[K,V]).Len() int
(*Cache[K,V]).Cap() int

#### Behavior

- Get moves the accessed element to the front (most recently used).
- Peek retrieves a value without changing order.
- If capacity is exceeded, the oldest (LRU) element is evicted.
- NewCache panics if capacity <= 0.
- Get and Peek return the zero value of V and false if the key is missing.
- Delete returns true if the key was present and removed.

#### Concurrency

- Internally protected by a single sync.Mutex.
- Safe for concurrent reads and writes (go test -race passes).
- For high-throughput workloads, this design can be extended via sharding (multiple independent LRU segments).

#### Testing

Two test suites are included:

- Internal tests (package lru) validate internal mechanics.
- Public API tests (package lru_test) ensure external behavior stability.

go test ./...
go test -race ./...

#### Future Improvements

- Sharded design to reduce contention under heavy load.
- OnEvict hooks for metrics or cleanup.
- Dynamic Resize and TTL/expiry support.
- Optional constructor with error return instead of panic.