package server

import (
	"net/url"
	"strings"
)

// buildSourceURL constructs the unescaped URL that represents the user's
// current project view. It is used as the value embedded in the ?from=
// query parameter on observation row links.
//
// Params are emitted in canonical order: type → q → sort. Only non-empty
// params are included. sort is omitted when sortExplicit is false (i.e., the
// handler defaulted to date_desc without the user explicitly requesting it).
//
// The returned string is NOT percent-encoded — the caller applies
// url.QueryEscape when embedding it into a ?from= attribute.
func buildSourceURL(project, typ, q, sort string, sortExplicit bool) string {
	base := "/p/" + project
	parts := make([]string, 0, 3)
	if typ != "" {
		parts = append(parts, "type="+url.QueryEscape(typ))
	}
	if q != "" {
		parts = append(parts, "q="+url.QueryEscape(q))
	}
	if sortExplicit && sort != "" {
		parts = append(parts, "sort="+url.QueryEscape(sort))
	}
	if len(parts) == 0 {
		return base
	}
	return base + "?" + strings.Join(parts, "&")
}
