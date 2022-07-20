package repo

import (
	"fmt"
	"postParser/internal/logger"
	"regexp"
	"strings"
	"unicode"
)

const (
	BoundaryField  = "boundary="
	Disposition    = "Content-Disposition: form-data; name=\"\"; filename=\"\"\r\nContent-Type: text/plain\r\n"
	CType          = "Content-Type: "
	Sep            = string(byte(13)) + string(byte(10))
	MaxHeaderLimit = 2 + 51 + 2 + len(Disposition) + 40 + 2 + len(CType) + 20 + 2
	MaxLineLimit   = len(Disposition) + 30
)

type Lines struct {
	CurLines []string
	IsWhole  bool
}

func NewLines(cl []string, i bool) Lines {
	return Lines{
		CurLines: cl,
		IsWhole:  i,
	}
}

func IsGraphic(r rune) bool {
	return unicode.IsGraphic(r)
}

func LineEndPosLimit(s string, fromIndex, limit int) int {
	i := fromIndex

	for i < fromIndex+limit {
		i++
		if s[i] == 13 {
			return i
		}

	}
	return -1
}

func FindPrintPosRight(s string, fromIndex, limit int) int {
	i := fromIndex

	for i < fromIndex+limit {
		i++
		if s[i] != 10 && s[i] != 13 {
			return i
		}
	}
	return -1

}

func LineStartPosLimit(s string, fromIndex, limit int) int {
	i := fromIndex

	for i > fromIndex-limit {
		i--
		if s[i] == 10 {
			return i + 1
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

func LastPrintPosLimit(s string, fromIndex, limit int) int {
	last := len(s) - 1

	for last > fromIndex-limit {

		if s[last] != 10 && s[last] != 13 {
			break
		}
		last--
	}
	return last
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

		bRoot = s[start:LineEndPosLimit(s, start, 100)]

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

func IsLinePrev(s string, fromIndex, limit int) bool {

	if limit > fromIndex {

		return false
	}
	for i := fromIndex; i >= fromIndex-limit; i-- {

		if s[i] == 10 {

			return true
		}
	}

	return false
}
func GetLinesRev(s string, limit int, voc Vocabulaty) []string {

	n := 0
	s = strings.Trim(s, Sep)
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
		line = strings.Trim(line, Sep)
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

func GetLinesFw(s string, prevLines []string, limit int, voc Vocabulaty) Lines {
	isWhole := false
	wholeLine := ""
	st := strings.Trim(s, Sep)
	if st != s {

		//logger.L.Infof("trimming done. \ns was %q\ns became %q\n", s, st)

		s = st
		isWhole = true
	}
	p := 0
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

		for p < len(s) {
			if byte(s[p]) == 13 {
				break
			}
			sb.WriteByte(s[p])
			p++

		}

		line := sb.String()
		line = strings.Trim(line, Sep)

		//	logger.L.Infof("in repo.GetLinesFw line: %q\n", line)

		if i == 0 && len(prevLines) > 0 && !isWhole {
			wholeLine += prevLines[0]
		}
		wholeLine += line

		logger.L.Infof("in repo.GetLinesFw wholeLine after trimming: %q\n", line)

		if i == 0 &&
			(wholeLine == voc.Boundary.Prefix+voc.Boundary.Root ||
				(strings.Contains(wholeLine, voc.CDisposition) &&
					((!strings.Contains(wholeLine, voc.FileName) && r1.MatchString(wholeLine)) ||
						(strings.Contains(wholeLine, voc.FileName) && r2.MatchString(wholeLine)))) ||
				(strings.Contains(wholeLine, voc.CType))) {

			logger.L.Infof("in repo.GetLinesFw i = %d line : %q\n", i, line)
			lines = append(lines, line)

			if strings.Contains(line, voc.CType) {
				break
			}

			s = s[p:]
			p = 0
			sb.Reset()
			s = strings.TrimLeft(s, Sep)

		} else if i == 0 {
			break
		}
		if i == 1 &&
			(strings.Contains(line, voc.CDisposition) ||
				(!strings.Contains(line, voc.FileName) && r1.MatchString(line)) ||
				(strings.Contains(line, voc.FileName) && r2.MatchString(line))) ||
			(strings.Contains(line, voc.CType)) {
			lines = append(lines, line)

			if strings.Contains(line, voc.CType) {
				break
			}

			s = s[p:]
			p = 0
			sb.Reset()
			s = strings.TrimLeft(s, Sep)

		} else if i == 1 {
			break
		}
		if i == 2 && strings.Contains(line, voc.CType) {

			lines = append(lines, line)
		}
	}

	return NewLines(lines, isWhole)
}

func JoinLines(prevLines []string, curLines Lines) []string {

	lines := make([]string, 0)
	switch curLines.IsWhole {
	case false:
		if len(curLines.CurLines) > 0 && len(prevLines) > 0 {
			if len(prevLines) > 1 {
				for i := len(prevLines); i > 1; i-- {
					lines = append(lines, prevLines[i-1])
				}
			}
			line := prevLines[0] + curLines.CurLines[0]
			lines = append(lines, line)
			for i := 1; i < len(curLines.CurLines); i++ {
				lines = append(lines, curLines.CurLines[i])
			}
		}

	default:
		if len(prevLines) == 3 {

			for i := len(prevLines); i > 0; i-- {

				lines = append(lines, prevLines[i])

			}
			break

		}
		if len(curLines.CurLines) == 3 {

			lines = append(lines, curLines.CurLines...)
			break

		}
		for i := len(prevLines); i >= 1; i-- {
			lines = append(lines, prevLines[i-1])
		}
		lines = append(lines, curLines.CurLines...)

	}
	return lines
}

func FindForm(s string, voc Vocabulaty) (fo, fi string) {
	fo, fi = "", ""
	if strings.Contains(s, voc.FormName) {
		fo = s[strings.Index(s, voc.FormName)+len(voc.FormName)+1 : FindNext(s, "\"", strings.Index(s, voc.FormName)+len(voc.FormName)+1)]
		if strings.Contains(s, voc.FileName) {
			fi = s[strings.Index(s, voc.FileName)+len(voc.FileName)+1 : FindNext(s, "\"", strings.Index(s, voc.FileName)+len(voc.FileName)+1)]
		}
	}
	return fo, fi
}

func FindFirstBodyPos(s string, prevLines, lines []string) int {
	nextLineBeginPos, firstBodyPos := 0, 0

	if s[0] == 10 || s[0] == 13 {

		firstBodyPos += FindPrintPosRight(s, 0, MaxLineLimit)
		s = s[firstBodyPos:]
	}

	curLineNum := len(lines)

	logger.L.Infof("in repo.FindFirstBodyPos \nprevLines : %q\ncurLines : %q\n", prevLines, lines)

	for i := 0; i < curLineNum; i++ {

		nextLineBeginPos = LineEndPosLimit(s, nextLineBeginPos, MaxLineLimit) + 2
		s = s[nextLineBeginPos:]
		firstBodyPos += nextLineBeginPos
		nextLineBeginPos = 0
	}
	logger.L.Infof("in repo.FindFirstBodyPos returning firstBodyPos = %d\n", firstBodyPos)

	return firstBodyPos

}

func FindLastBodyPos(s string, f int, b Boundary) (int, string) {
	lens := len(s)
	i := lens

	if strings.Contains(s, b.Prefix+b.Root) &&
		strings.Index(s, b.Prefix+b.Root) > 2 {

		return strings.Index(s, b.Prefix+b.Root) - 2, ""
	}

	switch c := s[i-1]; {
	case c == 10 || c == 13:
		i = LastPrintPosLimit(s, i, MaxLineLimit) + 1
		return i, s[i+1:]

	default:
		p := LineStartPosLimit(s, i, MaxLineLimit)
		if i < 0 {
			return lens, ""
		}
		line := s[p:]
		if len(line) < len(b.Prefix+b.Root) &&
			s[i:] == (b.Prefix + b.Root)[len(s[i:]):] {
			i -= 2
			return i, s[i+1:]
		}
	}
	return lens, ""
}

func GetPrevLineLimit(s string, startPos, limit int) string {
	line := ""
	lens := len(s)
	i := lens

	for i > lens-1-limit {
		i--
		if s[i] != 10 && s[i] != 13 {
			break
		}
	}
	s = s[:i+1]
	lens = len(s)
	i = lens

	for i > lens-1-limit {
		i--
		if s[i] == 10 {
			break
		}
	}
	line = s[i+1:]

	return line
}
