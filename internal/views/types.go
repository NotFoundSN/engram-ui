// Package views — TypeMeta table.
//
// Observation `type` strings are paired with a display label and a visual hue
// (HSL triple). The label is rendered in templ; the hue is for documentation
// and applied via CSS through `[data-type="<name>"]` selectors in app.css
// (see "Per-type hue assignments" block). Templ emits `data-type` on the
// element that should be tinted, CSS resolves the `--type-hue` custom prop,
// and consumers read it via `hsl(var(--type-hue))`.
//
// Adding a new type means adding one entry here AND one selector in app.css.
// Keep them in sync — there is no automated link between the two.
package views

// TypeMeta carries the per-type display data.
type TypeMeta struct {
	Hue   string // HSL triple "<H S% L%>" — drives --type-hue inline
	Label string // user-facing label (may differ from the raw type)
}

// TypeMetas covers engram's canonical observation types. Unknowns fall through
// to a neutral grey via TypeMetaOf — we never panic on an unfamiliar string.
var TypeMetas = map[string]TypeMeta{
	"bugfix":          {Hue: "0 78% 60%", Label: "bugfix"},
	"decision":        {Hue: "258 84% 71%", Label: "decision"},
	"architecture":    {Hue: "217 88% 66%", Label: "architecture"},
	"discovery":       {Hue: "38 95% 58%", Label: "discovery"},
	"pattern":         {Hue: "189 84% 51%", Label: "pattern"},
	"config":          {Hue: "215 16% 60%", Label: "config"},
	"preference":      {Hue: "329 84% 68%", Label: "preference"},
	"session_summary": {Hue: "160 70% 48%", Label: "session"},
	"note":            {Hue: "215 14% 55%", Label: "note"},
}

// typeFallback is returned for any name not in TypeMetas. Neutral grey so the
// UI keeps rendering without a colored claim about an unknown type.
var typeFallback = TypeMeta{Hue: "215 14% 55%", Label: "other"}

// TypeMetaOf returns the entry for name, or the neutral fallback. The fallback
// also reuses its hue/label for the empty string, so a missing type still
// renders something coherent.
func TypeMetaOf(name string) TypeMeta {
	if m, ok := TypeMetas[name]; ok {
		return m
	}
	return typeFallback
}
