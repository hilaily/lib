# I18n 国际化库

一个基于 `go-i18n` 的 Go 语言国际化库，支持多语言翻译、语言检测和中间件集成。

## 功能特性

- 🌍 支持多语言翻译
- 🔍 智能语言检测（查询参数、Header、Cookie）
- 🚀 Gin 和 Connect RPC 中间件支持
- 📁 YAML 格式翻译文件支持
- 🔧 灵活的配置选项
- 🔄 与 errorx 错误处理库集成

## 安装

## 快速开始

### 1. 创建翻译文件

在项目根目录创建 `locale` 目录，并添加翻译文件：

**locale/en.yaml**

```yaml
hello: "Hello"
welcome: "Welcome {{.Name}}"
user_count:
  one: "{{.Count}} user"
  other: "{{.Count}} users"
```

**locale/zh.yaml**

```yaml
hello: "你好"
welcome: "欢迎 {{.Name}}"
user_count:
  one: "{{.Count}} 个用户"
  other: "{{.Count}} 个用户"
```

### 2. 初始化 I18n 实例

```go
package main

import (
    "github.com/hilaily/lib/i18n"
    "golang.org/x/text/language"
)

func main() {
    // 使用默认配置
    i18nInstance, err := i18n.New()
    if err != nil {
        panic(err)
    }

    // 或者使用自定义配置
    i18nInstance, err = i18n.New(
        i18n.WithLocaleDir("translations"),           // 自定义翻译文件目录
        i18n.WithDefaultLanguage(language.English),   // 设置默认语言
        i18n.WithHeaderKey("Accept-Language"),        // 自定义语言检测 Header
    )
    if err != nil {
        panic(err)
    }
}
```

### 3. 基本翻译

```go
// 简单翻译
msg := i18nInstance.T("en", "hello")
// 输出: "Hello"

msg = i18nInstance.T("zh", "hello")
// 输出: "你好"

// 带参数的翻译
msg = i18nInstance.T("en", "welcome", map[string]interface{}{
    "Name": "John",
})
// 输出: "Welcome John"

// 复数翻译
msg = i18nInstance.T("en", "user_count", map[string]interface{}{
    "Count": 5,
})
// 输出: "5 users"
```

### 4. 上下文翻译

```go
import "context"

// 从上下文获取语言进行翻译
ctx := context.WithValue(context.Background(), "x-language", "zh")
msg := i18nInstance.TCtx(ctx, "hello")
// 输出: "你好"
```

## 中间件集成

### Gin 中间件

```go
import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default()

    // 添加 i18n 中间件
    r.Use(func(c *gin.Context) {
        i18nInstance.GinMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Request = r
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    })

    r.GET("/hello", func(c *gin.Context) {
        // 从上下文获取语言
        lang, _ := c.Request.Context().Value("language").(string)
        msg := i18nInstance.T(lang, "hello")
        c.JSON(200, gin.H{"message": msg})
    })

    r.Run(":8080")
}
```

### Connect RPC 拦截器

```go
import "connectrpc.com/connect"

// 添加 i18n 拦截器
interceptor := i18nInstance.ConnectRpcAuthInterceptor()

// 在 Connect RPC 服务中使用
// 这个拦截器会自动处理错误消息的国际化
```

## 语言检测策略

库会按以下优先级检测客户端语言：

1. **查询参数**: `?lang=zh`
2. **HTTP Header**: `X-Language: zh` (可通过 `WithHeaderKey` 自定义)
3. **Cookie**: `language=zh`
4. **默认语言**: 如果都没有，使用默认语言（默认为中文）

### 示例请求

```bash
# 通过查询参数指定语言
curl "http://localhost:8080/api?lang=en"

# 通过 Header 指定语言
curl -H "X-Language: zh" "http://localhost:8080/api"

# 通过 Cookie 指定语言
curl -H "Cookie: language=en" "http://localhost:8080/api"
```

## 配置选项

### WithLocaleDir(dir string)

设置翻译文件目录，默认为 `"locale"`。

```go
i18nInstance, err := i18n.New(
    i18n.WithLocaleDir("translations"),
)
```

### WithDefaultLanguage(lang language.Tag)

设置默认语言，默认为 `language.Chinese`。

```go
import "golang.org/x/text/language"

i18nInstance, err := i18n.New(
    i18n.WithDefaultLanguage(language.English),
)
```

### WithHeaderKey(key string)

设置用于语言检测的 HTTP Header 键名，默认为 `"X-Language"`。

```go
i18nInstance, err := i18n.New(
    i18n.WithHeaderKey("Accept-Language"),
)
```

## 翻译文件格式

翻译文件使用 YAML 格式，支持：

### 简单翻译

```yaml
key: "翻译内容"
```

### 参数化翻译

```yaml
greeting: "你好，{{.Name}}！"
```

### 复数翻译

```yaml
item_count:
  one: "{{.Count}} 个项目"
  other: "{{.Count}} 个项目"
```

### 嵌套翻译

```yaml
user:
  profile:
    name: "姓名"
    email: "邮箱"
```

## 错误处理

库与 `errorx` 错误处理库集成，在 Connect RPC 拦截器中会自动翻译错误消息：

```go
// 在业务代码中
err := errorx.New("user_not_found").SetParams(map[string]interface{}{
    "ID": userID,
})

// 拦截器会自动根据客户端语言翻译错误消息
```

## 最佳实践

1. **翻译文件命名**: 使用语言代码命名文件，如 `en.yaml`、`zh.yaml`
2. **键名规范**: 使用下划线分隔的小写字母，如 `user_not_found`
3. **参数命名**: 使用 PascalCase，如 `{{.UserName}}`
4. **复数处理**: 为需要复数的消息提供 `one` 和 `other` 形式
5. **回退机制**: 确保为所有消息提供默认语言的翻译

## 注意事项

- 翻译文件必须放在指定的 locale 目录中
- 文件名必须以 `.yaml` 结尾
- 如果翻译不存在，会返回原始的 messageID
- 在 Connect RPC 中，语言信息通过 `X-Language` header 传递

## 示例项目

完整的使用示例可以参考项目中的测试文件和其他模块的集成方式。
