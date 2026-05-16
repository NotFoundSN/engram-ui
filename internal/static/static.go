// Package static embeds the compiled static assets (CSS, fonts) for engram-ui.
package static

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var embedded embed.FS

// FS returns the embedded filesystem rooted at the inner "static" directory.
// Using an accessor (not a package-level exported var) keeps the embed.FS
// private; future callers can swap to an os.DirFS for dev hot-reload by
// changing only this function.
func FS() fs.FS {
	sub, err := fs.Sub(embedded, "static")
	if err != nil {
		panic("static: failed to sub embedded FS: " + err.Error())
	}
	return sub
}
