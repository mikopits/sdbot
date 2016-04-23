package sdbot

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	StartPattern = iota
	EndPattern
	RegularPattern
)

var (
	ErrInvalidAnchor = errors.New("pattern: invalid anchor type")
	ErrInvalidType   = errors.New("pattern: invalid conversion type")
)

type Pattern struct {
	Prefix  *regexp.Regexp
	Pattern *regexp.Regexp
	Suffix  *regexp.Regexp
}

// Convert one of [string, *regexp.Regexp, nil] to a regexp.Regexp
func interfaceToRegexp(i interface{}, anchor int) (*regexp.Regexp, error) {
	switch fmt.Sprintf("%T", i) {
	case "string":
		escaped := regexp.QuoteMeta(i.(string))
		var err error
		var reg *regexp.Regexp
		switch anchor {
		case StartPattern:
			reg, err = regexp.Compile("^" + escaped)
			if err != nil {
				Error(&Log, err)
			}
			return reg, nil
		case EndPattern:
			reg, err = regexp.Compile(escaped + "$")
			if err != nil {
				Error(&Log, err)
			}
			return reg, nil
		case RegularPattern:
			reg, err = regexp.Compile(escaped)
			if err != nil {
				Error(&Log, err)
			}
			return reg, nil
		default:
			return nil, ErrInvalidAnchor
		}
	case "<nil>":
		fallthrough
	case "*regexp.Regexp":
		return i.(*regexp.Regexp), nil
	default:
		return nil, ErrInvalidType
	}
}

func (p *Pattern) ToRegexp(m *Message) *regexp.Regexp {
	reg, err := regexp.Compile(fmt.Sprintf("%s%s%s", p.Prefix.String(), p.Pattern.String(), p.Suffix.String()))
	if err != nil {
		Error(&Log, err)
	}

	return reg
}
