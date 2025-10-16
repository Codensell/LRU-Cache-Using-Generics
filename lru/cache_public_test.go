package lru_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/Codensell/LRU-Cache-Using-Generics/lru"
)

var newSI = lru.NewCache[string, int]
var newII = lru.NewCache[int, int]

func Test_NewCache_PanicOnNonPositiveCapacity_Public(t *testing.T) {
	cases := []int{0, -1, -10}
	for _, capVal := range cases {
		t.Run("cap_"+strconv.Itoa(capVal), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for capacity %d", capVal)
				}
			}()
			_ = newSI(capVal)
		})
	}
}

func Test_LenCap_Public(t *testing.T) {
	c := newSI(3)
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

func Test_SetGet_Peek_Delete_Public(t *testing.T) {
	type step struct {
		op    string
		key   string
		value int
		hit   bool
		want  int
		deleted bool
	}
	cases := []struct {
		name     string
		capacity int
		steps    []step
	}{
		{
			name:     "basic_get_set_peek_delete",
			capacity: 2,
			steps: []step{
				{op: "get", key: "x", hit: false, want: 0},
				{op: "set", key: "a", value: 1},
				{op: "peek", key: "a", hit: true, want: 1},
				{op: "set", key: "b", value: 2},
				{op: "get", key: "a", hit: true, want: 1},
				{op: "del", key: "a", deleted: true},
				{op: "get", key: "a", hit: false, want: 0},
			},
		},
		{
			name:     "peek_does_not_affect_eviction",
			capacity: 2,
			steps: []step{
				{op: "set", key: "a", value: 1},
				{op: "set", key: "b", value: 2},
				{op: "peek", key: "a", hit: true, want: 1},
				{op: "set", key: "c", value: 3},
				{op: "get", key: "a", hit: false, want: 0},
				{op: "get", key: "b", hit: true, want: 2},
				{op: "get", key: "c", hit: true, want: 3},
			},
		},
		{
			name:     "delete_unknown_is_false",
			capacity: 1,
			steps: []step{
				{op: "del", key: "nope", deleted: false},
				{op: "set", key: "x", value: 7},
				{op: "del", key: "x", deleted: true},
				{op: "del", key: "x", deleted: false},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newSI(tc.capacity)
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
				case "peek":
					got, ok := c.Peek(s.key)
					if ok != s.hit {
						t.Fatalf("step %d peek(%q) hit=%v want=%v", i, s.key, ok, s.hit)
					}
					if got != s.want {
						t.Fatalf("step %d peek(%q) val=%v want=%v", i, s.key, got, s.want)
					}
				case "del":
					ok := c.Delete(s.key)
					if ok != s.deleted {
						t.Fatalf("step %d delete(%q) = %v want %v", i, s.key, ok, s.deleted)
					}
				default:
					t.Fatalf("unknown op %q", s.op)
				}
			}
		})
	}
}

func Test_Eviction_Public(t *testing.T) {
	c := newSI(3)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	_, _ = c.Get("a")
	_, _ = c.Get("c")
	c.Set("d", 4)
	if _, ok := c.Get("b"); ok {
		t.Fatalf("expected b evicted")
	}
	for k, v := range map[string]int{"a": 1, "c": 3, "d": 4} {
		got, ok := c.Get(k)
		if !ok || got != v {
			t.Fatalf("expected %q=%d", k, v)
		}
	}
}

func Test_Clear_Public(t *testing.T) {
	c := newSI(2)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("expected empty after Clear")
	}
	if _, ok := c.Get("a"); ok {
		t.Fatalf("unexpected hit after Clear")
	}
}

func Test_Concurrent_Public(t *testing.T) {
    workers := 8
    per := 2000
    cap := workers * per

    c := newII(cap)

    var wg sync.WaitGroup
    wg.Add(workers)
    for g := 0; g < workers; g++ {
        go func(id int) {
            defer wg.Done()
            base := id * per
            for i := 0; i < per; i++ {
                k := base + i
                c.Set(k, k+1)
                if v, ok := c.Get(k); !ok || v != k+1 {
                    t.Fatalf("unexpected get %d", k)
                }
            }
        }(g)
    }
    wg.Wait()
    if c.Len() > c.Cap() {
        t.Fatalf("len exceeds cap")
    }
}
