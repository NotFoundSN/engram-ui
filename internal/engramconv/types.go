// Package engramconv exposes constants and slices derived from the engram
// type taxonomy. The source of truth for the canonical type list is
// skills/engram-conventions/types.md — update this file in lockstep with
// that document. Order: alphabetical (matches UI display order per D2-1).
package engramconv

// CanonicalTypes is the 14-type taxonomy defined by
// skills/engram-conventions/types.md. The slice is alphabetically sorted
// so it can be rendered directly as a <select> option list without an
// extra sort step.
var CanonicalTypes = []string{
	"architecture",
	"bugfix",
	"config",
	"decision",
	"design",
	"discovery",
	"exploration",
	"pattern",
	"plan",
	"preference",
	"proposal",
	"report",
	"spec",
	"tasks",
}
