# LRU-Cache-Using-Generics

## Implement a Least Recently Used (LRU) cache using Go generics.

Focus: synchronization primitives (sync.Mutex, sync.Map), generic types, concurrent access safety.

Expected Result:
- Cache automatically evicts the least recently used entries after reaching capacity.
- Concurrent read/write operations are handled safely.
- The structure is reusable for any key/value types via generics.
