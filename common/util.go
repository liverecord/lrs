package common

import (
	"os"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func Env(key, def string) string {
	val, set := os.LookupEnv(key)
	if set {
		return val
	}

	return def
}

func BoolEnv(key string, def bool) bool {
	val, set := os.LookupEnv(key)
	if set {
		return strings.ToLower(val) == "true"
	}

	return def
}

func FilterHtml(html string) string {
	p := bluemonday.NewPolicy()
	p.AllowImages()
	p.AllowLists()
	p.AllowAttrs("cite").OnElements("blockquote")
	p.AllowElements("br", "hr", "p", "span", "code", "kbd", "sub", "sup", "b", "i", "u", "strong", "em")

	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src").OnElements("video")

	html = p.Sanitize(html)
	return html
}
