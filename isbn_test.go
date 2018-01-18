package isbn

import (
	"testing"
)

// most of the testing data for this package was lifted from:
// https://github.com/moraes/isbn/blob/e6388fb1bfd58164792a3b8da0ccc01e7140b52c/isbn_test.go

type test struct {
	isbn10 string
	isbn13 string
	valid  bool
}

// NB for all "valid" tests, the ISBN-10 and ISBN-13 values should
// be equivalents of each other.
var tests = []test{
	// Calvin and Hobbes, 1987
	{"0836220889", "9780836220889", true},
	// Something Under the Bed Is Drooling, 1988
	{"0836218256", "9780836218251", true},
	// Yukon Ho!, 1989
	{"0836218353", "9780836218350", true},
	// Weirdos from Another Planet!, 1990
	{"1449407102", "9781449407100", true},
	// Scientific Progress Goes 'Boink', 1991
	{"0836218787", "9780836218787", true},
	// Attack of the Deranged Mutant Killer Monster Snow Goons, 1992
	{"0836218833", "9780836218831", true},
	// The Days are Just Packed, 1993
	{"0836217357", "9780836217353", true},
	//  The Tales of Henry James (Literature & Life) USA edition (it has an X checksum!)
	{"080442957X", "9780804429573", true},
	// different representations of the above
	{"0-8044-2957-X", "978-0-8044-2957-3", true},
	{"0-8044-2957-x", "978-0-8044-2957-3", true},
	{"080442957-X", "urn:isbn:978-0-8044-2957-3", true},
	// nb I don't care about hyphens as long as there are not more than 4!
	{"urn:isbn:0-8-0-4-42957x", "urn:isbn:9780804429573", true},
	// spaces also OK (<= 4)
	{"urn:isbn:080 442 95 7x", "urn:isbn:97 808 0442 9573", true},
	{"urn:isbn:080 442-95-7x", "urn:isbn:97-808-0442 9573", true},
	// invalid: bad space/hypen
	{"urn:isbn:00 4 4 2 95 7x", "urn:isbn:97 8-0-8 0-4-4-2 9-5-7-3", false},
	// invalid: character set
	{"08044295XX", "97808X4429573", false},
	{"badformat!", "notremotelyok", false},
	// Invalid: too many characters
	{"08362208891", "97808362208891", false},
	{"08362182562", "97808362182512", false},
	{"08362183533", "97808362183503", false},
	{"08362186204", "97804391374924", false},
	{"08362187875", "97808362187875", false},
	{"08362188336", "97808362188316", false},
	{"08362173577", "97808362173537", false},
	{"urn:isbn:08362173577", "urn:isbn:97808362173537", false},
	{"urn:isbn:0836-2173577", "urn:isbn:978 08362173537", false},
	// Invalid: too few characters
	{"083622088", "978083622088", false},
	{"083621825", "978083621825", false},
	{"083621835", "978083621835", false},
	{"083621862", "978043913749", false},
	{"083621878", "978083621878", false},
	{"083621883", "978083621883", false},
	{"083621735", "978083621735", false},
	{"urn:isbn:083621883", "urn:isbn:978083621883", false},
	{"urn:isbn:0 8 3 6 21735", "urn-isbn:978-0-836-2173-5", false},
	// Invalid: bad check digit
	{"0836220888", "9780836220880", false},
	{"0836218255", "9780836218252", false},
	{"0836218352", "9780836218351", false},
	{"0836218629", "9780439137493", false},
	{"0836218786", "9780836218788", false},
	{"0836218832", "9780836218832", false},
	{"0836217356", "9780836217354", false},
}

func checkStringEqual(t *testing.T, errMsg, s1, s2 string) {
	if s1 != s2 {
		t.Errorf("%s (`%s` vs `%s`)", errMsg, s1, s2)
	}
}

func TestISBN(t *testing.T) {
	for _, v := range tests {
		n10, err10 := Parse(v.isbn10)
		n13, err13 := Parse(v.isbn13)
		if v.valid {
			ok := true
			if err10 != nil {
				t.Errorf("Failed to parse `%s`, error: %s", v.isbn10, err10)
				ok = false
			}
			if err13 != nil {
				t.Errorf("Failed to parse `%s`, error: %s", v.isbn13, err13)
				ok = false
			}
			if !ok {
				continue
			}
			if !n10.EquivalientTo(n13) {
				t.Errorf("Equivalence check failed for `%s` and `%s`", n10, n13)
			}
			checkStringEqual(t, "Canonical forms should match", n10.Canonical(), n13.Canonical())

			n13to10 := n13.To10()
			n10to13 := n10.To13()
			checkStringEqual(t, "String forms of the ISBN-10 and the ISBN-13 converted to 10 should be the same", n10.String(), n13to10.String())
			checkStringEqual(t, "String forms of the ISBN-13 and the ISBN-10 converted to 13 should be the same", n13.String(), n10to13.String())
		} else {
			if err10 == nil {
				t.Errorf("Incorrect parsed: %s", v.isbn10)
			}
			if err13 == nil {

				t.Errorf("Incorrect parsed: %s", v.isbn13)
			}
		}
	}
}

// from amazon https://www.amazon.co.uk/PIM-979-ISBN-13-Test/dp/B0019RL99E
const test979isbn = `979-5000000235`

func TestISBN13_979_Prefix(t *testing.T) {
	n, err := Parse(test979isbn)
	if err != nil {
		t.Errorf("Failed to parse 979-prefix ISBN-13 `%s`, error: %s", test979isbn, err)
		return
	}
	// it should conserve the prefix.
	checkStringEqual(t, "Conversion of ISBN-13 To10() and back should be lossless", n.String(), n.To10().To13().String())
}
