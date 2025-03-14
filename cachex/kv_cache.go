package cachex

import "time"

type IKVCache[T any] interface {
	Set(key string, value T) error
	SetWithTime(k string, v T, t time.Time)
	Get(key string) (T, bool, error)
	Del(k string)
}
