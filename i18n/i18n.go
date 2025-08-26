package i18n

import (
	"context"
	"net/http"
	"os"
	"path"
	"strings"

	"connectrpc.com/connect"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/hilaily/lib/errorx"
)

type I18n struct {
	bundle          *i18n.Bundle
	localeDir       string
	defaultLanguage language.Tag
	headerKey       string
}

// 初始化 i18n bundle
func New(opts ...Option) (*I18n, error) {
	i := &I18n{
		localeDir:       "locales",
		defaultLanguage: language.Chinese,
		headerKey:       "X-Language",
	}
	for _, opt := range opts {
		opt(i)
	}

	i.bundle = i18n.NewBundle(i.defaultLanguage)
	i.bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// 加载所有翻译文件
	entries, err := os.ReadDir(i.localeDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			filePath := path.Join(i.localeDir, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}

			_, err = i.bundle.ParseMessageFileBytes(data, filePath)
			if err != nil {
				return nil, err
			}
		}
	}
	return i, nil
}

// I18nMiddleware 从请求中检测语言并设置到上下文中
func (i *I18n) GinMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i.detectLanguage(r)
		ctx := r.Context()

		// 将语言信息保存到上下文中
		ctx = context.WithValue(ctx, "language", lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (i *I18n) ConnectRpcAuthInterceptor() connect.UnaryInterceptorFunc {
	adapter := new(errorx.ConnectRPCAdapter)
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			lang := req.Header().Get("X-Language")
			ctx = context.WithValue(ctx, "language", lang)
			res, err := next(ctx, req)
			if err != nil {
				e, ok := err.(*errorx.Err)
				if ok {
					e = e.SetMsg("%s", i.T(lang, e.GetMsg(), e.GetParams()))
					return res, adapter.ToConnectRpcError(e)
				}
				return res, err
			}
			return res, nil
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

// 语言检测策略：
// 1. 首先检查查询参数 ?lang=xx
// 2. 然后检查 Cookie
// 3. 最后检查 Accept-Language 头
func (i *I18n) detectLanguage(r *http.Request) string {
	// 1. 检查查询参数
	queryLang := r.URL.Query().Get("lang")
	if queryLang != "" {
		return queryLang
	}

	// 3. 检查 Accept-Language 头
	acceptLang := r.Header.Get(i.headerKey)
	if acceptLang != "" {
		// 提取首选语言
		langs := strings.Split(acceptLang, ",")
		if len(langs) > 0 {
			// 提取语言代码 (en-US -> en)
			return strings.Split(langs[0], "-")[0]
		}
	}

	// 3. 检查 Cookie
	langCookie, err := r.Cookie("language")
	if err == nil && langCookie.Value != "" {
		return langCookie.Value
	}

	// 默认返回英语
	return "en"
}

// 获取针对特定语言的本地化器
func (i *I18n) NewLocalizer(lang string) *i18n.Localizer {
	// 如果传入空字符串，使用默认语言
	if lang == "" {
		return i18n.NewLocalizer(i.bundle, i.defaultLanguage.String())
	}

	// 创建包含回退语言的本地化器
	return i18n.NewLocalizer(i.bundle, lang)
}

func (i *I18n) TCtx(ctx context.Context, messageID string, params ...map[string]interface{}) string {
	lang, _ := ctx.Value("x-language").(string)
	return i.T(lang, messageID, params...)
}

// TWithData 用于包含模板数据的翻译
func (i *I18n) T(lang, messageID string, params ...map[string]interface{}) string {
	localizer := i.NewLocalizer(lang)
	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	// 添加变量
	if len(params) > 0 {
		config.TemplateData = params[0]
		// 判断复数
		if count, ok := params[0]["Count"]; ok {
			config.PluralCount = count
		}
	}

	msg, err := localizer.Localize(config)
	if err != nil {
		return messageID
	}

	return msg
}
