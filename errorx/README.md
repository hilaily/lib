# ErrorX - Go 结构化错误处理库

ErrorX 是一个专为 Go 应用设计的结构化错误处理库，提供了统一的错误管理机制，支持业务错误和系统错误的分类处理，并提供了完善的适配器模式用于不同协议的错误转换。

## 设计原理

### 核心设计思想

ErrorX 基于以下核心设计原则：

1. **错误分类管理**：将错误分为业务错误和系统错误两大类
   - **业务错误**：业务逻辑上允许的失败情况（如用户不存在、权限不足等）
   - **系统错误**：系统异常状态（如数据库连接失败、网络超时等）

2. **并发安全**：通过 Clone 模式确保错误对象的并发安全性
3. **适配器模式**：支持不同协议的错误格式转换（如 Connect RPC、HTTP 等）
4. **链式调用**：提供流畅的 API 用于错误信息的构建
5. **标准兼容**：完全兼容 Go 标准库的 `error` 接口和 `errors.Is`/`errors.As` 机制

### 错误结构

每个错误包含以下字段：

- **code**: 错误码，0 表示正常，非 0 表示异常
- **message**: 错误消息，业务错误展示给用户，系统错误统一展示通用消息
- **err**: 底层错误，用于错误链和调试信息
- **params**: 参数信息，用于错误消息的参数化
- **extra**: 扩展信息，如堆栈跟踪、调试信息等

### 预定义错误类型

ErrorX 提供了标准的 HTTP 状态码对应的错误类型：

```go
// 业务错误 (4xx)
BadRequest       (400) - 请求参数错误
InvalidArgument  (400) - 参数无效
Auth             (401) - 认证失败
PermissionDenied (403) - 权限不足
NotFound         (404) - 资源不存在
AlreadyExists    (409) - 资源已存在
Canceled         (499) - 操作取消

// 系统错误 (5xx)
Internal         (500) - 内部服务器错误
Unknown          (500) - 未知错误
Timeout          (504) - 操作超时
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "your-project/errorx"
)

func main() {
    // 业务错误 - 使用预定义错误类型
    err := errorx.NotFound.SetMsg("用户不存在，ID: %d", 12345)
    fmt.Println(err) // code: 404, message: 用户不存在，ID: 12345

    // 系统错误 - 包装底层错误
    dbErr := errors.New("connection refused")
    err = errorx.Internal.SetErr(dbErr)
    fmt.Println(err) // code: 500, message: internal_server_error, err: connection refused

    // 创建自定义错误
    err = errorx.New(40001, "自定义业务错误")
    fmt.Println(err) // code: 40001, message: 自定义业务错误
}
```

## 详细功能

### 1. 业务错误处理

业务错误是应用逻辑中预期的错误情况，这些错误的消息会直接展示给用户：

```go
// 用户认证失败
err := errorx.Auth.SetMsg("用户名或密码错误")

// 权限不足
err := errorx.PermissionDenied.SetMsg("您没有访问此资源的权限")

// 资源不存在
err := errorx.NotFound.SetMsg("订单 %s 不存在", orderID)

// 资源冲突
err := errorx.AlreadyExists.SetMsg("邮箱 %s 已被注册", email)

// 参数验证错误
err := errorx.InvalidArgument.SetMsg("年龄必须在 18-100 之间")
```

### 2. 系统错误处理

系统错误是非预期的技术性错误，通常包含调试信息但不直接暴露给用户：

```go
// 使用 Errorf 快速创建系统错误
err := errorx.Errorf("数据库查询失败: %v", dbErr)

// 包装底层错误
err := errorx.Internal.SetErr(dbErr).SetMsg("用户服务暂时不可用")

// 超时错误
err := errorx.Timeout.SetErr(timeoutErr).SetMsg("请求处理超时")
```

### 3. 错误信息增强

```go
// 添加堆栈跟踪
err := errorx.Internal.SetErr(dbErr).WithStack()

// 添加扩展信息
err := errorx.NotFound.SetMsg("用户不存在").
    WithExtra("user_id", userID).
    WithExtra("request_id", requestID)

// 添加参数信息
params := map[string]any{
    "user_id": userID,
    "action":  "delete_user",
}
err := errorx.PermissionDenied.WithParams(params).SetMsg("操作权限不足")
```

### 4. 错误类型检查

ErrorX 完全兼容 Go 标准库的错误处理机制：

```go
err := someFunction()

// 使用 errors.Is 检查错误类型
if errors.Is(err, errorx.NotFound) {
    // 处理资源不存在的情况
    return handleNotFound()
}

if errors.Is(err, errorx.Auth) {
    // 处理认证失败
    return handleAuthError()
}

// 使用 errors.As 获取具体错误
var errx *errorx.Err
if errors.As(err, &errx) {
    fmt.Printf("错误码: %d, 消息: %s\n", errx.GetCode(), errx.GetMsg())
}
```

### 5. 适配器模式

ErrorX 支持通过适配器模式转换为不同协议的错误格式：

```go
// 设置 Connect RPC 适配器（默认）
errorx.SetAdapter(&errorx.ConnectRpcAdapter{})

// 创建错误时会自动适配为 Connect 错误格式
err := errorx.NotFound.SetMsg("用户不存在")
// err 现在包含了适当的 Connect RPC 错误信息

// 可以扩展其他适配器，如 HTTP 适配器
// errorx.SetAdapter(&errorx.HTTPAdapter{})
```

### 6. JSON 序列化

ErrorX 支持 JSON 序列化，便于日志记录和 API 响应：

```go
err := errorx.InvalidArgument.SetMsg("参数验证失败").
    WithExtra("field", "email").
    WithExtra("value", "invalid-email")

jsonData, _ := json.Marshal(err)
fmt.Println(string(jsonData))
// 输出: {"code":400,"message":"参数验证失败","err":null,"params":null,"extra":{"field":"email","value":"invalid-email"}}
```

## 最佳实践

### 1. 错误分类原则

```go
// ✅ 正确：业务逻辑错误，用户可以理解和处理
func GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, errorx.InvalidArgument.SetMsg("用户ID必须大于0")
    }

    user, err := db.GetUser(id)
    if err == sql.ErrNoRows {
        return nil, errorx.NotFound.SetMsg("用户不存在")
    }
    if err != nil {
        // 系统错误，包装底层错误用于调试
        return nil, errorx.Internal.SetErr(err).SetMsg("获取用户信息失败")
    }

    return user, nil
}
```

### 3. 错误链和调试

```go
func serviceMethod() error {
    if err := databaseOperation(); err != nil {
        // 保留错误链，便于调试
        return errorx.Internal.SetErr(err).
            WithStack().
            WithExtra("operation", "user_query").
            SetMsg("用户查询服务暂时不可用")
    }
    return nil
}
```

### 4. API 错误响应

```go
func handleError(w http.ResponseWriter, err error) {
    var errx *errorx.Err
    if errors.As(err, &errx) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(int(errx.GetCode()))

        response := map[string]any{
            "code":    errx.GetCode(),
            "message": errx.GetMsg(),
        }

        // 开发环境可以包含更多调试信息
        if isDevelopment() {
            if stack := errx.StackTrace(); stack != "" {
                response["stack_trace"] = stack
            }
        }

        json.NewEncoder(w).Encode(response)
        return
    }

    // 处理其他类型的错误
    http.Error(w, "内部服务器错误", 500)
}
```

## 扩展功能

### 自定义适配器

你可以实现自己的适配器来支持特定的协议或框架：

```go
type CustomAdapter struct{}

func (a *CustomAdapter) Adapt(e *errorx.Err) *errorx.Err {
    // 实现自定义的错误转换逻辑
    // 例如：转换为特定框架的错误格式
    return e
}

// 使用自定义适配器
errorx.SetAdapter(&CustomAdapter{})
```

### 扩展错误类型

```go
// 定义业务特定的错误码
const (
    CodeUserBlocked   uint32 = 40301
    CodeQuotaExceeded uint32 = 42901
)

// 创建业务特定的错误类型
var (
    UserBlocked   = errorx.New(CodeUserBlocked, "user_blocked")
    QuotaExceeded = errorx.New(CodeQuotaExceeded, "quota_exceeded")
)

// 使用
err := UserBlocked.SetMsg("用户账号已被冻结，请联系管理员")
```
