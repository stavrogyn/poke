package pokecache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache := NewCache(time.Second * 1)
	cache.StartReapLoop(time.Second * 1)
	cache.Add("test", []byte("test"))
	if _, exists := cache.Get("test"); !exists {
		t.Errorf("test value should exist")
	}
}

func TestCacheReap(t *testing.T) {
	cache := NewCache(time.Millisecond * 10)
	cache.StartReapLoop(time.Millisecond * 10)
	cache.Add("test", []byte("test"))
	if _, exists := cache.Get("test"); !exists {
		t.Errorf("test value should exist")
	}
	time.Sleep(time.Millisecond * 10)
	if _, exists := cache.Get("test"); exists {
		t.Errorf("test value should not exist")
	}
}

func TestCacheStartReapLoop(t *testing.T) {
	cache := NewCache(time.Millisecond * 10)
	cache.StartReapLoop(time.Millisecond * 10)
	cache.Add("test", []byte("test"))
	if _, exists := cache.Get("test"); !exists {
		t.Errorf("test value should exist")
	}
	time.Sleep(time.Millisecond * 10)
	if _, exists := cache.Get("test"); exists {
		t.Errorf("test value should not exist")
	}
}
