package dlock

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrLockAcquireFailed 表示锁获取失败
	ErrLockAcquireFailed = errors.New("failed to acquire lock")
	// ErrLockReleaseFailed 表示锁释放失败
	ErrLockReleaseFailed = errors.New("failed to release lock")
)

// RedisLock 是一个基于 Redis 的分布式锁实现
type redisLock struct {
	client     *redis.Client
	key        string
	value      string
	expiration time.Duration
}

// NewRedisLock 创建一个新的 Redis 分布式锁
func NewRedisLock(client *redis.Client, key, value string, expiration time.Duration) *redisLock {
	return &redisLock{
		client:     client,
		key:        key,
		value:      value,
		expiration: expiration,
	}
}

// TryLock 尝试获取锁，立即返回结果
func (rl *redisLock) TryLock(ctx context.Context) error {
	// 使用 Redis SET NX 命令尝试设置锁
	// NX 表示只有当 key 不存在时才会设置成功
	success, err := rl.client.SetNX(ctx, rl.key, rl.value, rl.expiration).Result()
	if err != nil {
		return err
	}

	if !success {
		return ErrLockAcquireFailed
	}

	return nil
}

// WaitLock 获取锁，如果获取失败会重试直到超时
func (rl *redisLock) WaitLock(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		// 尝试获取锁
		err := rl.TryLock(ctx)
		if err == nil {
			return nil
		}

		if err != ErrLockAcquireFailed {
			return err
		}

		// 检查是否超时
		if time.Now().After(deadline) {
			return ErrLockAcquireFailed
		}

		// 等待一段时间后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			// 继续尝试
		}
	}
}

// Unlock 释放锁
func (rl *redisLock) Unlock(ctx context.Context) error {
	// 使用 Lua 脚本保证原子性操作
	// 只有当锁的值匹配时才释放锁，防止释放其他客户端的锁
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`

	result, err := rl.client.Eval(ctx, script, []string{rl.key}, rl.value).Result()
	if err != nil {
		return err
	}

	if result.(int64) != 1 {
		return ErrLockReleaseFailed
	}

	return nil
}

// Refresh 刷新锁的过期时间
func (rl *redisLock) Refresh(ctx context.Context) error {
	// 使用 Lua 脚本保证原子性操作
	// 只有当锁的值匹配时才刷新过期时间
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1], ARGV[2])
	else
		return 0
	end
	`

	millis := int64(rl.expiration / time.Millisecond)
	result, err := rl.client.Eval(ctx, script, []string{rl.key}, rl.value, millis).Result()
	if err != nil {
		return err
	}

	if result.(int64) != 1 {
		return ErrLockReleaseFailed
	}

	return nil
}
