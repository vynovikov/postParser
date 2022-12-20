package repo

import "strings"

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
