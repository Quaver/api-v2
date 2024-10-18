package stringutil

import "github.com/microcosm-cc/bluemonday"

func SanitizeHTML(html string) string {
	p := bluemonday.NewPolicy()

	p.AllowElements("p", "b", "strong", "i", "em", "u", "ul", "ol", "li", "s", "a", "img", "table", "tr", "td", "center")
	p.AllowAttrs("style").OnElements("p")
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt", "width", "height", "style").OnElements("img")

	return p.Sanitize(html)
}
