# Redis 分布式锁

这是一个基于 Redis 实现的分布式锁库，用于在分布式系统中协调对共享资源的访问。

## 功能特点

- 基于 Redis 的 SET NX 命令实现分布式锁
- 支持锁超时自动释放，防止死锁
- 提供立即尝试和超时重试两种锁获取模式
- 使用 Lua 脚本确保锁释放的原子性和正确性
- 支持锁的过期时间刷新

## 使用方法

### 安装

```bash
go get github.com/yourproject/distributedlock
```

### 基础用法

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/yourproject/distributedlock"
)

func main() {
    // 创建 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // 创建分布式锁
    // 参数: Redis客户端, 锁名称, 锁值(通常是唯一ID), 锁过期时间
    lock := distributedlock.NewRedisLock(rdb, "my-resource-lock", "client-1", 30*time.Second)

    ctx := context.Background()

    // 尝试获取锁
    err := lock.TryLock(ctx)
    if err != nil {
        if err == distributedlock.ErrLockAcquireFailed {
            log.Println("资源已被锁定")
        } else {
            log.Printf("锁获取错误: %v", err)
        }
        return
    }

    log.Println("成功获取锁，开始处理资源...")

    // 在这里处理受保护的资源

    // 完成后释放锁
    if err := lock.Unlock(ctx); err != nil {
        log.Printf("锁释放错误: %v", err)
    }
}
```

### 等待锁直到超时

```go
// 尝试获取锁，最多等待5秒
err := lock.Lock(ctx, 5*time.Second)
if err != nil {
    log.Printf("锁获取失败: %v", err)
    return
}

// 处理资源...

// 释放锁
lock.Unlock(ctx)
```

### 刷新锁过期时间

```go
// 创建锁
lock := distributedlock.NewRedisLock(rdb, "my-lock", "client-1", 10*time.Second)

// 获取锁
err := lock.TryLock(ctx)
if err != nil {
    return err
}

// 启动一个 goroutine 定期刷新锁
go func() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := lock.Refresh(ctx); err != nil {
                log.Printf("锁刷新失败: %v", err)
                return
            }
        case <-ctx.Done():
            return
        }
    }
}()

// 执行长时间运行的任务...

// 完成后释放锁
lock.Unlock(ctx)
```

## 注意事项

1. **避免死锁**: 始终设置合理的锁过期时间，并在操作完成后主动释放锁。
2. **唯一标识**: 锁的值应该是调用者的唯一标识，避免一个客户端释放了另一个客户端的锁。
3. **锁续期**: 对于长时间运行的任务，应考虑定期刷新锁的过期时间。
4. **失败处理**: 总是检查锁获取和释放的错误，并进行适当的处理。
5. **网络问题**: 分布式锁受网络延迟和分区的影响，应在应用层面考虑这些因素。
