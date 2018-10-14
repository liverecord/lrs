package common

import (
	"os"
	"strconv"

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
	if !set {
		return def
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}

	return boolVal
}

func IntEnv(key string, def int) int {
	val, set := os.LookupEnv(key)
	if !set {
		return def
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return def
	}

	return intVal
}

func SanitizeHtml(html string) string {
	p := bluemonday.NewPolicy()
	p.AllowImages()
	p.AllowLists()
	p.AllowTables()
	p.AllowAttrs("cite").OnElements("blockquote")
	p.AllowElements("br", "hr", "p", "pre", "span", "code", "kbd", "sub", "sup", "b", "i", "u", "strong", "em")

	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src").OnElements("video")

	html = p.Sanitize(html)
	return html
}
