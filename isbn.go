package isbn

import (
	"bytes"
	"fmt"
	"strings"
)

// ISBN represents the number in an intermediate form.
// This allows us to efficiently convert/check/stringify
type ISBN struct {
	is13     bool
	prefix   [3]byte
	digits   [9]byte
	checksum byte
}

var allowedISBN13Prefixes = [][]byte{{9, 7, 8}, {9, 7, 9}}

const urnPrefix = `urn:isbn:`

// convert the rune to it's isbn digit value, returning
// -1 for invalid characters, which are stripped.
func runeToISBNDigit(r rune) rune {
	switch true {
	case r >= '0' && r <= '9':
		return r - 48
	case r == 'x' || r == 'X':
		return 10
	default:
		return -1
	}
}
func isbnDigitToByte(r byte) byte {
	switch true {
	case r >= 0 && r <= 9:
		return r + 48
	case r == 10:
		return 'X'
	default:
		panic("Invalid byte in ISBN data")
	}
}

// Validate returns whether the string is an ISBN, nothing else.
func Validate(s string) bool {
	_, err := Parse(s)
	return err == nil
}

// Parse turns a string into an ISBN, or throws an error.
// The string must be contain only digits and hyphens,
// expect for the optional prefix `urn:isbn:`
func Parse(s string) (*ISBN, error) {
	if strings.HasPrefix(s, urnPrefix) {
		s = s[len(urnPrefix):]
	}
	// now strip unwanted characters.
	// Note that the string itseflf may contain hyphens or spaces
	// but should not contain more than 4. So we can check length
	// here.
	if len(s) > 13+4 {
		return nil, fmt.Errorf("Invalid ISBN format")
	}
	// strip unwanted characters.
	m := strings.Map(runeToISBNDigit, s)
	// now it should be either 10 or 13 digits
	is13 := len(m) == 13
	if len(m) != 10 && !is13 {
		return nil, fmt.Errorf("Invalid ISBN digit count")
	}
	parsed := &ISBN{is13: is13, digits: [9]byte{}}
	// if 13, check prefix is 978
	offset := 0
	if is13 {
		// allowed prefixes? 978 and 979?
		parsed.prefix = [3]byte{m[0], m[1], m[2]}
		if !isAllowedPrefix(parsed.prefix) {
			return nil, fmt.Errorf("Unexpected ISBN-13 prefix: %s", s[0:3])
		}
		offset = 3
	}

	for i, c := range []byte(m[offset:]) {
		if c == 10 && (is13 || i != 9) {
			return nil, fmt.Errorf("Unexpected character in ISBN (X can only be the final digit of an ISBN-10)")
		}
		if i == 9 {
			parsed.checksum = c
		} else {
			parsed.digits[i] = c
		}
	}
	if !parsed.isValid() {
		return nil, fmt.Errorf("ISBN checksum was incorrect")
	}
	return parsed, nil
}

func isAllowedPrefix(p [3]byte) bool {
	s := p[:]
	for i := range allowedISBN13Prefixes {
		if bytes.Equal(s, allowedISBN13Prefixes[i]) {
			return true
		}
	}
	return false
}

// isValid ensures the given checksum matches what it should be
func (n *ISBN) isValid() bool {
	if n.is13 {
		return check13(n.prefix, n.digits) == n.checksum
	}
	return check10(n.digits) == n.checksum
}

// returns the checksum digit value of the nine digits using
// the ISBN-10 checksum algorithm
func check10(digits [9]byte) byte {
	sum := 0
	for i, d := range digits {
		sum += int(d) * (10 - i)
	}
	m := sum % 11
	if m == 0 {
		return 0
	}
	return byte(11 - m)
}

// returns the checksum digit value of the nine digits using
// the ISBN-13 checksum algorithm, prefix *assumed* to be 978
func check13(prefix [3]byte, digits [9]byte) byte {
	sum := 1*int(prefix[0]) + 3*int(prefix[1]) + 1*int(prefix[2])
	var weight int
	for i, d := range digits {
		weight = 1
		if i%2 == 0 { // even indices weigh 3 (we started from digit 4: weight = 3)
			weight = 3
		}
		sum += int(d) * weight
	}
	m := sum % 10
	if m == 0 {
		return 0
	}
	return byte(10 - m)
}

// To10 returns the ISBN-10 version of this ISBN, if it already is
// ISBN-10, this returns it's input
// Note, that we keep the prefix, so if this was a 979 prefixed ISBN-13
// `myISBN13.To10().To13()` is a lossless operation
func (n *ISBN) To10() *ISBN {
	if !n.is13 {
		return n
	}
	return &ISBN{
		is13:     false,
		prefix:   n.prefix, // keep the prefix anyway, in case we convert back
		digits:   n.digits,
		checksum: check10(n.digits),
	}
}

// To13 returns the ISBN-13 version of this ISBN, if it already is
// ISBN-13, this returns it's input
func (n *ISBN) To13() *ISBN {
	if n.is13 {
		return n
	}
	prefix := n.prefix
	if !isAllowedPrefix(prefix) {
		prefix = [3]byte{0, 0, 0}
		copy(prefix[:], allowedISBN13Prefixes[0])
	}
	return &ISBN{
		is13:     true,
		prefix:   prefix,
		digits:   n.digits,
		checksum: check13(prefix, n.digits),
	}
}

// Is13 checks if the ISBN is an ISBN-13
func (n *ISBN) Is13() bool {
	return n.is13
}

// Is10 checks if the ISBN is an ISBN-10
func (n *ISBN) Is10() bool {
	return !n.is13
}

// String formats ISBN-10 as just the digits, ISBN-13 gets a single
// hyphen after the prefix
func (n *ISBN) String() string {
	base := make([]byte, 10)
	for i, d := range n.digits {
		base[i] = isbnDigitToByte(d)
	}
	base[9] = isbnDigitToByte(n.checksum)
	if n.is13 {
		pre := make([]byte, 3)
		for i, d := range n.prefix {
			pre[i] = isbnDigitToByte(d)
		}
		return string(pre) + "-" + string(base)
	}
	return string(base)
}

// EquivalientTo checks equivalence, not strict equality
func (n *ISBN) EquivalientTo(other *ISBN) bool {
	if other == nil || n == nil {
		return false
	}
	// the digits should match, we don't care about 13 or 10, or the checksum
	for i, d := range n.digits {
		if d != other.digits[i] {
			return false
		}
	}
	return true
}

// ToURN retusn the string urn for this ISBN
func (n *ISBN) ToURN() string {
	return urnPrefix + n.String()
}

// Canonical returns the urn form of the ISBN-13 version
func (n *ISBN) Canonical() string {
	return n.To13().ToURN()
}
