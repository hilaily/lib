package dlock

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func Example() {
	// 创建一个 Redis 客户端
	// 注意：这里只是示例，实际使用时需要配置正确的 Redis 地址
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建上下文
	ctx := context.Background()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("无法连接到 Redis: %v", err)
	}

	// 创建一个分布式锁
	resourceKey := "example-resource"
	clientID := "client-1"
	lockTimeout := 30 * time.Second

	lock := NewRedisLock(rdb, resourceKey, clientID, lockTimeout)

	// 尝试获取锁（非阻塞方式）
	err = lock.TryLock(ctx)
	if err != nil {
		fmt.Printf("无法获取锁: %v\n", err)
		return
	}

	fmt.Println("成功获取锁")

	// 模拟处理受保护的资源
	fmt.Println("正在处理受保护的资源...")
	time.Sleep(2 * time.Second)

	// 释放锁
	if err := lock.Unlock(ctx); err != nil {
		fmt.Printf("释放锁失败: %v\n", err)
		return
	}

	fmt.Println("锁已释放")
}

func ExampleRedisLock_WaitLock() {
	// 创建一个 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	// 创建分布式锁
	lock1 := NewRedisLock(rdb, "concurrent-resource", "client-1", 10*time.Second)
	lock2 := NewRedisLock(rdb, "concurrent-resource", "client-2", 10*time.Second)

	// 先让 client-1 获取锁
	if err := lock1.TryLock(ctx); err != nil {
		fmt.Printf("client-1 获取锁失败: %v\n", err)
		return
	}

	fmt.Println("client-1 获取锁成功")

	var wg sync.WaitGroup
	wg.Add(1)

	// 并发尝试获取已被锁定的资源
	go func() {
		defer wg.Done()

		fmt.Println("client-2 尝试获取锁，最多等待 5 秒")

		// 使用 AcquireWithTimeout 方法等待锁，最多等待 5 秒
		err := lock2.WaitLock(ctx, 5*time.Second)
		if err != nil {
			fmt.Printf("client-2 获取锁失败: %v\n", err)
			return
		}

		fmt.Println("client-2 获取锁成功")

		// 使用锁
		fmt.Println("client-2 正在处理资源...")
		time.Sleep(1 * time.Second)

		// 释放锁
		if err := lock2.Unlock(ctx); err != nil {
			fmt.Printf("client-2 释放锁失败: %v\n", err)
		} else {
			fmt.Println("client-2 释放锁成功")
		}
	}()

	// 模拟 client-1 持有锁 3 秒后释放
	fmt.Println("client-1 持有锁 3 秒")
	time.Sleep(3 * time.Second)

	if err := lock1.Unlock(ctx); err != nil {
		fmt.Printf("client-1 释放锁失败: %v\n", err)
	} else {
		fmt.Println("client-1 释放锁成功")
	}

	// 等待所有协程完成
	wg.Wait()
}

func ExampleRedisLock_Refresh() {
	// 创建一个 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 创建一个短期锁 (5秒)
	lock := NewRedisLock(rdb, "refreshable-resource", "long-task-client", 5*time.Second)

	// 获取锁
	if err := lock.TryLock(ctx); err != nil {
		fmt.Printf("获取锁失败: %v\n", err)
		return
	}

	fmt.Println("获取锁成功，锁将在 5 秒后过期")

	// 创建一个通道用于停止刷新
	stopRefresh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	// 启动一个协程定期刷新锁
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := lock.Refresh(ctx); err != nil {
					fmt.Printf("刷新锁失败: %v\n", err)
					return
				}
				fmt.Println("锁已刷新，过期时间延长 5 秒")
			case <-stopRefresh:
				fmt.Println("停止刷新锁")
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// 模拟长时间任务 (10秒)
	fmt.Println("开始长时间任务...")
	for i := 1; i <= 10; i++ {
		select {
		case <-time.After(1 * time.Second):
			fmt.Printf("任务运行中: %d 秒\n", i)
		case <-ctx.Done():
			fmt.Println("任务被取消")
			return
		}
	}
	fmt.Println("任务完成")

	// 停止锁刷新并释放锁
	close(stopRefresh)
	wg.Wait()

	if err := lock.Unlock(ctx); err != nil {
		fmt.Printf("释放锁失败: %v\n", err)
	} else {
		fmt.Println("锁已释放")
	}
}
