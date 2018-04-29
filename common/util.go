package common

import (
	"os"
	"strconv"

	"github.com/microcosm-cc/bluemonday"
)

func S2BA(value string) []byte {
	return []byte(value)
}

func Env(key, def string) string {
	val, set := os.LookupEnv(key)
	if set {
		return val
	}

	return def
}

func BoolEnv(key string, def bool) bool {
	val, set := os.LookupEnv(key)
	if !set {
		return def
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}

	return boolVal
}

func StripTags(html string) string {
	p := bluemonday.StripTagsPolicy()
	return p.Sanitize(html)
}

func SanitizeHtml(html string) string {
	p := bluemonday.NewPolicy()
	p.AllowImages()
	p.AllowLists()
	p.AllowTables()
	p.AllowAttrs("cite").OnElements("blockquote")
	p.AllowElements("br", "hr", "p", "span", "code", "kbd", "sub", "sup", "b", "i", "u", "strong", "em")

	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src").OnElements("video")

	html = p.Sanitize(html)
	return html
}
