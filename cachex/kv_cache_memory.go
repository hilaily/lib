package cachex

import (
	"time"

	"github.com/hilaily/kit/mapx"
)

var (
	_ IKVCache[any] = &memoryKVCache[any]{}
)

func NewKVCacheFromMemory[T any](timeout time.Duration) *memoryKVCache[T] {
	return &memoryKVCache[T]{
		data: mapx.NewCacheMap2[T](timeout, true),
	}
}

type memoryKVCache[T any] struct {
	data *mapx.CacheMap2[T]
}

func (c *memoryKVCache[T]) Set(key string, value T) error {
	c.data.Set(key, value)
	return nil
}

func (c *memoryKVCache[T]) Get(key string) (T, bool, error) {
	v, ok := c.data.Get(key)
	return v, ok, nil
}

func (c *memoryKVCache[T]) SetWithTime(key string, value T, t time.Time) {
	c.data.SetWithTime(key, value, t)
}

func (c *memoryKVCache[T]) Del(key string) {
	c.data.Del(key)
}
