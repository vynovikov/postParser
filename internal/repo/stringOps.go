package repo

import (
	"fmt"
	"postParser/internal/logger"
	"regexp"
	"strings"
	"unicode"
)

const (
	BoundaryField = "boundary="
	Disposition   = "Content-Disposition: form-data; name=\"\"; filename=\"\"\r\nContent-Type: text/plain\r\n"
)

func IsGraphic(r rune) bool {
	return unicode.IsGraphic(r)
}

func FindFirstNonPrintPosLimit(s string, fromIndex, limit int) int {

	for i, v := range s {
		if i > fromIndex &&
			i < limit+fromIndex &&
			!unicode.IsGraphic(v) {
			return i
		}
	}
	return -1
}

func FindNext(s, occ string, fromIndex int) int {
	lens := len(s)

	nextIndex := 0

	if fromIndex >= 0 && fromIndex < lens {

		nextIndex = strings.Index(s[fromIndex+1:], occ) + fromIndex + 1

		if nextIndex > fromIndex {

			return nextIndex

		}
	}
	if nextIndex > lens {
		return -1
	}

	return -1
}

func FindLast(s, occ string) int {

	lastPos := 0

	for strings.Contains(s, occ) {

		curPos := FindNext(s, occ, 0)

		if curPos < 0 {

			break

		}
		s = s[curPos+1:]

		lastPos += curPos + 1

	}

	return lastPos
}

func FindBoundary(s string) (string, string) {

	bPrefix, bRoot := "", ""

	if strings.Contains(s, BoundaryField) {
		start := strings.Index(s, BoundaryField) + len(BoundaryField)
		ss := s[start+1:]

		bRoot = s[start:FindFirstNonPrintPosLimit(s, start, 100)]

		secBoundaryIndex := strings.Index(ss, bRoot)

		bPrefix = StringBefore(ss, secBoundaryIndex-1)

	}

	return bPrefix, bRoot
}

func NumHyphens(s string, fromIndex int) int {
	i := fromIndex
	n := 0
	for {
		i--
		if s[i:fromIndex] != strings.Repeat("-", fromIndex-i) {
			break
		}
		n++
	}
	return n
}

func FindPrev(s, occ string, fromIndex int) bool {
	for i := fromIndex - len(occ); rune(s[i]) == 10; i-- {

		logger.L.Infof("in repo.Findprev observing %s\n", s[i:fromIndex])

		if s[i:fromIndex] == occ {
			logger.L.Info("string found")
			return true
		}
	}
	logger.L.Info("string not found")
	return false
}

func FindPrevIndex(s, occ string, fromIndex int) int {
	for i := fromIndex - len(occ); rune(s[i]) == 10; i-- {

		logger.L.Infof("in repo.Findprev observing %s\n", s[i:fromIndex])

		if s[i:fromIndex] == occ {
			logger.L.Info("string found")
			return i
		}
	}
	logger.L.Info("string not found")
	return -1
}
func FindPrevString(s, occ string, fromIndex int) string {
	for i := fromIndex - len(occ); rune(s[i]) == 10; i-- {

		logger.L.Infof("in repo.Findprev observing %s\n", s[i:fromIndex])

		if s[i:i+len(occ)] == occ {
			logger.L.Info("string found")
			return s[i : i+len(occ)]
		}
	}
	logger.L.Info("string not found")
	return ""
}

func StringBefore(s string, fromIndex int) string {
	ss := ""
	if fromIndex <= 0 {
		return ss
	}
	for i := fromIndex; s[i] != 10 || i == 0; i-- {
		ss += string(s[i])
	}
	rss := Reverse(ss)
	return rss
}

func Reverse(ss string) string {
	rss := ""
	for i := len(ss) - 1; i >= 0; i-- {
		rss += string(ss[i])
	}
	return rss
}

func IsLine(s string, fromIndex, limit int) bool {

	if limit > fromIndex {
		//logger.L.Infof("in repo.IsLine limit %d is greater than fromIndex %d", limit, fromIndex)
		return false
	}
	for i := fromIndex; i >= fromIndex-limit; i-- {
		//logger.L.Infof("%v ", s[i])
		if s[i] == 10 {
			//logger.L.Infof("in repo.IsLine \\n found on pos %d", i)
			return true
		}
	}

	return false
}
func GetLines(s string, limit int, voc Vocabulaty) []string {

	sep := string(byte(13)) + string(byte(10))
	n := 0
	s = strings.Trim(s, sep)
	p := len(s) - 1
	var sb strings.Builder
	lines := make([]string, 0)

	r1, err := regexp.Compile("Content-Disposition:\\sform-data;\\sname=\"\\w+\"")
	if err != nil {
		fmt.Println(err)
	}
	r2, err := regexp.Compile("Content-Disposition:\\sform-data;\\sname=\"\\w+\";\\sfilename=\"\\w+\"")
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < 3; i++ {

		for (s[p] != 10) && (n < limit) {

			sb.WriteByte(s[p])
			p--
			n++

		}

		line := sb.String()
		line = strings.Trim(line, sep)
		line = Reverse(line)
		lineLen := len(line)

		if i == 0 &&
			((lineLen <= len(voc.CType) && line == voc.CType[:lineLen]) ||
				(lineLen > len(voc.CType) && line[:len(voc.CType)] == voc.CType) ||
				(lineLen <= len(voc.CDisposition) && line == voc.CDisposition[:lineLen]) ||
				(lineLen > len(voc.CDisposition) && line[:len(voc.CDisposition)] == voc.CDisposition) ||
				(lineLen <= len(voc.Boundary.Prefix+voc.Boundary.Root) && line == (voc.Boundary.Prefix + voc.Boundary.Root)[:lineLen]) ||
				(lineLen > len(voc.Boundary.Prefix+voc.Boundary.Root) && line[:len(voc.Boundary.Prefix+voc.Boundary.Root)] == voc.Boundary.Prefix+voc.Boundary.Root)) {

			lines = append(lines, line)

			if i == 0 &&
				(lineLen <= len(voc.Boundary.Prefix+voc.Boundary.Root) &&
					line == (voc.Boundary.Prefix + voc.Boundary.Root)[:lineLen]) ||
				(lineLen > len(voc.Boundary.Prefix+voc.Boundary.Root) &&
					line[:len(voc.Boundary.Prefix+voc.Boundary.Root)] == voc.Boundary.Prefix+voc.Boundary.Root) {
				break
			}

			s = s[:p]
			p = len(s) - 1
			sb.Reset()
		} else if i == 0 {
			lines = make([]string, 0)
			break
		}
		if i == 1 &&
			((strings.Contains(line, voc.FileName) && (r2.MatchString(line))) ||
				!strings.Contains(line, voc.FileName) && (r1.MatchString(line))) ||
			(line == voc.Boundary.Prefix+voc.Boundary.Root) {

			lines = append(lines, line)

			if (lineLen <= len(voc.Boundary.Prefix+voc.Boundary.Root) &&
				line == (voc.Boundary.Prefix + voc.Boundary.Root)[:lineLen]) ||
				(lineLen > len(voc.Boundary.Prefix+voc.Boundary.Root) &&
					line[:len(voc.Boundary.Prefix+voc.Boundary.Root)] == voc.Boundary.Prefix+voc.Boundary.Root) {
				break
			}

			s = s[:p]
			p = len(s) - 1
			sb.Reset()
			continue

		} else if i == 1 {
			lines = make([]string, 0)
			break
		}
		if i == 2 && (line == voc.Boundary.Prefix+voc.Boundary.Root) {
			lines = append(lines, line)
		} else if i == 2 {
			lines = make([]string, 0)
		}
	}

	return lines
}
