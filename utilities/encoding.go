package utilities

import (
	"errors"
	"unicode/utf8"
)

// The string encoding options.
const (
	UTF8 = iota
)

var InvalidEncodingTypeErr = errors.New("sdbot/utilities: invalid string encoding type")

func EncodeIncoming(s string, encoding int) (string, error) {
	switch encoding {
	case UTF8:
		if !utf8.ValidString(s) {
			v := make([]rune, 0, len(s))
			for i, r := range s {
				if r == utf8.RuneError {
					_, size := utf8.DecodeRuneInString(s[i:])
					if size == 1 {
						continue
					}
				}
				v = append(v, r)
			}
			s = string(v)
		}
		return s, nil
	default:
		return s, InvalidEncodingTypeErr
	}
}
