package main

import (
	"strings"
	"testing"
)

// UTF-8 byte escapes keep the source unambiguous for invisible characters.
const (
	zwsp  = "\xe2\x80\x8b" // U+200B zero-width space
	zwj   = "\xe2\x80\x8d" // U+200D zero-width joiner
	bom   = "\xef\xbb\xbf" // U+FEFF
	rlo   = "\xe2\x80\xae" // U+202E right-to-left override
	lri   = "\xe2\x81\xa6" // U+2066 left-to-right isolate
	ls    = "\xe2\x80\xa8" // U+2028 line separator
	ps    = "\xe2\x80\xa9" // U+2029 paragraph separator
	combC = "\xcc\x82"     // U+0302 combining circumflex
	eHat  = "\xc3\xaa"     // U+00EA precomposed e with circumflex
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		max     int
		want    string
		wantErr bool
	}{
		{"trims", "  kho ", 30, "kho", false},
		{"empty", "", 30, "", true},
		{"only spaces", "   \t ", 30, "", true},
		{"exactly 30 runes", strings.Repeat("a", 30), 30, strings.Repeat("a", 30), false},
		{"31 runes", strings.Repeat("a", 31), 30, "", true},
		{"vietnamese ok", "Lưu Thành Danh", 30, "Lưu Thành Danh", false},
		{"counts runes not bytes", strings.Repeat("ư", 30), 30, strings.Repeat("ư", 30), false},
		{"31 multibyte runes", strings.Repeat("ư", 31), 30, "", true},
		{"score name max 20", strings.Repeat("a", 21), 20, "", true},
		{"strips control chars", "k\x00h\x07o\n\t", 30, "kho", false},
		{"only control chars", "\x00\x01", 30, "", true},
		{"tab/newline become single space", "a\tb \n  c", 30, "a b c", false},
		{"zero-width only is empty", zwsp + zwj + bom, 30, "", true},
		{"strips bidi override and zero-width", "k" + rlo + "h" + zwsp + "o" + lri, 30, "kho", false},
		{"NFC before counting", strings.Repeat("e"+combC, 30), 30, strings.Repeat(eHat, 30), false},
		{"line separator becomes space", "a" + ls + "b", 30, "a b", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateName(tc.in, tc.max)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"empty", "", "", true},
		{"only whitespace", " \n ", "", true},
		{"trims", "  HPBD  ", "HPBD", false},
		{"280 runes", strings.Repeat("ê", 280), strings.Repeat("ê", 280), false},
		{"281 runes", strings.Repeat("ê", 281), "", true},
		{"collapses 4 newlines", "a\n\n\n\nb", "a\n\nb", false},
		{"keeps double newline", "a\n\nb", "a\n\nb", false},
		{"normalizes CRLF", "a\r\n\r\n\r\nb", "a\n\nb", false},
		{"strips control but keeps newline", "a\x00\x07b\nc", "ab\nc", false},
		{"whitespace-only lines collapse", "a\n \n \n\nb", "a\n\nb", false},
		{"zero-width lines collapse", "a\n" + zwsp + "\n" + zwsp + "\nb", "a\n\nb", false},
		{"U+2028 and U+2029 are newlines", "a" + ls + ps + ls + ps + "b", "a\n\nb", false},
		{"tab becomes space", "a\tb", "a b", false},
		{"trailing spaces per line trimmed", "a   \nb", "a\nb", false},
		{"NFC before counting", strings.Repeat("e"+combC, 280), strings.Repeat(eHat, 280), false},
		{"only zero-width is empty", zwsp + zwsp, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateMessage(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateMessageCapsNewlines(t *testing.T) {
	got, err := validateMessage(strings.TrimSuffix(strings.Repeat("x\n", 12), "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(got, "\n"); n != 8 {
		t.Fatalf("newlines = %d, want 8 in %q", n, got)
	}
	if len(strings.Fields(got)) != 12 {
		t.Fatalf("content lost: %q", got)
	}
}

func TestValidateScore(t *testing.T) {
	for _, tc := range []struct {
		n       int
		wantErr bool
	}{{-1, true}, {0, false}, {500, false}, {1000, false}, {1001, true}} {
		if err := validateScore(tc.n); (err != nil) != tc.wantErr {
			t.Errorf("validateScore(%d) err = %v, wantErr %v", tc.n, err, tc.wantErr)
		}
	}
}
