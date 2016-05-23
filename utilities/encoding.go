package utilities

import (
	"errors"
	"unicode/utf8"
)

// The string encoding options. Currently only support UTF-8. Don't really
// see the merit in supporting anything else at the moment.
const (
	UTF8 = iota
)

// InvalidEncodingTypeErr is returned when the encoding type is not one that
// is supported.
var InvalidEncodingTypeErr = errors.New("sdbot/utilities: invalid string encoding type")

// EncodeIncoming returns a strings that is encoded in the provided encoding
// type. If the encoding type is invalid then we return the original string,
// but also return an error.
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
