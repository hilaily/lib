package dlock

import (
	"context"
	"time"
)

type IDLock interface {
	// TryLock 尝试获取锁，立即返回结果
	TryLock(ctx context.Context) error
	// WaitLock 获取锁，如果获取失败会重试直到超时
	WaitLock(ctx context.Context, timeout time.Duration) error
	// Unlock 释放锁
	Unlock(ctx context.Context) error
	// Refresh 刷新锁的过期时间
	Refresh(ctx context.Context) error
}
