package i18n

import "golang.org/x/text/language"

type Option func(*I18n)

func WithHeaderKey(key string) Option {
	return func(i *I18n) {
		i.headerKey = key
	}
}

func WithDefaultLanguage(language language.Tag) Option {
	return func(i *I18n) {
		i.defaultLanguage = language
	}
}

func WithLocaleDir(dir string) Option {
	return func(i *I18n) {
		i.localeDir = dir
	}
}
