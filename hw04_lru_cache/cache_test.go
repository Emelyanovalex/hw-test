package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)

		c.Set("1", 1)
		c.Set("2", 2)
		c.Set("3", 3)
		c.Set("4", 4)

		_, ok1 := c.Get("1")
		val2, ok2 := c.Get("2")

		require.False(t, ok1)
		require.True(t, ok2)
		require.Equal(t, 2, val2)
	})

	t.Run("purge logic lru", func(t *testing.T) {
		c := NewCache(3)

		c.Set("A", "A1")
		c.Set("B", "B1")
		c.Set("C", "C1")

		c.Get("B")       // [B,C,A]
		c.Set("A", "A2") // [A,B,C]
		c.Get("C")       // [C,A,B]
		c.Set("D", "D1") // [D,C,A]

		_, okB := c.Get("B")

		require.False(t, okB)
	})

	t.Run("clear cache", func(t *testing.T) {
		c := NewCache(5)
		c.Set("key1", 1)
		c.Set("key2", 2)

		c.Clear()

		_, ok1 := c.Get("key1")
		_, ok2 := c.Get("key2")
		require.False(t, ok1)
		require.False(t, ok2)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
