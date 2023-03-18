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

// returns string with single whitespaces
func SingleSpaces(s string) string {
	res := ""
	words := strings.Fields(s)
	//logger.L.Infof("in repo.SingleSpace words: %q\n", words)
	for i, v := range words {

		if i > 0 && i < len(words) {
			res = res + " "
		}

		res = res + v

	}
	//logger.L.Infof("in repo.SingleSpace res: %q\n", res)
	return res
}
