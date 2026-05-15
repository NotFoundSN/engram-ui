package server

import (
	"strings"
	"unicode/utf8"
)

const maxFromLen = 2048

// validateFrom returns s if it is a safe same-origin path suitable for use
// as a back-link href, or "" if it must be rejected. A return of "" instructs
// the caller to fall back to "/".
//
// The input is expected to be already URL-decoded (chi/net/http decode query
// params by default). validateFrom does NOT decode further — this closes the
// double-encoding bypass surface.
//
// Rules (evaluated in order):
//  1. Empty string                    -> reject ""
//  2. Length > 2048 runes             -> reject ""
//  3. Does not start with "/"         -> reject "" (catches "http://...", "javascript:", etc.)
//  4. Starts with "//"                -> reject "" (catches protocol-relative URLs)
//  5. Contains the substring ".."     -> reject "" (catches path traversal)
//
// If all rules pass, return s unchanged.
func validateFrom(s string) string {
	if s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) > maxFromLen {
		return ""
	}
	if s[0] != '/' {
		return ""
	}
	if len(s) >= 2 && s[1] == '/' {
		return ""
	}
	if strings.Contains(s, "..") {
		return ""
	}
	return s
}
