package lrs

import "github.com/microcosm-cc/bluemonday"

// S2BA converts string to bytes
func S2BA(value string) []byte {
	return []byte(value)
}

// StripTags removes all tags from string
func StripTags(html string) string {
	p := bluemonday.StripTagsPolicy()
	return p.Sanitize(html)
}