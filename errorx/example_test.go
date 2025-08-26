package errorx

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorx(t *testing.T) {
	err := New(100000, "not found")
	e := err.SetErr(errors.New("user not found")).WithExtra("user_id", 123).WithStack()
	en, _ := e.MarshalJSON()
	fmt.Println(string(en))
}

// 测试并发安全性
func TestConcurrentSafety(t *testing.T) {
	done := make(chan bool, 100)

	// 启动100个goroutine同时使用同一个错误模板
	for i := 0; i < 100; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 每个goroutine独立修改错误消息
			err := NotFound.Clone().SetMsg("用户 %d 不存在", id)

			// 验证错误消息是否正确
			t.Log(err)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 100; i++ {
		<-done
	}
}

// 测试错误类型检查
func TestErrorTypes(t *testing.T) {
	err1 := NotFound.Clone().SetMsg("资源不存在")
	err2 := Internal.Clone().SetMsg("服务器错误")

	// 使用 errors.Is 进行类型检查
	if !errors.Is(err1, NotFound) {
		t.Error("err1 should be NotFound type")
	}

	if errors.Is(err1, Internal) {
		t.Error("err1 should not be Internal type")
	}

	if !errors.Is(err2, Internal) {
		t.Error("err2 should be Internal type")
	}
}

// 测试HTTP状态码映射
// func TestHTTPStatusCode(t *testing.T) {
// 	testCases := []struct {
// 		err      *Err
// 		expected int
// 	}{
// 		{BadRequest, 400},
// 		{Auth, 401},
// 		{PermissionDenied, 403},
// 		{NotFound, 404},
// 		{AlreadyExists, 409},
// 		{Internal, 500},
// 		{Timeout, 504},
// 	}

// 	for _, tc := range testCases {
// 		if got := tc.err.GetHTTPStatusCode(); got != tc.expected {
// 			t.Errorf("Expected HTTP status %d for error %d, got %d",
// 				tc.expected, tc.err.Code, got)
// 		}
// 	}
// }

// 测试Connect错误码转换
func TestConnectCodeMapping(t *testing.T) {
	testCases := []struct {
		err      *Err
		expected int
	}{
		{BadRequest, 3},       // INVALID_ARGUMENT
		{Auth, 16},            // UNAUTHENTICATED
		{PermissionDenied, 7}, // PERMISSION_DENIED
		{NotFound, 5},         // NOT_FOUND
		{AlreadyExists, 6},    // ALREADY_EXISTS
		{Canceled, 1},         // CANCELLED
		{Internal, 13},        // INTERNAL
		{Timeout, 4},          // DEADLINE_EXCEEDED
	}

	for _, tc := range testCases {
		if got := new(ConnectRPCAdapter).ToConnectCode(tc.err); int(got) != tc.expected {
			t.Errorf("Expected Connect code %d for error %d, got %d",
				tc.expected, tc.err.code, got)
		}
	}
}

// 基准测试：错误创建性能
func BenchmarkErrorCreation(b *testing.B) {
	b.Run("Clone", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NotFound.Clone().SetMsg("test message %d", i)
		}
	})

	b.Run("New", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New(404001, fmt.Sprintf("test message %d", i))
		}
	})

}
