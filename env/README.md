# Env

一个简单的 .env 文件加载工具，用于管理环境变量。

## 功能特点

- 从当前目录开始递归向上查找 `.env` 文件
- 支持多环境配置（如 .env.development、.env.production）
- 提供了安全和必需两种加载模式
- 支持环境变量覆盖（系统环境变量优先级高于 .env 文件）

## 使用方法

### 基本用法

```
go
import "your-project/env"
func main() {
// 加载 .env 文件
env.Load()
// 获取环境变量
dbHost := env.Get("DB_HOST")
// 获取环境变量，如果不存在则返回默认值
port := env.GetDefault("PORT", "3000")
}
```

### 环境文件示例

```
env
.env 文件示例
DB_HOST=localhost
DB_PORT=5432
API_KEY=your-api-key
```

### 多环境支持

项目支持根据不同环境加载对应的环境文件：

- `.env` - 默认环境文件
- `.env.development` - 开发环境
- `.env.test` - 测试环境
- `.env.production` - 生产环境

## 注意事项

- 环境文件应该被添加到 .gitignore 中
- 系统环境变量的优先级高于 .env 文件中的配置
- 建议在项目根目录下放置环境文件
