package lru

import (
	"strconv"
	"sync"
	"testing"
)

func Test_NewCache_PanicOnNonPositiveCapacity(t *testing.T) {
	cases := []int{0, -1, -10}
	for _, cap := range cases {
		t.Run("cap_"+strconv.Itoa(cap), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for capacity %d", cap)
				}
			}()
			_ = NewCache[string, int](cap)
		})
	}
}

func Test_LenCap(t *testing.T) {
	c := NewCache[string, int](3)
	if c.Cap() != 3 {
		t.Fatalf("Cap mismatch: got %d want %d", c.Cap(), 3)
	}
	if c.Len() != 0 {
		t.Fatalf("Len mismatch: got %d want %d", c.Len(), 0)
	}
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Fatalf("Len mismatch: got %d want %d", c.Len(), 2)
	}
}

func Test_SetGet_BasicAndMoveToFront(t *testing.T) {
	type step struct {
		op    string
		key   string
		value int
		hit   bool
		want  int
	}
	cases := []struct {
		name     string
		capacity int
		steps    []step
	}{
		{
			name:     "basic_hit_and_miss",
			capacity: 2,
			steps: []step{
				{op: "get", key: "x", hit: false, want: 0},
				{op: "set", key: "a", value: 1},
				{op: "get", key: "a", hit: true, want: 1},
				{op: "set", key: "b", value: 2},
				{op: "get", key: "b", hit: true, want: 2},
			},
		},
		{
			name:     "update_existing_moves_to_front",
			capacity: 3,
			steps: []step{
				{op: "set", key: "a", value: 1},
				{op: "set", key: "b", value: 2},
				{op: "set", key: "c", value: 3},
				{op: "set", key: "a", value: 10},
				{op: "set", key: "d", value: 4},
				{op: "get", key: "a", hit: true, want: 10},
				{op: "get", key: "b", hit: false, want: 0},
				{op: "get", key: "c", hit: true, want: 3},
				{op: "get", key: "d", hit: true, want: 4},
			},
		},
		{
			name:     "get_moves_to_front_affects_eviction",
			capacity: 3,
			steps: []step{
				{op: "set", key: "a", value: 1},
				{op: "set", key: "b", value: 2},
				{op: "set", key: "c", value: 3},
				{op: "get", key: "a", hit: true, want: 1},
				{op: "set", key: "d", value: 4},
				{op: "get", key: "b", hit: false, want: 0},
				{op: "get", key: "a", hit: true, want: 1},
				{op: "get", key: "c", hit: true, want: 3},
				{op: "get", key: "d", hit: true, want: 4},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCache[string, int](tc.capacity)
			for i, s := range tc.steps {
				switch s.op {
				case "set":
					c.Set(s.key, s.value)
				case "get":
					got, ok := c.Get(s.key)
					if ok != s.hit {
						t.Fatalf("step %d get(%q) hit=%v want=%v", i, s.key, ok, s.hit)
					}
					if got != s.want {
						t.Fatalf("step %d get(%q) val=%v want=%v", i, s.key, got, s.want)
					}
				default:
					t.Fatalf("unknown op %q", s.op)
				}
			}
		})
	}
}

func Test_Evict_RarelyUsed_Items(t *testing.T) {
	type access struct {
		op    string
		key   string
		value int
	}
	seq := []access{
		{op: "set", key: "a", value: 1},
		{op: "set", key: "b", value: 2},
		{op: "set", key: "c", value: 3},
		{op: "get", key: "a"},
		{op: "get", key: "c"},
		{op: "set", key: "d", value: 4},
	}
	c := NewCache[string, int](3)
	for i, s := range seq {
		switch s.op {
		case "set":
			c.Set(s.key, s.value)
		case "get":
			_, _ = c.Get(s.key)
		default:
			t.Fatalf("step %d unknown op", i)
		}
	}
	if _, ok := c.Get("b"); ok {
		t.Fatalf("expected b to be evicted")
	}
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("expected a present with 1")
	}
	if v, ok := c.Get("c"); !ok || v != 3 {
		t.Fatalf("expected c present with 3")
	}
	if v, ok := c.Get("d"); !ok || v != 4 {
		t.Fatalf("expected d present with 4")
	}
}

func Test_Evict_On_Capacity_Exceeded(t *testing.T) {
	cases := []struct {
		name     string
		capacity int
		sets     []struct {
			k string
			v int
		}
		expectPresent []string
		expectEvicted []string
	}{
		{
			name:     "cap_2_simple_chain",
			capacity: 2,
			sets: []struct {
				k string
				v int
			}{
				{"a", 1}, {"b", 2}, {"c", 3},
			},
			expectPresent: []string{"b", "c"},
			expectEvicted: []string{"a"},
		},
		{
			name:     "cap_3_with_access",
			capacity: 3,
			sets: []struct {
				k string
				v int
			}{
				{"a", 1}, {"b", 2}, {"c", 3},
				{"b", 20}, {"d", 4},
			},
			expectPresent: []string{"b", "c", "d"},
			expectEvicted: []string{"a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCache[string, int](tc.capacity)
			for _, p := range tc.sets {
				c.Set(p.k, p.v)
			}
			for _, k := range tc.expectPresent {
				if _, ok := c.Get(k); !ok {
					t.Fatalf("expected %q present", k)
				}
			}
			for _, k := range tc.expectEvicted {
				if _, ok := c.Get(k); ok {
					t.Fatalf("expected %q evicted", k)
				}
			}
			if c.Len() != tc.capacity {
				t.Fatalf("Len mismatch: got %d want %d", c.Len(), tc.capacity)
			}
		})
	}
}

func Test_Clear_EmptiesCache(t *testing.T) {
	c := NewCache[string, int](3)
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Fatalf("precondition failed")
	}
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("expected empty after Clear")
	}
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected miss after Clear")
	}
	if _, ok := c.Get("b"); ok {
		t.Fatalf("expected miss after Clear")
	}
}

func Test_Concurrent_Access_Basic(t *testing.T) {
	c := NewCache[int, int](1000)
	var wg sync.WaitGroup
	n := 8
	per := 1000
	wg.Add(n)
	var errs []string
	var mu sync.Mutex
	for g := 0; g < n; g++ {
		go func(id int) {
			defer wg.Done()
			base := id * per
			for i := 0; i < per; i++ {
				k := base + i
				c.Set(k, k+1)
				if v, ok := c.Get(k); !ok || v != k+1 {
					mu.Lock()
					errs = append(errs, "unexpected get for "+strconv.Itoa(k))
					mu.Unlock()
				}
			}
		}(g)
	}
	wg.Wait()
	if len(errs) > 0 {
		t.Fatalf("errors: %v", errs)
	}
	if c.Len() > c.Cap() {
		t.Fatalf("len exceeds cap")
	}
}
