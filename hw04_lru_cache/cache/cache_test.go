package cache

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
		c := NewCache(2)

		require.False(t, c.Set("first", 1))
		require.False(t, c.Set("second", 2))

		val, ok := c.Get("first")
		require.True(t, ok)
		require.Equal(t, 1, val)

		require.False(t, c.Set("third", 3))

		val, ok = c.Get("second")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("first")
		require.True(t, ok)
		require.Equal(t, 1, val)

		val, ok = c.Get("third")
		require.True(t, ok)
		require.Equal(t, 3, val)
	})

	t.Run("update existing item does not evict another key", func(t *testing.T) {
		c := NewCache(2)

		require.False(t, c.Set("first", 1))
		require.False(t, c.Set("second", 2))
		require.True(t, c.Set("first", 10))

		val, ok := c.Get("first")
		require.True(t, ok)
		require.Equal(t, 10, val)

		val, ok = c.Get("second")
		require.True(t, ok)
		require.Equal(t, 2, val)
	})

	t.Run("clear removes all items", func(t *testing.T) {
		c := NewCache(2)

		require.False(t, c.Set("first", 1))
		require.False(t, c.Set("second", 2))

		c.Clear()

		val, ok := c.Get("first")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("second")
		require.False(t, ok)
		require.Nil(t, val)
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
