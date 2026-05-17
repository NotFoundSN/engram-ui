// Package views — theme context helpers.
//
// Theme is injected into the request context by the server middleware so that
// every templ component (most importantly Layout) can read it directly via
// `ctx` without each handler having to thread a `Theme string` through every
// view-specific Data struct.
package views

import "context"

// themeValid lists the values accepted as a theme. Anything else (or missing)
// falls back to "dark", which matches the :root defaults in app.css.
var themeValid = map[string]struct{}{
	"dark":  {},
	"light": {},
}

type themeCtxKey struct{}

// ContextWithTheme returns a derived context carrying the resolved theme.
// Unknown values are normalized to "dark" so consumers never have to validate.
func ContextWithTheme(ctx context.Context, theme string) context.Context {
	if _, ok := themeValid[theme]; !ok {
		theme = "dark"
	}
	return context.WithValue(ctx, themeCtxKey{}, theme)
}

// ThemeFromContext returns the theme stored in ctx, or "dark" when absent.
// Safe to call on any context — never panics, always returns a usable value.
func ThemeFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(themeCtxKey{}).(string); ok {
		return v
	}
	return "dark"
}
