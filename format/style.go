package format

import (
	"errors"
)

type Style struct {
	name string
}

var (
	StyleNone    = Style{""}        // do not check formatting
	StyleDefault = Style{"default"} // enable default format checks
	StyleWrap    = Style{"wrap"}    // enable with support for line breaks

	ErrUnsupportedFormat = errors.New("unsupported style")
)

// String returns the format style name.
func (s Style) String() string {
	return s.name
}

// FromString returns the style or an error if unmatched.
func FromString(s string) (Style, error) {
	switch s {
	case StyleNone.name, "disable":
		return StyleNone, nil
	case StyleDefault.name, "enable":
		return StyleDefault, nil
	case StyleWrap.name:
		return StyleWrap, nil
	}

	return StyleNone, ErrUnsupportedFormat
}
