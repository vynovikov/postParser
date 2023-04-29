package repo

import (
	"bytes"
	"regexp"
)

// IsCDRight returns true if b is part of Content-Disposition header line cut from right.
// Tested in regexpOps_test.go
func IsCDRight(b []byte) bool {
	CD := []byte("Content-Disposition: form-data; name=\"")
	if len(b) <= len(CD) && bytes.Contains(CD, b) {
		return true
	}
	if len(b) > len(CD) {
		if bytes.Contains(b, CD) {

			switch bytes.Count(b[len(CD):], []byte("\"")) {

			case 0:

				r0 := regexp.MustCompile(`^[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+$`)

				return r0.Match(b[len(CD):])

			case 1:
				r1 := regexp.MustCompile(`^[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"`)

				index := RepeatedIntex(b, []byte("\""), 2)

				return r1.Match(b[len(CD):]) &&
					(BeginningEqual([]byte("; filename="), b[index+1:]) ||
						index-1 == len(b))
			case 2:
				r2 := regexp.MustCompile(`^[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"; filename="[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]*$`)
				return r2.Match(b[len(CD):])
			case 3:
				r3 := regexp.MustCompile(`^[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"; filename="[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"$`)
				return r3.Match(b[len(CD):])
			}

		}
		return false
	}
	return false
}

// Sufficiency determines whether b is header for string data or for file data
func Sufficiency(b []byte) sufficiency {
	r0 := regexp.MustCompile(`^Content-Disposition: form-data; name="[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"$`)
	r1 := regexp.MustCompile(`^Content-Disposition: form-data; name="[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"; filename="[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"$`)

	if r0.Match(b) {
		return Sufficient
	}
	if r1.Match(b) {
		return Insufficient
	}

	return Incomplete
}

// IsCDLeft returns true if b is part of Content-Disposition header line cut from left.
// Tested in regexpOps_test.go
func IsCDLeft(b []byte) bool {
	CD := []byte("Content-Disposition: form-data; name=")

	switch bytes.Count(b, []byte("\"")) {
	case 1:
		if len(b) == 1 {
			return bytes.Contains(b, []byte("\""))
		}
		r1 := regexp.MustCompile(`^[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"$`)
		return r1.Match(b)
	case 2:
		CDF := []byte("; filename=")
		pre := b[:bytes.Index(b, []byte("\""))]
		r2 := regexp.MustCompile(`^"[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"$`)

		return (EndingOf(CD, pre) || EndingOf(CDF, pre)) && r2.Match(b[len(pre):])
	case 3:
		colonIndex := bytes.Index(b, []byte("\""))
		r30 := regexp.MustCompile(`"; filename="[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"$`)
		r31 := regexp.MustCompile(`^[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+$`)

		if colonIndex > 0 {

			return r30.Match(b) && r31.Match(b[:colonIndex])
		}
		return r30.Match(b)
	case 4:
		colonIndex := bytes.Index(b, []byte("\""))
		r4 := regexp.MustCompile(`"[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"; filename="[a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+"$`)

		if colonIndex > 0 {
			return r4.Match(b) && EndingOf(CD, b[:colonIndex])
		}
		return r4.Match(b)
	}

	return false
}

// IsCTRight returns true if b is part of Content-Type header line cut from right.
// Tested in regexpOps_test.go
func IsCTRight(b []byte) bool {

	CT := []byte("Content-Type:")

	spaceIndex := bytes.Index(b, []byte(" "))

	r0 := regexp.MustCompile(`^[a-zA-z0-9_.%^&\$\+\!\*]*\/?[a-zA-z0-9_.%^&\$\+\!\*]*$`)

	if len(b) < 1 {
		return true
	}
	if spaceIndex < 0 {
		return BeginningEqual(CT, b)
	}
	return BeginningEqual(CT, b[:spaceIndex]) && r0.Match(b[spaceIndex+1:])

}

// IsCTLeft returns true if b is part of Content-Type header line cut from left.
// Tested in regexpOps_test.go
func IsCTLeft(b []byte) bool {

	CT := []byte("Content-Type:")

	spaceIndex := bytes.Index(b, []byte(" "))

	r0 := regexp.MustCompile(`^[a-zA-z0-9_.%^&\$\+\!\*]*\/?[a-zA-z0-9_.%^&\$\+\!\*]+$`)

	if spaceIndex < 0 { // line is only part after space

		return len(b) < 13 && r0.Match(b)
	}

	return EndingOf(CT, b[:spaceIndex]) && r0.Match(b[spaceIndex+1:])
}

// IsCTFull returns true if b in Content-Type header line.
// Tested in regexpOps_test.go
func IsCTFull(b []byte) bool {
	r0 := regexp.MustCompile(`^Content-Type: [a-zA-zа-яА-Я0-9_.-:@#%^&\$\+\!\*\(\[\{\)\]\}]+$`)

	return r0.Match(b)
}
