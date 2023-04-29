package repo

import "strings"

// GetLastIndex returns index of last occurence of occ in s.
// Tested in stringOps_test.go
func GetLastIndex(s, occ string) int {
	index, n := 0, 0
	for strings.Contains(s, occ) {
		i := strings.Index(s, occ)

		if n > 0 {
			index += i + 1
		} else {
			index += i
		}
		s = s[i+1:]
		n++
	}
	return index
}

// SingleSpaces returns s with no repeated whitespaces
func SingleSpaces(s string) string {
	res := ""
	words := strings.Fields(s)
	for i, v := range words {

		if i > 0 && i < len(words) {
			res = res + " "
		}

		res = res + v

	}
	return res
}
