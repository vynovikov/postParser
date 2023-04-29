// Helper pachage for types and functions
package repo

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	BoundaryField  = "boundary="
	CD             = "Content-Disposition"
	CType          = "Content-Type"
	Sep            = "\r\n"
	MaxLineLimit   = 100
	MaxHeaderLimit = 210
)

// FindNext returns index of occ next to fromIndex
func FindNext(b, occ []byte, fromIndex int) int {
	return bytes.Index(b[fromIndex:], occ) + fromIndex
}

// FinsBoundary returns Boundary found in b
// Tested in byteOps_test.go
func FindBoundary(b []byte) Boundary {

	var err error
	bPrefix, bRoot := []byte{}, []byte{}

	if bytes.Contains(b, []byte(BoundaryField)) {

		startIndex := bytes.Index(b, []byte(BoundaryField)) + len(BoundaryField)

		bRoot = LineRightLimit(b, startIndex, 70)

		bb := b[startIndex+1:]

		secBoundaryIndex := bytes.Index(bb, bRoot) - 1

		bPrefix, err = GetCurrentLineLeft(bb, secBoundaryIndex, MaxLineLimit)

		if err != nil {
			return Boundary{}
		}
	}
	return Boundary{
		Prefix: bPrefix,
		Root:   bRoot,
	}

}

// LineRightEndIndexLimit returns last index before CR to the right of fromIndex in b.
// Tested in byteOps_test.go
func LineRightEndIndexLimit(b []byte, fromIndex, limit int) int {
	i := fromIndex

	for i < len(b) && i < fromIndex+limit {
		if b[i] == 13 {
			return i - 1
		}
		i++

	}
	return -1
}

// LineRightLimit returns byte slice to the right of fromIndex in b. Stops when CR found of limit exceeded
// Tested in byteOps_test.go
func LineRightLimit(b []byte, fromIndex, limit int) []byte {
	bb := make([]byte, 0)

	if fromIndex < 0 {
		return nil
	}

	for i := fromIndex; b[i] != 13 && i < fromIndex+limit; i++ {
		bb = append(bb, b[i])
	}
	if len(bb) == limit {
		return nil
	}

	return bb
}

// CurrentLineFirstPrintIndexRight returns index of first printable symbol of b or error if limit exceeded.
func CurrentLineFirstPrintIndexRight(b []byte, limit int) (int, error) {

	p := 0
	if len(b) == 0 {
		return -1, errors.New("passed byte slice with zero length")
	}

	for i := 0; (b[i] == 13 || b[i] == 10) && i < limit && i < len(b)-1; i++ {
		p++
	}
	if p == limit {
		return -1, errors.New("no printable characters before limit")
	}

	return p, nil
}

// GetCurrentLineRight returns byte slice to the right of fromIndex in b.
// Stops when CR found
// Returns error if limit exceeded.
func GetCurrentLineRight(b []byte, fromIndex, limit int) ([]byte, error) {
	bb, i, lenb := make([]byte, 0), fromIndex, len(b)

	if lenb == 0 {
		return bb, errors.New("passed byte slice with zero length")
	}

	// checking if line is ending part of last boundary
	if lenb < limit && LineRightEndIndexLimit(b, fromIndex, limit) < 0 {

		for i < lenb {
			bb = append(bb, b[i])
			i++
		}
		return bb, nil
	}

	for i < fromIndex+limit && b[i] != 13 {

		bb = append(bb, b[i])

		if i == fromIndex+limit {
			return bb, errors.New("line limit exceeded. No separator met")
		}
		if i == lenb-1 {
			return bb, errors.New("body end reached. No separator met")
		}
		i++
	}
	if b[len(bb)+1] == 10 {
		bb = append(bb, []byte("\r\n")...)
		return bb, nil
	}

	return bb, fmt.Errorf("in repo.GetCurrentLineRight bb is not ending with LF")
}

// Tested in byteOps_test.go
func SingleLineRightUnchanged(b []byte, limit int) ([]byte, error) {

	return GetCurrentLineRight(b, 0, limit)
}

// Returns number of non-pritnable syblols, byte slice from first printable syblol to CR.
// Returns error if limit exceeded.
// Tested in byteOps_test.go
func SingleLineRightTrimmed(b []byte, limit int) (int, []byte, error) {
	trimmed := 0
	index, err := CurrentLineFirstPrintIndexRight(b, limit)
	if err != nil {
		return trimmed, nil, err
	}
	trimmed += index
	line, err := GetCurrentLineRight(b, index, limit-index)
	return trimmed, line, err
}

// CurrentLineFirstPrintIndexLeft returns last printable synbol index of b with limit.
// Returns error if limit exceeded.
// Tested in byteOps_test.go
func CurrentLineFirstPrintIndexLeft(b []byte, limit int) (int, error) {
	lenb := len(b)
	i := lenb - 1
	if lenb == 0 {
		return -1, errors.New("in repo.CurrentLineFirstPrintIndexLeft passed byte slice with zero length")
	}

	for (b[i] == 13 || b[i] == 10) &&
		i > lenb-1-limit &&
		i > 0 {
		i--
	}
	if i == lenb-1-limit {
		return -1, errors.New("in repo.CurrentLineFirstPrintIndexLeft no actual characters before limit")
	}

	return i, nil
}

// GetCurrentLineLeft checks b from fromIndex to the left with given limit.
// When meets LF, returns byte slice from that LF to fromIndex.
// Returns error if limit exceedes.
// Tested in byteOps_test.go
func GetCurrentLineLeft(b []byte, fromIndex, limit int) ([]byte, error) {
	bb, i, lenb := make([]byte, 0), fromIndex, len(b)
	if lenb == 0 {
		return bb, errors.New("in repo.GetCurrentLineLeft passed byte slice with zero length")
	}

	if fromIndex < limit {
		limit = fromIndex
	}

	if len(b) < limit {
		limit = len(b)
	}
	for i > fromIndex-limit && b[i] != 10 {
		bb = append(bb, b[i])
		i--
	}
	if i == fromIndex-limit {
		return nil, errors.New("in repo.GetCurrentLineLeft line limit exceeded. No separator met")
	}
	bbr := Reverse(bb)

	return bbr, nil
}

// SingleLineLeftTrimmed combines CurrentLineFirstPrintIndexLeft and GetCurrentLineLeft
// Tested in byteOps_test.go
func SingleLineLeftTrimmed(b []byte, limit int) ([]byte, error) {
	index, err := CurrentLineFirstPrintIndexLeft(b, limit)
	if err != nil {
		return nil, err
	}
	return GetCurrentLineLeft(b, index, limit)
}

// Reverse returns byte slice which symbols ordered back to front relative to passed one.
// Tested in byteOps_test.go
func Reverse(bb []byte) []byte {

	bbs := make([]byte, 0)

	for i := len(bb) - 1; i >= 0; i-- {
		bbs = append(bbs, bb[i])
	}

	return bbs
}

// Slicer checks b and extracts partial filled dataPiece units from it.
// Tested in byteOps_test.go
func Slicer(b []byte, bou Boundary) (AppPieceUnit, []AppPieceUnit, AppSub) {
	var (
		m []AppPieceUnit
	)
	boundaryCore := GenBoundary(bou)[2:]
	boundaryMiddle := append([]byte("\r\n"), boundaryCore...)
	boundaryMiddle = append(boundaryMiddle, []byte("\r\n")...)
	boundNum := bytes.Count(b, boundaryMiddle)

	aphb := NewAppPieceHeader()
	apbb := NewAppPieceBodyEmpty()
	aphb.SetB(True)

	if bytes.Contains(b, boundaryMiddle) {
		aphb.SetE(False)
		apub := NewAppPieceUnitEmpty()

		fbi := bytes.Index(b, boundaryMiddle)
		apbb.SetBody(b[:fbi])

		if !bytes.Contains(b, []byte("boundary=")) {
			apbb.SetBody(b[:fbi])
			apub.SetAPB(apbb)
			apub.SetAPH(aphb)
		}

		b = b[fbi+len(boundaryMiddle):]

		if boundNum > 1 { // >1 boundaries
			for i := 0; i < boundNum-1; i++ {

				aph := NewAppPieceHeader()
				aph.SetB(False)
				aph.SetE(False)

				ni := bytes.Index(b[1:], boundaryMiddle) + 1
				apb := NewAppPieceBodyFilled(b[:ni])

				apum := NewAppPieceUnit(aph, apb)
				m = append(m, apum)

				b = b[ni+len(boundaryMiddle):]
			}
		}
		// last boundary piece
		apbe := NewAppPieceBodyEmpty()
		aphe := NewAppPieceHeader()
		aphe.SetB(False)
		be := make([]byte, 0)
		if len(b) > MaxHeaderLimit {
			be = b[len(b)-MaxLineLimit:]
		} else {
			be = b
		}
		lenbe := len(be)
		ll := GetLineWithCRLFLeft(be, lenbe-1, MaxLineLimit, bou)

		if len(ll) > 2 && BeginningEqual(ll[2:], boundaryCore) { // last line equal to boundary begginning or vice versa

			apbe := NewAppPieceBodyFilled(b[:len(b)-len(ll)])

			if IsLastBoundary(ll, []byte(""), bou) { // last boundary in last line

				aphe.SetE(False)
				apume := NewAppPieceUnit(aphe, apbe)

				m = append(m, apume)

				return apub, m, AppSub{}
			}

			aphe.SetE(Probably)

			apue := NewAppPieceUnit(aphe, apbe)
			m = append(m, apue)

			as := NewAppSub()
			as.SetBody(ll)
			return apub, m, as
		}
		if (len(ll) == 1 && bytes.Contains(ll, []byte("\r"))) || // last line is CR
			(len(ll) == 2 && bytes.Contains(ll, []byte("\r\n"))) { // last line is LF
			aphe.SetE(Probably)
			apbe.SetBody(b[:len(b)-len(ll)])
			apue := NewAppPieceUnit(aphe, apbe)

			m = append(m, apue)

			as := NewAppSub()
			as.SetBody(ll)

			return apub, m, as
		}

		apbe.SetBody(b)

		aphe.SetE(True)

		apue := NewAppPieceUnit(aphe, apbe)
		m = append(m, apue)

		return apub, m, AppSub{}

	} // 0 boundaries

	switch bytes.Count(b, []byte("\r")) {
	case 0:

		apbb.SetBody(b)

		aphb.SetE(True)

		apub := NewAppPieceUnit(aphb, apbb)

		return apub, nil, AppSub{}

	default:
		if len(b) < MaxLineLimit && (b[len(b)-1] == 13 || (b[len(b)-1] == 10 && b[len(b)-2] == 13)) {
			aphb.SetE(False)
			apbb.SetBody(b)

			apub := NewAppPieceUnit(aphb, apbb)
			return apub, nil, AppSub{}
		}
		be := make([]byte, 0)
		if len(b) > MaxHeaderLimit {
			be = b[len(b)-MaxLineLimit:]
		} else {
			be = b
		}
		lenbe := len(be)

		ll := GetLineWithCRLFLeft(be, lenbe-1, MaxLineLimit, bou)
		if (len(ll) > 2 && BeginningEqual(ll[2:], boundaryCore)) || IsLastBoundary(ll, []byte(""), bou) { // last line equal to boundary begginning or vice versa
			if len(ll) == len(b) {
				apbb.SetBody(b)
			} else {
				apbb.SetBody(b[:len(b)-len(ll)])
			}
			if IsLastBoundary(ll, []byte(""), bou) { // last boundary in last line
				aphb.SetE(False)
				apub := NewAppPieceUnit(aphb, apbb)

				return apub, nil, AppSub{}
			}

			aphb.SetE(Probably)

			apub := NewAppPieceUnit(aphb, apbb)
			as := NewAppSub()
			as.SetBody(ll)

			return apub, nil, as
		}

		if (len(ll) == 1 && bytes.Contains(ll, []byte("\r"))) || // last line is CR
			(len(ll) == 2 && bytes.Contains(ll, []byte("\r\n"))) { // last line is LF
			aphb.SetE(Probably)
			apbb.SetBody(b[:len(b)-len(ll)])
			apub := NewAppPieceUnit(aphb, apbb)
			as := NewAppSub()
			as.SetBody(ll)

			return apub, nil, as

		}
		aphb.SetE(True)
		apbb.SetBody(b)

		apub := NewAppPieceUnit(aphb, apbb)
		return apub, nil, AppSub{}
	}
}

// IsPartlyBoundaryRight returns true if boundary contains b and starts with it.
// Tested in byteOps_test.go
func IsPartlyBoundaryRight(b []byte, bou Boundary) bool {
	boundary := GenBoundary(bou)
	if len(b) > len(boundary) {
		return false
	}
	for i := len(b); i > 0; i-- {
		if len(boundary) > i &&
			b[i-1] != boundary[len(boundary)-1-len(b)+i] {
			return false
		}
	}
	return true
}

// GenBoundary generates byte slice based on given Boundary struct.
// Tested in byteOps_test.go
func GenBoundary(bou Boundary) []byte {
	boundary := make([]byte, 0)
	boundary = append(boundary, []byte("\r\n")...)
	boundary = append(boundary, bou.Prefix...)
	boundary = append(boundary, bou.Root...)
	return boundary
}

// LineLeftPosLimit returns index of first printable symbol from last LF or -1 if limit exceeded
func LineLeftPosLimit(b []byte, fromIndex, limit int) int {

	i := fromIndex

	for i > fromIndex-limit {
		if b[i] == 10 || i == 0 {
			return i + 1
		}
		i--
	}
	return -1
}

// AllPrintalbe returrns true if b has all printable charackters.
// Tested in byteOps_test.go
func AllPrintalbe(b []byte) bool {
	for _, v := range b {
		if !unicode.IsPrint(rune(v)) {
			return false
		}
	}
	return true
}

// NoDigits returns true if b contains no digits.
// Tested in byteOps_test.go
func NoDigits(b []byte) bool {
	for _, v := range b {
		if unicode.IsDigit(rune(v)) {
			return false
		}
	}
	return true
}

// GetLinesLeft looks for header lines in ending of b.
// Returns if any found.
// Tested in byteOps_test.go
func GetLinesLeft(b []byte, limit int, voc Vocabulaty) ([][]byte, error) {
	lines, boundary := make([][]byte, 0), GenBoundary(voc.Boundary)

	if len(b) == 0 {
		return lines, errors.New("in repo.GetLinesLeft passed byte slice with zero length")
	}

	for i := 0; i < 3; i++ {

		lenb := len(b)

		if limit > lenb {
			limit = lenb
		}

		l, err := SingleLineLeftTrimmed(b[lenb-limit:], limit)
		if err != nil {
			return lines, err
		}
		lenl := len(l)

		if i == 0 &&
			((lenl <= len(voc.CType) && bytes.Contains(l, []byte(voc.CType[:lenl]))) || // true if line is part of Content-Type word
				(lenl > len(voc.CType) && bytes.Contains(l, []byte(voc.CType))) || // true if line contains Content-Type word
				(lenl <= len(voc.CDisposition) && bytes.Contains(l, []byte(voc.CDisposition[:lenl]))) || // true if line is part of Content-Disposition word
				(lenl > len(voc.CDisposition) && bytes.Contains(l, []byte(voc.CDisposition))) || // true if line contains Content-Disposition word
				(lenl <= len(boundary) && bytes.Contains(l, boundary[:lenl]))) { // true if line is part of boundary or equal to boundary

			lines = append(lines, l)

			if lenl <= len(boundary) && bytes.Contains(l, boundary[:lenl]) { // true if line is part of boundary or equal to boundary
				break
			}
		} else if i == 0 {
			break
		}
		if i == 1 &&
			(bytes.Contains(l, boundary) || // true if line contains boundary
				(lenl > len(voc.CDisposition) && bytes.Contains(l, []byte(voc.CDisposition)))) { // true if line contains Content-Disposition word
			lines = append(lines, l)

			if bytes.Contains(l, boundary) {
				break
			}
		} else if i == 1 {
			lines = make([][]byte, 0)
		}
		if i == 2 &&
			bytes.Contains(l, boundary) { // true if line contains boundary

			lines = append(lines, l)

			break
		} else if i == 2 {
			lines = make([][]byte, 0)
		}
		b = b[:lenb-lenl-2]
		limit -= lenl

	}

	return lines, nil

}

// GetLinesRightBegin looks for header lines in the beginning of b.
// Returns if any found.
// Tested in byteOps_test.go
func GetLinesRightBegin(b []byte, limit int, bou Boundary) ([][]byte, int, error) {
	//logger.L.Infof("repo.GetLinesRightBegin ivoked with params b: %q, limit= %d, bou: %v\n", b, limit, bou)
	lines, i, cut := make([][]byte, 0), 0, 0

	r1, err := regexp.Compile("Content-Disposition:\\sform-data;\\sname=\"\\w+\"")
	if err != nil {
		return lines, cut, errors.New("in repo.GetLinesLeft error while compiling regexp")
	}
	r2, err := regexp.Compile(".*\"?[a-zA-Z0-9_.-]+\"$")
	if err != nil {
		return lines, cut, errors.New("in repo.GetLinesLeft error while compiling regexp")
	}

	if limit > len(b) {
		limit = len(b)
	}

	for i < 3 && limit > 0 {
		c, l, err := SingleLineRightTrimmed(b[:limit], limit)
		if err != nil {
			return lines, cut, err
		}

		if i == 0 &&
			(EndingOf(GenBoundary(bou), l) || //check if line is last part of boundary
				IsCDLeft(l) || //check if line is Content-Disposityon header line, cut from left
				IsCTLeft(l)) { //check if line is Content-Type header line, cut from left

			lines = append(lines, l)
			cut += c
			if IsCTLeft(l) {

				cut += len(l) + 2
				return lines, cut, nil
			}
		} else if i == 0 {
			return make([][]byte, 0), 0, errors.New("first line \"" + string(l) + "\" is unexpected")
		}
		if i == 1 &&
			(IsCDLeft(l) || ////check if line is Content-Disposityon header line
				IsCTLeft(l)) { //check if line is Content-Type header line

			if r1.Match(l) &&
				(!EndingOf(GenBoundary(bou), lines[0]) ||
					IsCTLeft(l)) && !r2.Match(lines[0]) && !EndingOf(GenBoundary(bou), l) {
				return make([][]byte, 0), 0, errors.New("first line \"" + string(lines[0]) + "\" is unexpected")
			}

			lines = append(lines, l)

			if IsCTLeft(l) {
				break
			}
		} else if i == 1 {
			return make([][]byte, 0), 0, errors.New("second line \"" + string(l) + "\" is unexpected")
		}

		if i == 2 &&
			IsCTLeft(l) { //check if line contains Conternt-Type word

			lines = append(lines, l)

			cut += len(l) + 2

			break
		}

		i++
		lenl := len(l) + 2
		limit -= lenl
		cut += lenl
		if len(b) < len(l)+2 || limit < 0 {
			break
		}

		b = b[len(l)+2:]
	}

	return lines, cut, nil
}

// GetLinesRightMiddle looks for header lines in the middle of b.
// Returns if any found.
// Tested in byteOps_test.go
func GetLinesRightMiddle(b []byte, limit int) ([][]byte, int, error) {
	lenb, lenl := len(b), 0
	lines, i, cut := make([][]byte, 0), 0, 0

	if len(b) == 0 {
		return lines, cut, fmt.Errorf("in repo.GetLinesRightMiddle zero length byte slice passed")
	}
	if limit > len(b) {
		limit = len(b)
	}

	for i < 2 && limit > 0 {
		c, l, err := SingleLineRightTrimmed(b[:limit], limit)

		if i == 0 &&
			IsCDRight(l) {

			lines = append(lines, l)
			cut += c
			lenl += len(l)

			if err != nil {
				if err.Error() == "body end reached. No separator met" {
					return lines, len(l), fmt.Errorf("in repo.GetLinesRightMiddle header \"%s\" is not full", l)
				}
			}
			if !bytes.Contains(b, []byte("filename=\"")) {
				return lines, c + len(l) + 4, nil
			}

		} else if i == 0 { //line is different from CD header line
			return make([][]byte, 0), 0, fmt.Errorf("in repo.GetLinesRightMiddle first line \"" + string(l) + "\" is unexpected")
		}
		if i == 1 &&
			IsCTRight(l) {

			lines = append(lines, l)

			if err != nil {
				if err.Error() == "body end reached. No separator met" {
					s := InOneLine(lines)
					return lines, cut + len(l), fmt.Errorf("in repo.GetLinesRightMiddle header \"%s\" is not full", s)
				}
			}

		} else if i == 1 { //line is different from Content-Type header line
			return make([][]byte, 0), 0, fmt.Errorf("in repo.GetLinesRightMiddle second line \"" + string(l) + "\" is unexpected")
		}

		if i == 0 &&
			(lenb == lenl+1 ||
				lenb == lenl+2) { // b ends between header lines
			return lines, cut + lenb - lenl - 2, fmt.Errorf("in repo.GetLinesRightMiddle header \"%s\" is not full", l)
		}
		i++
		b = b[lenl+2:]
		lenl += 2
		limit -= lenl
		cut += lenl
	}

	return lines, cut + 2, nil
}

// PartlyBoundaryLeft returns part of boundary found in the end of b.
// Tested in byteOps_test.go
func PartlyBoundaryLeft(b []byte, bou Boundary) ([]byte, error) {
	boundary := GenBoundary(bou)
	i := 0

	l, err := SingleLineLeftTrimmed(b, MaxLineLimit)
	if err != nil {
		return nil, err
	}

	for bytes.Contains(boundary, l[:i]) && i < len(b) {
		i++
	}
	if i <= len(boundary)+1 {
		return l, nil
	}

	return nil, nil
}

// PartlyBoundaryRight returns part of boundary found in the beginning of b.
// Tested in byteOps_test.go
func PartlyBoundaryRight(b []byte, limit int) ([]byte, error) {

	l, err := SingleLineRightUnchanged(b, limit)

	if err != nil {
		return nil, err
	}
	return l, nil
}

// LastBoundary returns true if boundary is the last
func LastBoundary(b, boundary []byte) bool {
	if len(b) > bytes.Index(b, boundary)+len(boundary) &&
		bytes.Contains(b, boundary) {
		return b[bytes.Index(b, boundary)+len(boundary)] != 13
	}
	return false

}

// GetLineRight return byte slice starting at fromIndex, ending at the CT.
// Returns error if limit exceeded
func GetLineRight(b []byte, fromIndex, limit int) ([]byte, error) {
	lenb, bb := len(b), make([]byte, 0)
	for i := fromIndex; i < Min(fromIndex+limit, len(b)); i++ {
		if b[i] == 13 {
			break
		}
		bb = append(bb, b[i])
		if i == fromIndex+limit-1 {
			return bb, errors.New("in repo.GetLineRight limit exceeded. No separator found")
		}
		if i == lenb-1 {
			return bb, errors.New("in repo.GetLineRight EOF reached. No separator found")
		}
	}
	return bb, nil

}

// GetLastLine returning last line with preceding CRLF and with following CR if it's met.
// If last bytes are CRLF, returning just them.
// Tested in byteOps_test.go
func GetLastLine(b, boundary []byte) []byte {
	lenb := len(b)

	switch lc := b[lenb-1]; {
	case lc == 10:
		if b[lenb-2] == 13 {
			return b[lenb-2:]
		}
	default:
		l := make([]byte, 0)
		npl, err := CurrentLineFirstPrintIndexLeft(b, len(boundary))
		if err != nil {
			return nil
		}

		line, err := SingleLineLeftTrimmed(b, len(boundary)+1)
		if err != nil {
			return nil
		}
		if lenb > len(line)+2 && // checking if there is separator before lastline
			b[npl-len(line)-1] == 13 {
			l = append(l, b[npl-len(line)-1], b[npl-len(line)]) // appending lastline and separator
			l = append(l, b[bytes.Index(b, line):]...)

			return l
		}
	}
	return nil
}

// BoundaryPartInLastLine returns last line of b if it is the beginning of boundary.
// Tested in byteOps_test.go
func BoundaryPartInLastLine(b []byte, bou Boundary) ([]byte, error) {
	i, lenb, line := 0, 0, make([]byte, 0)
	lenb = len(b)
	i = lenb - 1
	boundary := GenBoundary(bou)

	if i < 1 {
		return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary got 0-length last line")
	}
	if !bytes.Contains(boundary, b[lenb-1:]) {
		l, err := GetCurrentLineLeft(b, i, MaxLineLimit)
		if err != nil { //Test that
			return nil, err
		}
		if bytes.Contains(b[lenb-len(l)-2:lenb-len(l)], []byte("\r\n")) {
			ll := []byte("\r\n")
			ll = append(ll, l...)

			if len(ll) < len(boundary) {
				boundary = boundary[:len(ll)]
			}
			for j, v := range ll[:len(boundary)] {
				if boundary[j] != v {
					return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary")
				}
			}
			return append([]byte("\r\n"), l...), nil
		}
		return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary")
	}
	if lenb > len(boundary) &&
		b[lenb-1] == 13 &&
		bytes.Contains(boundary, b[lenb-len(boundary)-1:lenb-1]) {
		return b[(lenb - len(boundary) - 1):], nil
	}
	for i >= 0 && bytes.Contains(boundary, b[i:]) {
		line = b[i:]
		i--
	}
	for j, w := range line {
		if boundary[j] != w {
			return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary")
		}
	}

	return line, nil
}

// WordRightBorderLimit returns part of b between beg and end, or error if limit between beg and end exceeded.
// Tested in byteOps_test.go
func WordRightBorderLimit(b, beg, end []byte, limit int) ([]byte, error) {

	i := 0

	fromIndex := bytes.Index(b, beg)

	if fromIndex < 0 {
		return []byte(""), errors.New("beginning not found")
	}
	b = b[fromIndex+len(beg):]

	for i < limit {
		if i == bytes.Index(b, end) {
			return b[:i], nil
		}
		i++
	}
	return []byte(""), errors.New("limit exceeded")

}

// RepeatedIntex returns index of not first occurence of occ in byte slice.
// Tested in byteOps_test.go
func RepeatedIntex(b, occ []byte, i int) int {
	index, n := 0, 0

	for n < i {
		n++
		indexN := bytes.Index(b, occ)
		if n == 1 {
			index += indexN
		} else {
			index += indexN + len(occ)
		}
		cutted := indexN + len(occ)

		b = b[cutted:]

	}
	return index
}

// BeginningEqual returns true if first and second slices have equal characters in the same positions.
// Tested in byteOps_test.go
func BeginningEqual(s1, s2 []byte) bool {
	if len(s1) > len(s2) {
		s1 = s1[:len(s2)]
	} else {
		s2 = s2[:len(s1)]
	}
	for i, v := range s2 {
		if s1[i] != v {
			return false
		}
	}
	return true
}

// EndingOf returns true if first slice contains second and second slice is the ending of the first.
// Tested in byteOps_test.go
func EndingOf(long, short []byte) bool {
	longtLE, shortLE, lenLong, lenShort := byte(0), byte(0), len(long), len(short)
	if lenShort < 1 {
		return true
	}
	if lenLong < 1 {
		return false
	}
	shortLE = short[lenShort-1]
	longtLE = long[lenLong-1]
	if longtLE != shortLE {
		return false
	}
	for i := lenShort - 1; i > -1; i-- {
		if short[i] != long[lenLong-lenShort+i] {
			return false
		}
	}

	return true
}

// InOneLine returns concatenation of all lines
func InOneLine(lines [][]byte) string {
	sb := strings.Builder{}
	for i, v := range lines {
		sb.WriteString(string(v))
		if i < len(lines)-1 {
			sb.WriteString("\r\n")
		}
	}
	return sb.String()
}

// GetFoFi returns formname and filename found in b
func GetFoFi(b []byte) (string, string) {
	fo, fi, foPre, fiPre := "", "", []byte(" name=\""), []byte(" filename=\"")
	if len(b) > 0 {
		if bytes.Contains(b, foPre) {
			fo = string(b[bytes.Index(b, foPre)+len(foPre) : FindNext(b, []byte("\""), bytes.Index(b, foPre)+len(foPre))])
			if bytes.Contains(b, fiPre) {
				fi = string(b[bytes.Index(b, fiPre)+len(fiPre) : FindNext(b, []byte("\""), bytes.Index(b, fiPre)+len(fiPre))])
			}
		}
	}
	return fo, fi
}

// GetHeaderLines returns header lines found in b
// Tested in byteOps_test.go
func GetHeaderLines(b []byte, bou Boundary) ([]byte, error) {
	resL := make([]byte, 0)
	if len(b) == 0 {
		return resL, fmt.Errorf("in repo.GetHeaderLines zero len byte slice passed")
	}
	if b[0] == 10 { // preceding LF

		switch bytes.Count(b, []byte("\r\n")) {

		case 0: //  LF + rand
			resL = append(resL, b[0])
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)

		case 1: // LF + CRLF + rand
			resL = append(resL, b[0])
			resL = append(resL, []byte("\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)

		case 2: // LF + CT + 2*CRLF + rand || LF + CDSuff + 2*CRLF + rand
			l0 := b[1:bytes.Index(b, []byte("\r\n"))]
			resL = append(resL, b[0])
			resL = append(resL, l0...)
			resL = append(resL, []byte("\r\n\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)

		default: //  LF + CDinsuf + CRLF + CT + 2*CRLF + rand
			l0 := b[1:bytes.Index(b, []byte("\r\n"))]
			l1 := b[bytes.Index(b, []byte("\r\n"))+2 : RepeatedIntex(b, []byte("\r\n"), 2)]
			if Sufficiency(l0) == Insufficient {
				resL = append(b[:1], l0...)
				resL = append(resL, []byte("\r\n")...)
				if IsCTFull(l1) {
					resL = append(resL, l1...)
					resL = append(resL, []byte("\r\n\r\n")...)
					return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
				}
			}
			resL = append(resL, b[0])
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}

	}
	if b[len(b)-1] == 13 { // succeeding CR

		switch bytes.Count(b, []byte("\r\n")) {

		case 0: //  CD full + CR
			if Sufficiency(b[:len(b)-1]) != Incomplete {
				resL = append(resL, b...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
			}

		case 1: // CDsuf + CRLF + CR || CDinsuf + CRLF + CT + CR

			l0 := b[:bytes.Index(b, []byte("\r\n"))]

			if Sufficiency(l0) == Sufficient {
				resL = append(resL, b...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
			}

			if Sufficiency(l0) == Insufficient {

				resL = append(l0, []byte("\r\n")...)

				l1 := b[bytes.Index(b, []byte("\r\n"))+2 : len(b)-1]

				if IsCTFull(l1) {
					resL = append(resL, l1...)
					resL = append(resL, []byte("\r")...)

					return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
				}
			}

		case 2: // CDinsuf + CRLF + CT + CRLF + CR
			l0 := b[:bytes.Index(b, []byte("\r\n"))]
			l1 := b[bytes.Index(b, []byte("\r\n"))+2 : RepeatedIntex(b, []byte("\r\n"), 2)]
			if Sufficiency(l0) == Insufficient {
				resL = append(l0, []byte("\r\n")...)
				if IsCTFull(l1) {
					resL = append(resL, l1...)
					resL = append(resL, []byte("\r\n\r")...)
					return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
				}
			}

		default: // CDinsuf + CRLF + CT + 2*CRLF + rand + CR
			l0 := b[:bytes.Index(b, []byte("\r\n"))]
			l1 := b[bytes.Index(b, []byte("\r\n"))+2 : RepeatedIntex(b, []byte("\r\n"), 2)]
			if Sufficiency(l0) == Insufficient {
				resL = append(l0, []byte("\r\n")...)
				if IsCTFull(l1) {
					resL = append(resL, l1...)
					resL = append(resL, []byte("\r\n\r\n")...)
					return resL, nil
				}
			}

			return nil, fmt.Errorf("in repo.GetHeaderLines no header found")
		}
	}
	// no precending LF no succeding CR

	switch bytes.Count(b, []byte("\r\n")) {
	case 0: // CD ->

		if IsCDRight(b) {
			return b, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", b)
		}
		if IsLastBoundaryPart(b, bou) {
			return b, nil
		}

		return nil, fmt.Errorf("in repo.GetHeaderLines no header found")

	case 1: // CD full + CRLF || CD full + CRLF + CT -> || CRLF || <-LastBoundary + CRLF

		l0 := b[:bytes.Index(b, []byte("\r\n"))]
		l1 := b[bytes.Index(b, []byte("\r\n"))+2:]

		if len(l0) == 0 {
			resL = append(l0, []byte("\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}
		if Sufficiency(l0) == Sufficient {
			resL = append(l0, []byte("\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
		}
		if Sufficiency(l0) == Insufficient {
			resL = append(l0, []byte("\r\n")...)

			if IsCTRight(l1) {
				resL = append(resL, l1...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
			}

		}
		if len(b) == bytes.Index(b, []byte("\r\n"))+2 { //last boundary
			resL = append(resL, b...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is the last", resL)
		}

		return nil, fmt.Errorf("in repo.GetHeaderLines no header found")

	case 2: // CD full insufficient + CRLF + CT full + CRLF || CD full sufficient + 2 CRLF + rand || CT full + 2 CRLF + rand || <-CT + 2CRLF + rand || 2 CRLF + rand

		l0 := b[:bytes.Index(b, []byte("\r\n"))]
		l1 := b[bytes.Index(b, []byte("\r\n"))+2 : RepeatedIntex(b, []byte("\r\n"), 2)]

		if len(l0) == 0 { // on ending part is impossible, on beginning part: 2 * CRLF + rand || CRLF + rand + CRLF + rand
			resL = append(resL, []byte("\r\n")...)
			if len(l1) == 0 {
				resL = append(resL, []byte("\r\n")...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
			}
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}

		if Sufficiency(l0) == Sufficient { // on ending part CDSuf + 2 * CRLF || CDSuf + 2 * CRLF + rand, on beginning part CDSuf + 2 * CRLF + rand
			resL = append(l0, []byte("\r\n\r\n")...)
			return resL, nil
		}

		if Sufficiency(l0) == Insufficient { // on ending part CDInsuf + CRLF + CT + CRLF, on beginning part is impossible
			resL = append(l0, []byte("\r\n")...)
			if IsCTFull(l1) {
				resL = append(resL, l1...)
				resL = append(resL, []byte("\r\n")...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
			}
		}
		if IsCDLeft(l0) && len(l1) == 0 { // on ending part is impossible, on beginning part <-CDSufficient + 2 * CRLF + rand
			resL = append(l0, []byte("\r\n\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}

		if IsCTLeft(l0) && len(l1) == 0 { // on ending part is impossible, on beginning part <-CT + 2 * CRLF + rand
			resL = append(l0, []byte("\r\n\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}

		return nil, fmt.Errorf("in repo.GetHeaderLines no header found")

	default: // CD full insufficient + CRLF + CT full + 2*CRLF || CD full sufficient + 2*CRLF + rand + CRLF

		l0 := b[:bytes.Index(b, []byte("\r\n"))]
		l1 := b[RepeatedIntex(b, []byte("\r\n"), 1)+2 : RepeatedIntex(b, []byte("\r\n"), 2)]
		l2 := b[RepeatedIntex(b, []byte("\r\n"), 2)+2 : RepeatedIntex(b, []byte("\r\n"), 3)]

		if len(l0) >= 0 && EndingOf(GenBoundary(bou)[2:], l0) && (Sufficiency(l1) == Insufficient || Sufficiency(l1) == Sufficient) {

			resL = append(l0, []byte("\r\n")...)
			if Sufficiency(l1) == Insufficient {
				resL = append(resL, l1...)
				resL = append(resL, []byte("\r\n")...)
				if IsCTFull(l2) {
					resL = append(resL, l2...)
					resL = append(resL, []byte("\r\n\r\n")...)
					return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
				}
			}

		}
		if Sufficiency(l0) == Insufficient &&
			IsCTFull(l1) {
			resL = append(resL, l0...)
			resL = append(resL, []byte("\r\n")...)
			resL = append(resL, l1...)
			resL = append(resL, []byte("\r\n\r\n")...)

			return resL, nil
		}
		if len(l0) == 0 { // on ending part is impossible, on beginning part: CRLF + CDsuf + 2*CRLF + rand || CRLF + CDinsuf + CRLF + CT + 2*CRLF + rand || CRLF + CT + 2*CRLF + rand || CRLF + rand // 2*CRLF + rand
			resL = append(l0, []byte("\r\n")...)

			if len(l1) == 0 {
				resL = append(resL, []byte("\r\n")...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
			}

			if Sufficiency(l1) == Sufficient {
				resL = append(resL, l1...)
				resL = append(resL, []byte("\r\n\r\n")...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
			}
			if Sufficiency(l1) == Insufficient {
				resL = append(resL, l1...)
				resL = append(resL, []byte("\r\n")...)
				if IsCTFull(l2) {
					resL = append(resL, l2...)
					resL = append(resL, []byte("\r\n\r\n")...)
					return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
				}
			}
			if IsCTFull(l1) {
				resL = append(resL, l1...)
				resL = append(resL, []byte("\r\n\r\n")...)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
			}
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}
		if len(l1) == 0 { // on ending part CDsuf + 2*CRLF + rand, on beginning part <-CDsuf + 2*CRLF + rand || <-CT + 2 * CRLF + rand
			resL = append(l0, []byte("\r\n\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}
		if len(l2) == 0 { // on ending part CDinsuf + CRLF + CT + 2*CRLF + rand, on beginning part CRLF + CDsuf + 2*CRLF = rand || <-Bound + CRLF + CDsuf + 2*CRLF = rand || <-CDinsuf + CRLF + CT + 2*CRLF
			resL = append(l0, []byte("\r\n")...)
			resL = append(resL, l1...)
			resL = append(resL, []byte("\r\n\r\n")...)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}
		return nil, fmt.Errorf("in repo.GetHeaderLines no header found")

	}
}

// Returns known part of last boundary.
// Tested in byteOps_test.go
func KnownBoundaryPart(b []byte, bou Boundary) []byte {
	res := make([]byte, 0)
	realBoundary := GenBoundary(bou)[2:]

	for i := 0; i < len(b); i++ {
		//logger.L.Infof("in repo.KnownBoundaryPart i = %d, b[i]: %q\n", i, b[i])
		if !bytes.Contains(realBoundary, append(res, b[i])) {
			return res
		}
		res = append(res, b[i])
		//logger.L.Infof("in repo.KnownBoundaryPart res: %q\n", res)

	}
	return res
}

// IsBoundary returns true if p + n == boundary.
// Tested in byteOps_test.go
func IsBoundary(p, n []byte, bou Boundary) bool {
	cd := []byte("Content-Disposition")

	if !bytes.Contains(n, cd) {
		return false
	}

	nBou := n[:bytes.Index(n, cd)]

	if len(nBou) < 1 {
		return false
	}

	boundary := append(GenBoundary(bou)[2:], []byte("\r\n")...)

	bs := append(p, nBou...)
	return bytes.Contains(bs, boundary)
}

// IsLastBoundary returns true if p + n form last boundary
func IsLastBoundary(p, n []byte, bou Boundary) bool {
	realBoundary := GenBoundary(bou)
	combined := append(p, n...)
	if len(combined) > len(realBoundary) &&
		bytes.Contains(combined, realBoundary) &&
		!bytes.Contains(combined[len(realBoundary):len(realBoundary)+2], []byte("\r\n")) {
		return true
	}

	return false
}

// GetLineWithCRLFLeft returns CRLF and succeeding line before given index.
// If line ends with CR (or CRLF) and contains boundary, returns CRLF + line + CR (CRLF).
// Tested in byteOps_test.go
func GetLineWithCRLFLeft(b []byte, fromIndex, limit int, bou Boundary) []byte {

	l, lenb, c, n := make([]byte, 0), len(b), 0, 0

	if lenb < 1 {
		return l
	}
	if fromIndex > lenb-1 {
		fromIndex = lenb - 1
	}

	if fromIndex < limit {
		c = 0
	} else {
		c = fromIndex - limit
	}
	for i := fromIndex; i > c; i-- {
		if n == 0 &&
			(i == lenb-1 && b[i] == 13 ||
				i == lenb-2 && b[i] == 13 && b[i+1] == 10) &&
			(i >= 14 && ContainsBouEnding(b[i-14:i], bou)) {
			n++
			continue
		} else if i == lenb-1 && b[i] == 13 ||
			i == lenb-2 && b[i] == 13 && b[i+1] == 10 {

			return b[i:]
		}
		if b[i] == 13 && b[i+1] == 10 {
			return b[i:]
		}
	}
	return b
}

// ContainsBouEnding returns true if b contains boundary ending.
// Tested in byteOps_test.go
func ContainsBouEnding(b []byte, bou Boundary) bool {
	n, boundary := 0, GenBoundary(bou)
	for i := 0; i < len(b); i++ {
		if !bytes.Contains(boundary, b[:i]) && n > 4 {
			return true
		}
		if !bytes.Contains(boundary, b[:i]) {
			return false

		}
		n++
	}
	return true
}

// IsLastBoundaryPart returns true if b is ending part of last boundary.
// Tested in byteOps_test.go
func IsLastBoundaryPart(b []byte, bou Boundary) bool {
	lenb, suffix := len(b), make([]byte, 0)
	i, lastSymbol := lenb, b[lenb-1]

	for i >= 1 {
		if i == 1 {
			return true
		}
		if i > 1 && b[i-1] != lastSymbol {
			break
		}
		i--
	}
	suffix = b[i:]
	rootLen := lenb - len(suffix)

	if rootLen < lenb && bytes.Contains(GenBoundary(bou), b[:rootLen]) {
		return true
	}
	return false
}
