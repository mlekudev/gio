// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios

package app

import (
	"unicode/utf8"

	"github.com/mlekudev/gio/io/key"
)

// areSnippetsConsistent reports whether the content of the old snippet is
// consistent with the content of the new.
func areSnippetsConsistent(old, new key.Snippet) bool {
	// Compute the overlapping range.
	r := old.Range
	r.Start = max(r.Start, new.Start)
	r.End = max(r.End, r.Start)
	r.End = min(r.End, new.End)
	return snippetSubstring(old, r) == snippetSubstring(new, r)
}

func snippetSubstring(s key.Snippet, r key.Range) string {
	for r.Start > s.Start && r.Start < s.End {
		_, n := utf8.DecodeRuneInString(s.Text)
		s.Text = s.Text[n:]
		s.Start++
	}
	for r.End < s.End && r.End > s.Start {
		_, n := utf8.DecodeLastRuneInString(s.Text)
		s.Text = s.Text[:len(s.Text)-n]
		s.End--
	}
	return s.Text
}
