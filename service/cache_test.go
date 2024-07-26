package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Entry struct {
	key   interface{}
	value interface{}
	ttl   time.Duration
}

func TestAdd(t *testing.T) {
	cache := NewICache(3)
	testTable := []struct {
		name            string
		data            []Entry
		tts             time.Duration // time to sleep
		expecetedLength int
	}{
		{
			name: "Overload",
			data: []Entry{
				{key: 1, value: "WB1"}, {key: 2, value: "WB2"}, {key: 3, value: "WB3"},
				{key: 4, value: "WB4"},
			},
			expecetedLength: 3,
		},
		{
			name: "TTL-1 (remove)",
			data: []Entry{
				{key: 1, value: "WB1"}, {key: 3, value: "WB3", ttl: 1 * time.Second},
			},
			tts:             2 * time.Second,
			expecetedLength: 1,
		},
		{
			name: "Different type of data",
			data: []Entry{
				{key: "George", value: time.Thursday}, {key: 4.4, value: 555},
			},
			expecetedLength: 2,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < len(tc.data); i++ {
				if tc.data[i].ttl != 0 {
					cache.AddWithTTL(tc.data[i].key, tc.data[i].value, tc.data[i].ttl)
					continue
				}
				cache.Add(tc.data[i].key, tc.data[i].value)
			}
			time.Sleep(tc.tts)
			assert.Equal(t, tc.expecetedLength, cache.Len())
			cache.Clear()
		})
	}
}

func TestGet(t *testing.T) {
	cache := NewICache(3)

	cache.Add(44, "Lewis")
	cache.Add(5.5, time.Thursday)
	cache.Add("Russell", 63)

	testTable := []struct {
		name            string
		key             interface{}
		expecetedValue  interface{}
		expecetedStatus bool
	}{
		{
			name:            "Get by key",
			key:             44,
			expecetedValue:  "Lewis",
			expecetedStatus: true,
		},
		{
			name:            "Get none",
			key:             45,
			expecetedValue:  nil,
			expecetedStatus: false,
		},
		{
			name:            "Get by key 2",
			key:             5.5,
			expecetedValue:  time.Thursday,
			expecetedStatus: true,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			actualValue, actualStatus := cache.Get(tc.key)
			assert.Equal(t, tc.expecetedValue, actualValue)
			assert.Equal(t, tc.expecetedStatus, actualStatus)
		})
	}
}

func TestRemove(t *testing.T) {
	cache := NewICache(3)

	testTable := []struct {
		name            string
		key             interface{}
		expecetedLength int
	}{
		{
			name:            "Remove by key",
			key:             44,
			expecetedLength: 2,
		},
		{
			name:            "Remove nothing",
			key:             "abcd",
			expecetedLength: 3,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			cache.Add(44, "Lewis")
			cache.Add(5.5, time.Thursday)
			cache.Add("Russell", 63)

			cache.Remove(tc.key)
			if _, ok := cache.Get(tc.key); ok {
				t.Error("data had to be deleted")
			}
			assert.Equal(t, tc.expecetedLength, cache.Len())
		})
	}
}

func TestLRUOrder(t *testing.T) {
	cache := NewICache(3)

	cache.Add("Russell", 63)
	NewValue := 64
	cache.Add("Russell", NewValue)
	cache.Add(44, "Lewis")
	cache.Add(5.5, time.Thursday)

	cache.Get("Russell") // move this node in front, now node with key=44 is a last one

	cache.Add(time.Friday, "John") // overload -> node with key=44 has to be deleted

	if _, ok := cache.Get(44); ok {
		t.Errorf("data had to be deleted but exsist")
	}

	val, ok := cache.Get("Russell")
	if !ok {
		t.Errorf("data has to exists but doesn't")
	}

	if val != NewValue {
		t.Errorf("data has wrong value, expeceted %d, but got %d", NewValue, val)
	}

}

func TestConcurrency(t *testing.T) {
	cap := 3

	cache := NewICache(cap)
	wg := &sync.WaitGroup{}

	// fill up the cache
	for i := range 100 {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			val := fmt.Sprintf("Value #%d", i)
			cache.Add(i, val)
			length := cache.Len()
			if length > cap {
				t.Errorf("Limit of capacity was exceeded: [%d]", length)
			}
		}()
	}

	wg.Wait()
}
