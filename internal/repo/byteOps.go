package repo

import (
	"bytes"
	"errors"
	"fmt"
	"postParser/internal/logger"
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

func FindNext(b, occ []byte, fromIndex int) int {
	return bytes.Index(b[fromIndex:], occ) + fromIndex
}

func FindBoundary(b []byte) Boundary {

	bPrefix, bRoot := []byte{}, []byte{}

	if bytes.Contains(b, []byte(BoundaryField)) {

		start := bytes.Index(b, []byte(BoundaryField)) + len(BoundaryField)

		bRoot = LineRightLimit(b, start, 70)

		//	logger.L.Infof("in repo.FindBoundary bRoot = %q\n", bRoot)

		bb := b[start+1:]

		secBoundaryIndex := bytes.Index(bb, bRoot) - 1

		bPrefix = LineLeftLimit(bb, secBoundaryIndex, MaxLineLimit)

		//	logger.L.Infof("in repo.FindBoundary bPrefix = %q\n", bPrefix)
	}

	return Boundary{
		Prefix: bPrefix,
		Root:   bRoot,
	}

}

// Searches byte slice to the right with limit. Returning last index before CR
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

// Returns previous part of line from passed index till LF or zero symbol to the left direction. FromIndex<=len(b)-1
func LineLeftLimit(b []byte, fromIndex, limit int) []byte {
	bb, i := make([]byte, 0), fromIndex

	if fromIndex <= 0 {
		return bb
	}
	//logger.L.Infof("in LineLeftLimit fromIndex-limit = %d,Max(fromIndex-limit, 0) = %d\n", fromIndex-limit, Max(fromIndex-limit, 0))
	/*	for i := fromIndex; b[i] != 10 &&
			i >= Max(fromIndex-limit, 0); i-- {
			bb = append(bb, b[i])
			logger.L.Infof("in repo.LineLeftLimit i = %d bb: %q\n", i, bb)
		}
	*/
	for b[i] != 10 &&
		i >= Max(fromIndex-limit, 0) {
		bb = append(bb, b[i])
		//logger.L.Infof("in repo.LineLeftLimit i = %d bb: %q\n", i, bb)

		if i == 0 {
			break
		}
		i--
	}
	if len(bb) == limit {
		return make([]byte, 0)
	}
	bbr := Reverse(bb)

	//logger.L.Infof("in repo.LineLeftLimit bbr:%q\n", bbr)

	return bbr
}

// Returns rest of line from passed index till СR or EOF to the right direction
func LineRightLimit(b []byte, fromIndex, limit int) []byte {
	bb := make([]byte, 0)

	if fromIndex < 0 {
		return nil
	}

	for i := fromIndex; b[i] != 13 && i < fromIndex+limit; i++ {
		bb = append(bb, b[i])
	}

	//logger.L.Infof("in repo.LineRightLimit len = %d, limit = %d\n", len(bb), limit)
	if len(bb) == limit {
		return nil
	}

	return bb
}

// Searches for first printable character index in first line of byte slice with limit. Returns error if limit exceeded.
func CurrentLineFirstPrintIndexRight(b []byte, limit int) (int, error) {
	//logger.L.Infof("repo.CurrentLineFirstPrintIndexRight invoked with b: %q, limit = %d\n", b, limit)

	p := 0 // statring number of printable character is -1 in order of correct work of incremention
	if len(b) == 0 {
		return -1, errors.New("passed byte slice with zero length")
	}

	for i := 0; (b[i] == 13 || b[i] == 10) && i < limit && i < len(b)-1; i++ {
		//logger.L.Infof("in repo.CurrentLineFirstPrintIndexRight i = %d, b[i] = %d\n", i, b[i])
		p++
	}
	if p == limit {
		return -1, errors.New("no actual characters before limit")
	}

	return p, nil
}

// Returns character byte slice to the right of passed index before CR with limit.
func GetCurrentLineRight(b []byte, fromIndex, limit int) ([]byte, error) {
	bb, i, lenb := make([]byte, 0), fromIndex, len(b)
	//logger.L.Infof("repo.GetCurrentLineRight invoked with parameters b: %q with length = %d, fromindex = %d, limit = %d\n", b, lenb, fromIndex, limit)

	if lenb == 0 {
		return bb, errors.New("passed byte slice with zero length")
	}
	//logger.L.Infof("in GetCurrentLineRight LineRightEndIndexLimit(b, fromIndex, limit) = %d\n", LineRightEndIndexLimit(b, fromIndex, limit))
	// checking if line is ending part of last boundary
	if lenb < limit && LineRightEndIndexLimit(b, fromIndex, limit) < 0 {

		//logger.L.Infof("in GetCurrentLineRight get into case with i = %d, lenb = %d\n", i, lenb)

		for i < lenb {
			bb = append(bb, b[i])
			logger.L.Infof("in GetCurrentLineRight in first loop bb: %q\n", bb)

			i++
		}
		return bb, nil
	}

	for i < fromIndex+limit && b[i] != 13 {

		bb = append(bb, b[i])
		//logger.L.Infof("in GetCurrentLineRight in second loop bb: %q\n", bb)

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
func SingleLineRightUnchanged(b []byte, limit int) ([]byte, error) {

	return GetCurrentLineRight(b, 0, limit)
}

func SingleLineRightTrimmed(b []byte, limit int) (int, []byte, error) {
	trimmed := 0
	index, err := CurrentLineFirstPrintIndexRight(b, limit)
	if err != nil {
		return trimmed, nil, err
	}
	//logger.L.Infof("in repo.SingleLineRightTrimmed first printed index %d\n", index)
	trimmed += index
	line, err := GetCurrentLineRight(b, index, limit-index)
	return trimmed, line, err
}

// Returns index of first printable character index to the left of byte slice end with limit, or error if limit exceeded
func CurrentLineFirstPrintIndexLeft(b []byte, limit int) (int, error) {
	//logger.L.Infof("repo.CurrentLineFirstPrintIndexLeft invoked with b: %q, limit = %d\n", b, limit)
	lenb := len(b)
	i := lenb - 1
	//logger.L.Infof("in repo.CurrentLineFirstPrintIndexLeft before loop b: %q, lenb = %d, i = %d\n", b, lenb, i)
	if lenb == 0 {
		return -1, errors.New("in repo.CurrentLineFirstPrintIndexLeft passed byte slice with zero length")
	}

	//logger.L.Infof("in repo.CurrentLineFirstPrintIndexLeft before loop i = %d b[i] = %d\n", i, b[i])

	for (b[i] == 13 || b[i] == 10) &&
		i > lenb-1-limit &&
		i > 0 {
		//logger.L.Infof("in repo.CurrentLineFirstPrintIndexLeft in loop i = %d b[i] = %d\n", i, b[i])
		i--
	}
	if i == lenb-1-limit {
		return -1, errors.New("in repo.CurrentLineFirstPrintIndexLeft no actual characters before limit")
	}

	return i, nil
}

// Checks byte slice to the left, returns line before separator or error if limit exceeded
func GetCurrentLineLeft(b []byte, fromIndex, limit int) ([]byte, error) {
	//logger.L.Infof("repo.GetCurrentLineLeft invoked with b = %q, fromindex = %d limit = %d\n", b, fromIndex, limit)
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

// Checks byte slice to the left, trims it if necessary, returns line before separator or error if limit exceeded
func SingleLineLeftTrimmed(b []byte, limit int) ([]byte, error) {
	//logger.L.Infof("repo.SingleLineLeftTrimmed invoked with b = %q, limit = %d\n", b, limit)
	index, err := CurrentLineFirstPrintIndexLeft(b, limit)
	if err != nil {
		return nil, err
	}
	//logger.L.Infof("in repo.SingleLineLeftTrimmed in lenb = %d printIndex = %d\n", len(b), index)
	return GetCurrentLineLeft(b, index, limit)
}

func Reverse(bb []byte) []byte {

	bbs := make([]byte, 0)

	for i := len(bb) - 1; i >= 0; i-- {
		bbs = append(bbs, bb[i])
	}

	return bbs
}

// Slices whole page into byte chunks based on boundaries met
func Slicer(b []byte, bou Boundary) (AppPieceUnit, []AppPieceUnit, AppSub) {
	//	logger.L.Infof("in repo.Slicer body: %q and bounddary = %q\n", b, boundary)
	var (
		m []AppPieceUnit
	//	bp  []byte
	//	err error
	)
	boundaryCore := GenBoundary(bou)[2:]
	//boundary = append(boundary, []byte("\r\n")...)
	boundaryMiddle := append([]byte("\r\n"), boundaryCore...)
	boundaryMiddle = append(boundaryMiddle, []byte("\r\n")...)
	//boundaryFull := append(boundary, bou.Suffix...)
	//_ := append(boundary, bou.Suffix...)
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
			//logger.L.Infof("in repo.Slicer apub header %v, body: %q\n", aphb, apbb.B)
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
		//logger.L.Infof("in repo.Slicer apue header %v, body: %q\n", aphe, apbe)
		be := make([]byte, 0)
		if len(b) > MaxHeaderLimit {
			be = b[len(b)-MaxLineLimit:]
		} else {
			be = b
		}
		//logger.L.Infof("in repo.Slicer be: %q\n", be)
		lenbe := len(be)
		ll := GetLineWithCRLFLeft(be, lenbe-1, MaxLineLimit, bou)

		if len(ll) > 2 && BeginningEqual(ll[2:], boundaryCore) { // last line equal to boundary begginning or vice versa

			apbe := NewAppPieceBodyFilled(b[:len(b)-len(ll)])

			if IsLastBoundary(ll, []byte(""), bou) { // last boundary in last line

				aphe.SetE(False)
				apume := NewAppPieceUnit(aphe, apbe)
				//logger.L.Infof("in repo.Slicer apume header: %v, body: %q\n", apume.APH, apume.APB)

				m = append(m, apume)

				return apub, m, AppSub{}
			}

			aphe.SetE(Probably)

			apue := NewAppPieceUnit(aphe, apbe)
			//logger.L.Infof("in repo.Slicer apue header %v, body: %q\n", aphe, apbe.B)
			m = append(m, apue)

			as := NewAppSub()
			as.SetBody(ll)
			/*
				if apub.TS() == "" {
					return AppPieceUnit{}, m, as
				}
			*/
			return apub, m, as
		}
		if (len(ll) == 1 && bytes.Contains(ll, []byte("\r"))) || // last line is CR
			(len(ll) == 2 && bytes.Contains(ll, []byte("\r\n"))) { // last line is LF
			aphe.SetE(Probably)
			apbe.SetBody(b[:len(b)-len(ll)])
			apue := NewAppPieceUnit(aphe, apbe)
			//logger.L.Infof("in repo.Slicer apue header %v, body: %q\n", apue.APH, apue.APB)

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
			//logger.L.Infoln("in repo.Slicer ll is boundary beginning")
			//apbb := NewAppPieceBodyFilled(b[:len(b)-len(ll)])
			if len(ll) == len(b) {
				apbb.SetBody(b)
			} else {
				apbb.SetBody(b[:len(b)-len(ll)])
			}
			if IsLastBoundary(ll, []byte(""), bou) { // last boundary in last line
				aphb.SetE(False)
				apub := NewAppPieceUnit(aphb, apbb)
				//logger.L.Infof("in repo.Slicer last line apub header %v, body: %q\n", apub.APH, apub.APB)

				return apub, nil, AppSub{}
			}

			//func GetLineWithCRLFLeft(b []byte, fromIndex, limit int, bou Boundary) []byte

			aphb.SetE(Probably)

			apub := NewAppPieceUnit(aphb, apbb)
			//logger.L.Infof("in repo.Slicer case 0 full boundary apub header %v, body: %q\n", apub.APH, apub.APB)
			as := NewAppSub()
			as.SetBody(ll)

			return apub, nil, as
		}

		if (len(ll) == 1 && bytes.Contains(ll, []byte("\r"))) || // last line is CR
			(len(ll) == 2 && bytes.Contains(ll, []byte("\r\n"))) { // last line is LF
			aphb.SetE(Probably)
			apbb.SetBody(b[:len(b)-len(ll)])
			apub := NewAppPieceUnit(aphb, apbb)
			//logger.L.Infof("in repo.Slicer apub header %v, body: %q\n", apub.APH, apub.APB)
			as := NewAppSub()
			as.SetBody(ll)

			return apub, nil, as

		}
		aphb.SetE(True)
		apbb.SetBody(b)

		apub := NewAppPieceUnit(aphb, apbb)
		//logger.L.Infof("in repo.Slicer no last line apub header %v, body: %q\n", apub.APH, apub.APB)
		return apub, nil, AppSub{}
	}
}

//logger.L.Infof("in repo.Slicer body: %q and boundNum = %d\n", b, boundNum)
/*
	switch boundNum {
	case 0:
		//logger.L.Infof("in repo.Slicer boundary: %q, len = %d\n", boundary, len(boundary))
		//logger.L.Infof("in repo.Slicer b: %q, len = %d\n", b, lenb)
		aphb := NewAppPieceHeader()
		apbb := NewAppPieceBodyEmpty()
		aphb.SetB(true)

		be := b[lenb-MaxLineLimit:]
		lenbe := len(be)
		ll := GetLineWithCRLFLeft(be, lenbe-1, MaxLineLimit)

		if BeginningEqual(ll, boundary) { // last line equal to boundary begginning or vice versa

			apbb.SetBody(b[:lenb-len(ll)])
			aphb.SetE(Probably)
			apub := NewAppPieceUnit(aphb, apbb)

			as := NewAppSub()
			as.SetBody(ll)

			return apub, nil, as
		}

		apbb.SetBody(b)
		aphb.SetE(True)
		apub := NewAppPieceUnit(aphb, apbb)

		return apub, nil, AppSub{}

	default:

		//logger.L.Infof("in repo.Slicer body: %q and boundNum = %d\n", b, boundNum)

		bi := bytes.Index(b, boundary)
		//logger.L.Infof("in repo.Slicer boundary = %q, bi = %d\n", boundary, bi)

		apub := NewAppPieceUnitEmpty()

		m := make([]AppPieceUnit, 0)

		aphb := NewAppPieceHeader()
		aphb.SetB(true)

		apbb := NewAppPieceBodyEmpty()

		//logger.L.Infof("in repo.Slicer b piece body: %q\n", apbb.B)
		//logger.L.Infof("in repo.Slicer LastBoundary in begining piece %t\n", LastBoundary(b, boundary))
		be := b[lenb-MaxLineLimit:]
		lenbe := len(be)
		ll := GetLineWithCRLFLeft(be, lenbe-1, MaxLineLimit)

		if BeginningEqual(ll, boundary) { // last line equal to boundary begginning or vice versa

			apbb.SetBody(b[:lenb-len(ll)])
			aphb.SetE(Probably)
			apub := NewAppPieceUnit(aphb, apbb)

			as := NewAppSub()
			as.SetBody(ll)

			return apub, nil, as

		// partial boundary is not detected in b
		aphb.SetE(False)
		apbb.SetBody(b[:bi])

		if !bytes.Contains(b, []byte(BoundaryField)) {
			apub.SetAPH(aphb)
			apub.SetAPB(apbb)

		}
		//logger.L.Infof("in repo.Slicer before loop apub header: %v, body : %q\n", apub.APH, apub.APB.B)

		if lenb > bi+len(boundary)+2 {
			b = b[bi+len(boundary)+2:]
			lenb = len(b)
		}
		boundNum--
		//logger.L.Infof("in repo.Slicer before loop b: %q, boundnum = %d\n", b, boundNum)

		for i := 0; i < boundNum; i++ {
			//logger.L.Infof("in repo.Slicer got into loop\n")

			ni := bytes.Index(b[1:], boundary) + 1

			aph := NewAppPieceHeader()
			aph.SetB(false)

			apb := NewAppPieceBodyFilled(b[:ni])
			//logger.L.Infof("in repo.Slicer in loop %d apb: %q\n", i, apb.B)
			if ni > 0 {
				aph.SetE(False)
			}
			//logger.L.Infof("in repo.Slicer in loop %d aph: %v\n", i, aph)
			apu := NewAppPieceUnit(aph, apb)

			m = append(m, apu)

			//logger.L.Infof("in repo.Slicer LastBoundary(b, boundary) = %t/n", LastBoundary(b, boundary))

			b = b[ni+len(boundary)+2:]
			lenb = len(b)
		}

		//logger.L.Infof("in repo.Slicer got out of loop\n")
		aphe := NewAppPieceHeader()
		apbe := NewAppPieceBodyEmpty()
		aphe.SetB(false)

		if LastBoundary(b, boundary) { //last boundary detected in b
			//logger.L.Infof("in repo.Slicer got to last boundary handling for b: %q and boundary %q\n", b, boundary)

			if bytes.Contains(b[lenb-2:], []byte("\r\n")) { // last boundary is on current b fully
				//logger.L.Infof("in repo.Slicer got to last boundary handling for b: %q and boundary %q fully present\n", b, boundary)
				aphe.SetE(False)
				apbe.SetBody(b[:bytes.Index(b, boundary)])
				apume := NewAppPieceUnit(aphe, apbe)
				m = append(m, apume)
				return apub, m, AppSub{}
			}

			aphe.SetE(Probably)
			apbe.SetBody(b[:bytes.Index(b, boundary)])
			apume := NewAppPieceUnit(aphe, apbe)
			m = append(m, apume)
			as := NewAppSub()
			as.SetBody(b[bytes.Index(b, boundary)+2:])

			return apub, m, as

		}
		// no last boundary detected

		if lenb < len(boundary) { // checking for partial boundary in b
			bp, err = BoundaryPartInLastLine(b, bou)
		} else {
			bp, err = BoundaryPartInLastLine(b[lenb-1-len(boundary):], bou)
		}
		if err != nil { //no partial boundary detected in b

			//logger.L.Infof("in repo.Slicer last boundary handling err branch with b : %q\n",b)

			aphe.SetE(True)
			apbe.SetBody(b)
			apume := NewAppPieceUnit(aphe, apbe)
			m = append(m, apume)

			return apub, m, AppSub{}
		}

		// partial boundary detected in b
		//logger.L.Infof("in repo.Slicer got partial boundary %q\n", bp)

		aphe.SetE(Probably)
		apbe.SetBody(b[:bytes.Index(b, bp)])
		m = append(m, NewAppPieceUnit(aphe, apbe))

		as := NewAppSub()

		switch lenbp := len(bp); {
		case lenbp < 3:
			as.SetBody([]byte(""))
		default:
			as.SetBody(bp[2:])
		}

		return apub, m, as

		//logger.L.Infof("in repo.Slicer before computing lastLine: %q\n", b)

		//logger.L.Infof("in repo.Slicer lastLine: %q\n", lastLine)

	}
*/

func GetBoundaryFirstPart(b, boundary []byte) int {
	lenb := len(b)
	last := b[lenb-1]
	occ := 0

	//finding char in boundary
	for i := 0; i < len(boundary); i++ {
		if boundary[i] == last {
			occ++
			break
		}

	}
	if occ == 0 {
		return -1
	}
	boundaryStartPos := LineLeftPosLimit(b, lenb-1, len(boundary))
	boundaryLen := 0
	if boundaryStartPos < 0 {
		return boundaryLen
	}
	boundaryLen = lenb - boundaryStartPos

	occ = 0

	for i := lenb - 1; i >= boundaryStartPos && i-lenb+1+boundaryLen > 0; i-- {
		if boundary[i-lenb+boundaryLen] != b[i] {
			return -1
		}
		occ++
	}

	return occ

}

func IsPartlyBoundaryRight(b []byte, bou Boundary) bool {
	boundary := GenBoundary(bou)
	if len(b) > len(boundary) {
		return false
	}
	for i := len(b); i > 0; i-- {
		//logger.L.Infof("in repo.IsPartlyBoundaryRight i = %d, b[i-1] = %q,boundary[len(boundary)-1-len(b)+i] = %q\n", i, b[i-1], boundary[len(boundary)-1-len(b)+i])
		if len(boundary) > i &&
			b[i-1] != boundary[len(boundary)-1-len(b)+i] {
			return false
		}
	}
	return true
}

// true if last or only line of byte slice is the beginning of boundary
/*
func IsPartlyBoundaryLeft(b []byte, bou Boundary) bool {
	boundary := GenBoundary(bou)
	if len(b) > len(boundary) &&
		len(LineRightLimit(b, len(boundary), MaxLineLimit)) > 0 {

			return false
	}



	return true
}
*/

// returrns true if byte slice has all printable charackters/
func IsPrintable(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if !unicode.IsPrint(rune(b[i])) {
			return false
		}
	}
	return true
}

// generates boundary based on given Boundary struct
func GenBoundary(bou Boundary) []byte {
	boundary := make([]byte, 0)
	boundary = append(boundary, []byte("\r\n")...)
	boundary = append(boundary, bou.Prefix...)
	boundary = append(boundary, bou.Root...)
	return boundary
}

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

// True if all character in byte slice are printable
func AllPrintalbe(b []byte) bool {
	for _, v := range b {
		if !unicode.IsPrint(rune(v)) {
			return false
		}
	}
	return true
}

// True if byte slice have no digits
func NoDigits(b []byte) bool {
	for _, v := range b {
		if unicode.IsDigit(rune(v)) {
			return false
		}
	}
	return true
}
func CheckLineLimit(b []byte, limit int) bool {
	return false
}

func SliceParser(p AppPieceUnit, bou Boundary) {

	boundary := GenBoundary(bou)
	bNum := 0

	for i, v := range p.APB.B {
		if boundary[i] != v {
			logger.L.Infof("in repo.SliceParser piece body %q is not a part of boundary %q\n", p.APB.B, boundary)
			bNum++
			break
		}
	}

	if bNum == 0 {
		logger.L.Infof("in repo.SliceParser piece body %q is part of boundary %q\n", p.APB.B, boundary)
	}
}

// ToDo
// returns header lines on current data chunk which are hyphennated to the next data chunk
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
		//logger.L.Infof("in repo.GetLinesLeft in iteration %d got line %q with length %d\n", i, l, lenl)

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

			logger.L.Infof("in repo.GetLinesLeft in iteration %d lines %q\n", i, lines)

			break
		} else if i == 2 {
			lines = make([][]byte, 0)
		}
		b = b[:lenb-lenl-2]
		limit -= lenl

	}

	return lines, nil

}

/*
func GetLinesRight(b []byte, limit int, voc Vocabulaty) ([][]byte, error) {
	lines, i := make([][]byte, 0), 0

	r1, err := regexp.Compile("Content-Disposition:\\sform-data;\\sname=\"\\w+\"")
	if err != nil {
		return lines, errors.New("in repo.GetLinesLeft error while compiling regexp")
	}
	if len(b) == 0 {
		return lines, errors.New("in repo.GetLinesLeft zero length byte slice passed")
	}
	if limit > len(b) {
		limit = len(b)
	}

	for i < 3 && limit > 0 {
		c,l, err := SingleLineRightTrimmed(b[:limit], limit)
		if err != nil {
			return lines, err
		}
		//	logger.L.Infof("in repo.GetLinesRight in loop iteration %d line: %q\n", i, l)

		if i == 0 &&
			(IsPartlyBoundaryRight(l, voc.Boundary) || //check if line is last part of boundary
				(len(l) > 0 && l[len(l)-1] == 34 && AllPrintalbe(l)) || //check if line ends with " and has printable characters only
				(len(l) > 0 && l[len(l)-1] != 34 && NoDigits(l))) { //check if line doesn't end with " and has no digits

			lines = append(lines, l)
			//	logger.L.Infof("in repo.GetLinesRight in loop iteration %d lines: %q\n", i, lines)
			if bytes.Contains(l, []byte(voc.CType)) {
				//	logger.L.Infof("in repo.GetLinesRight inside 0 loop iteration got CType, leaving")
				break
			}
		} else if i == 0 {
			break
		}
		if i == 1 &&
			(r1.Match(l) || //check if line matches regexp
				bytes.Contains(l, []byte(voc.CType))) { //check if line contains Conternt-Type word
			lines = append(lines, l)

			if bytes.Contains(l, []byte(voc.CType)) {
				//	logger.L.Infof("in repo.GetLinesRight inside 1 loop iteration got CType, leaving")
				break
			}
		} else if i == 1 {
			break
		}

		if i == 2 &&
			bytes.Contains(l, []byte(voc.CType)) { //check if line contains Conternt-Type word

			lines = append(lines, l)
			break
		}

		i++
		limit -= len(l) + 2
		if len(b) < len(l)+2 || limit < 0 {
			break
		}

		b = b[len(l)+2:]

	}

	return lines, nil
}
*/
// Returns header lines on current data chunk which are hyphennated from previous data chunk
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
			//logger.L.Errorf("in repo.GetLinesRightBegin c = %d, l : %q, err = %v, limit = %d\n", c, l, err, limit)
			return lines, cut, err
		}

		//logger.L.Infof("in repo.GetLinesRightBegin in loop iteration %d line: %q EndingOfBoundary? %t, IsCDLeft? %t, IsCTLeft? %t, cut = %d\n", i, l, EndingOf(GenBoundary(bou), l), IsCDLeft(l), IsCTLeft(l), cut)

		if i == 0 &&
			(EndingOf(GenBoundary(bou), l) || //check if line is last part of boundary
				IsCDLeft(l) || //check if line is Content-Disposityon header line, cut from left
				IsCTLeft(l)) { //check if line is Content-Type header line, cut from left

			lines = append(lines, l)
			cut += c
			//logger.L.Infof("in repo.GetLinesRight in loop iteration %d lines: %q\n", i, lines)
			if IsCTLeft(l) {

				cut += len(l) + 2
				//logger.L.Infof("in repo.GetLinesRight in loop iteration %d CT met, breaking with cut %d\n", i, cut)
				return lines, cut, nil
			}
		} else if i == 0 {
			return make([][]byte, 0), 0, errors.New("first line \"" + string(l) + "\" is unexpected")
		}
		if i == 1 &&
			(IsCDLeft(l) || ////check if line is Content-Disposityon header line
				IsCTLeft(l)) { //check if line is Content-Type header line

			//logger.L.Infof("in repo.GetLinesRight in loop iteration %d l: %q, r1.Match(l)? %t, EndingOf(GenBoundary(voc.Boundary), lines[0])?%t\n", i, l, r1.Match(l), EndingOf(GenBoundary(bou), lines[0]))

			if r1.Match(l) &&
				(!EndingOf(GenBoundary(bou), lines[0]) ||
					IsCTLeft(l)) && !r2.Match(lines[0]) && !EndingOf(GenBoundary(bou), l) { //checking previous line REFACTOR THIS USING CDLeft
				return make([][]byte, 0), 0, errors.New("first line \"" + string(lines[0]) + "\" is unexpected")
			}

			lines = append(lines, l)

			//logger.L.Infof("in repo.GetLinesRight in loop iteration %d lines: %q\n", i, lines)

			if IsCTLeft(l) {
				//logger.L.Infof("in repo.GetLinesRight inside 1 loop iteration got CType, leaving")
				break
			}
		} else if i == 1 {
			return make([][]byte, 0), 0, errors.New("second line \"" + string(l) + "\" is unexpected")
		}
		//logger.L.Infof("in repo.GetLinesRight loop iteration %d done\n", i)

		if i == 2 &&
			IsCTLeft(l) { //check if line contains Conternt-Type word

			lines = append(lines, l)

			cut += len(l) + 2

			//logger.L.Infof("in repo.GetLinesRight in loop iteration %d lines: %q\n", i, lines)

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
		//logger.L.Infof("in repo.GetLinesRight loop iteration %d in the end b became %q\n", i, b)
	}

	return lines, cut, nil
}

// Returns header lines on data piece
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
		logger.L.Infof("in repo.GetLinesRightMiddle in loop iteration %d line: %q IsCDRight? %t, IsCTRight? %t\n", i, l, IsCDRight(l), IsCTRight(l))

		if i == 0 &&
			IsCDRight(l) {

			lines = append(lines, l)
			cut += c
			lenl += len(l)
			//logger.L.Infof("in repo.GetLinesMiddle in loop iteration %d lines: %q\n", i, lines)

			if err != nil {
				if err.Error() == "body end reached. No separator met" {
					return lines, len(l), fmt.Errorf("in repo.GetLinesRightMiddle header \"%s\" is not full", l)
				}
			}
			if !bytes.Contains(b, []byte("filename=\"")) {
				return lines, c + len(l) + 4, nil
			}
			//errors.New("in GetLinesRightMiddle header \"" + string(l) + "\" is not full")
			//	logger.L.Infof("in repo.GetLinesRight in loop iteration %d lines: %q\n", i, lines)

		} else if i == 0 { //line is different from CD header line
			return make([][]byte, 0), 0, fmt.Errorf("in repo.GetLinesRightMiddle first line \"" + string(l) + "\" is unexpected")
		}
		if i == 1 &&
			IsCTRight(l) {

			lines = append(lines, l)

			//logger.L.Infof("in repo.GetLinesRightMiddle in loop iteration %d lines: %q\n", i, lines)

			if err != nil {
				if err.Error() == "body end reached. No separator met" {
					s := InOneLine(lines)
					return lines, cut + len(l), fmt.Errorf("in repo.GetLinesRightMiddle header \"%s\" is not full", s)
					//errors.New("in GetLinesRightMiddle header \"" + s + "\" is not full")
				}
			}

		} else if i == 1 { //line is different from Content-Type header line
			return make([][]byte, 0), 0, fmt.Errorf("in repo.GetLinesRightMiddle second line \"" + string(l) + "\" is unexpected")
		}

		if i == 0 &&
			(lenb == lenl+1 ||
				lenb == lenl+2) { // b ends between header lines
			logger.L.Infof("in repo.GetLinesRight in loop iteration %d b is all in l: %q\n", i, l)
			return lines, cut + lenb - lenl - 2, fmt.Errorf("in repo.GetLinesRightMiddle header \"%s\" is not full", l)
		}
		i++
		b = b[lenl+2:]
		lenl += 2
		logger.L.Infof("in repo.GetLinesRightMiddle in loop iteration %d lenl = %d, len(b) = %d\n", i, len(l), len(b))
		limit -= lenl
		cut += lenl
	}

	return lines, cut + 2, nil
}

// returns part of boundary on the end of current data chunk
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

// returns part of boundary which was hyphetinated from previous data chunk
func PartlyBoundaryRight(b []byte, limit int) ([]byte, error) {

	l, err := SingleLineRightUnchanged(b, limit)

	if err != nil {
		return nil, err
	}
	return l, nil
}

/*
// returns lines of data chank header

	func GetLinesMiddle(b []byte, limit int, voc Vocabulaty) ([][]byte, error) {
		lines := make([][]byte, 0)

		r1, err := regexp.Compile("Content-Disposition:\\sform-data;\\sname=\"\\w+\"")
		if err != nil {
			logger.L.Error(err)
		}
		r2, err := regexp.Compile("Content-Disposition:\\sform-data;\\sname=\"\\w+\";\\sfilename=\"\\w+\\.\\w+\"")
		if err != nil {
			logger.L.Error(err)
		}

		for i := 0; i < 2; i++ {
			l, err := SingleLineRightTrimmed(b, limit)
			if err != nil {
				return nil, err
			}

			lines = append(lines, l)

			//logger.L.Infof("in repo.GetLinesMiddle in loop iteration %d lines: %q\n", i, lines)

			if r1.Match(l) && !r2.Match(l) { // check if line matches second regexp; if false header is finished, breaking
				break
			}
			b = b[len(l)+2:]
			limit -= len(l)

		}

		return lines, nil
	}
*/
func LastBoundary(b, boundary []byte) bool {
	//logger.L.Infof("in repo.LastBoundary b: %q wilth lenght = %d boundary: %q with length = %d, sum = %d\n", b, len(b), boundary, len(boundary), bytes.Index(b, boundary)+len(boundary))
	if len(b) > bytes.Index(b, boundary)+len(boundary) &&
		bytes.Contains(b, boundary) {
		//logger.L.Infof("in repo.LastBoundary b: %q, boundary: %q b[bytes.Index(b, boundary)+len(boundary)] = %d\n", b, boundary, b[bytes.Index(b, boundary)+len(boundary)])
		return b[bytes.Index(b, boundary)+len(boundary)] != 13
	}
	return false

}

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

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func GetLinesRight1(b []byte, fromIndex, limit int) ([][]byte, error) {
	//logger.L.Infof("in repo.GetLinerRight1 passed b: %q\n", b)

	lines := make([][]byte, 0)
	for i := 0; i < 2; i++ {
		l, err := GetLineRight(b, fromIndex, limit)
		//logger.L.Infof("in repo.GetLinerRight1 l: %q, err: %v\n", l, err)
		if err != nil {
			if err == errors.New("in repo.GetLineRight limit exceeded. No separator found") {
				return lines, err
			}
			//logger.L.Infof("in repo.GetLinerRight1 almost came err: %v \n", err)
			if err.Error() == "in repo.GetLineRight EOF reached. No separator found" {
				lines = append(lines, l)
				//logger.L.Infof("in repo.GetLinerRight1 err path l: %q, err: %v\n", l, err)
				return lines, err
			}
		}
		if len(l) > 0 &&
			(len(l) < len(CD) && bytes.Contains([]byte(CD), l) ||
				len(l) >= len(CD) && bytes.Contains(l, []byte(CD))) {
			lines = append(lines, l)
			//logger.L.Infof("in repo.GetLinerRight1 lines: %q, err: %v\n", lines, err)
			if len(b) < len(l)+len(Sep) {
				return lines, nil
			}
			b = b[len(l)+len(Sep):]
			//logger.L.Infof("in repo.GetLinerRight1 b after slicing: %q\n", b)
		}
		if len(l) > 0 &&
			(len(l) < len(CType) && bytes.Contains([]byte(CType), l) ||
				len(l) >= len(CType) && bytes.Contains(l, []byte(CType))) {
			lines = append(lines, l)
			//logger.L.Infof("in repo.GetLinerRight1 lines: %q, err: %v\n", lines, err)
			break
		}
		//logger.L.Infof("in repo.GetLinerRight1 b brfore slicing: %q\n", b)

		//logger.L.Infof("in repo.GetLinerRight1 b after slicing: %q\n", b)
	}
	return lines, nil
}

func BoundaryLastPart(b, boundary []byte, fromIndex int) []byte {

	bb := LineRightLimit(b, fromIndex, len(boundary))

	if len(bb) == 0 {
		return nil
	}
	if !bytes.Contains(boundary, bb) {
		return nil
	}

	return bb
}

func BoundaryFirstPart(b, boundary []byte, fromIndex int) []byte {

	//logger.L.Infof("in BoundaryFirstPart b: %q, len(b) = %d\n", b, len(b))

	bb := LineLeftLimit(b, fromIndex, len(boundary))

	if len(bb) == 0 {
		return nil
	}
	if !bytes.Contains(boundary, bb) {
		return nil
	}

	return bb
}

// ToDo Slicer refactor

// Returning last line with preceding CRLF and with following CR if it's met. If last bytes are CRLF, returning just it
func GetLastLine(b, boundary []byte) []byte {
	lenb := len(b)

	switch lc := b[lenb-1]; {
	case lc == 10:
		if b[lenb-2] == 13 {
			//logger.L.Infof("in repo.GetLastLine 10 branch returning: %q\n", b[lenb-2:])
			return b[lenb-2:]
		}
	default:
		l := make([]byte, 0)
		npl, err := CurrentLineFirstPrintIndexLeft(b, len(boundary))
		if err != nil {
			return nil
		}

		//line := LineLeftLimit(b, npl, len(boundary)+1)
		line, err := SingleLineLeftTrimmed(b, len(boundary)+1)
		if err != nil {
			return nil
		}
		//logger.L.Infof("in repo.GetLastLine line: %q\n", line)
		if lenb > len(line)+2 && // checking if there is separator before lastline
			b[npl-len(line)-1] == 13 {
			l = append(l, b[npl-len(line)-1], b[npl-len(line)]) // appending lastline and separator
			l = append(l, b[bytes.Index(b, line):]...)
			//l = append(l, b[npl+1:]...)

			return l
		}
	}
	return nil
}

// Returns last line of b if boundary contains it
func BoundaryPartInLastLine(b []byte, bou Boundary) ([]byte, error) {
	i, lenb, line := 0, 0, make([]byte, 0)
	lenb = len(b)
	i = lenb - 1
	boundary := GenBoundary(bou)
	//logger.L.Infof("in repo.BoundaryPartInLastLine b: %q, boundary: %q\n", b, boundary)

	if i < 1 {
		return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary got 0-length last line")
	}
	if !bytes.Contains(boundary, b[lenb-1:]) {
		//logger.L.Infof("in BoundaryPartInLastLine default and if occured %q has no %q\n", boundary, b[lenb-1])
		l, err := GetCurrentLineLeft(b, i, MaxLineLimit)
		if err != nil { //Test that
			return nil, err
		}
		//logger.L.Infof("in repo.BoundaryPartInLastLine Sep : %q\n", b[lenb-len(l)-2:lenb-len(l)])
		if bytes.Contains(b[lenb-len(l)-2:lenb-len(l)], []byte("\r\n")) {
			ll := []byte("\r\n")
			ll = append(ll, l...)

			if len(ll) < len(boundary) {
				boundary = boundary[:len(ll)]
			}

			//logger.L.Infof("in repo.BoundaryPartInLastLine l: %q, ll : %q\n", l, ll)
			for j, v := range ll[:len(boundary)] {
				if boundary[j] != v {
					return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary")
				}
			}
			return append([]byte("\r\n"), l...), nil
		}
		return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary")
	}
	//	logger.L.Infof("in repo.BoundaryPartInLastLine lenb > len(boundary)? %t\n", lenb > len(boundary))
	/*
		if lenb > len(boundary) {
			logger.L.Infof("in repo.BoundaryPartInLastLine b[lenb-1] == 13? %t, bytes.Contains(boundary, b[lenb-len(boundary)-1:lenb-1]? %t\n", b[lenb-1] == 13, bytes.Contains(boundary, b[lenb-len(boundary)-1:lenb-1]))
		}
	*/
	if lenb > len(boundary) &&
		b[lenb-1] == 13 &&
		bytes.Contains(boundary, b[lenb-len(boundary)-1:lenb-1]) {
		//  logger.L.Infof("in repo.BoundaryPartInLastLine got here\n")
		return b[(lenb - len(boundary) - 1):], nil
	}
	for i >= 0 && bytes.Contains(boundary, b[i:]) {
		line = b[i:]
		//logger.L.Infof("in repo.BoundaryPartInLastLine line: %q is part of boundary: %q\n", line, boundary)
		i--
	}
	//logger.L.Infof("in repo.BoundaryPartInLastLine line: %q\n", line)
	for j, w := range line {
		//logger.L.Infof("in BoundaryPartInLastLine comparing %d with %d\n", boundary[j], w)
		if boundary[j] != w {
			return nil, fmt.Errorf("in repo.BoundaryPartInLastLine no boundary")
		}
	}

	return line, nil
}

// Returns part of b between beg and end, or error
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

// Returns index of not-first occurence of occ in byte slice
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

// Returns true if first and second slices have equal characters on joint indexes
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

// Returns true if second slice is ending part of long one
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

// Parses header line to get formName and fileName
func GetFoFi(b []byte) (string, string) {
	//logger.L.Infof("repo.GetFoFi invoked with b: %q\n", b)
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

// Returns concatenated lines f and l based on their lengths
func JoinLines(f, l [][]byte) [][]byte {
	j := make([][]byte, 0)
	//logger.L.Infof("repo.Joinlines is invoked for f: %q and l:%q\n", f, l)
	if len(f) == 0 || len(l) == 0 {
		return j
	}
	j = append(j, f...)
	switch len(f) {
	case 1: // 3 possible cases
		switch len(l) {
		case 1:
			j = append(j, l...)
		case 2:
			j[len(j)-1] = append(j[len(j)-1], l[0]...)
			j = append(j, l[1])
		case 3:
			j[len(j)-1] = append(j[len(j)-1], l[0]...)
			j = append(j, l[1:]...)
		}
	case 2: //only possible if len(l)==1
		j[len(j)-1] = append(j[len(j)-1], l[0]...)
	}
	return j
}

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

		//logger.L.Infof("in repo.GetHeaderLines line: %q IsCDRight? %t, IsCTLeft? %t\n", b, IsCDRight(b), IsCTLeft(b))

		if IsCDRight(b) {
			return b, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", b)
		}

		return nil, fmt.Errorf("in repo.GetHeaderLines no header found")

	case 1: // CD full + CRLF || CD full + CRLF + CT -> || CRLF || <-LastBoundary + CRLF

		l0 := b[:bytes.Index(b, []byte("\r\n"))]
		l1 := b[bytes.Index(b, []byte("\r\n"))+2:]

		//logger.L.Infof("in repo.GetHeaderLines l0: %q, KnownBoundaryPart(l0, bou); len(lb) > 0 && len(lb) < len(GenBoundary(bou)[2:]) && b[len(lb)+1] != 13 ? %t\n", l0, len(KnownBoundaryPart(l0, bou)) > 0 && len(KnownBoundaryPart(l0, bou)) < len(GenBoundary(bou)[2:]) && b[len(KnownBoundaryPart(l0, bou))+1] != 13)
		//logger.L.Infof("in repo.GetHeaderLines l1: %q\n", l1)

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
		//logger.L.Infof("in repo.GetHeaderLines l0: %q, Sufficiency? %d, IsCDLeft? %t, l1: %q, IsCTFull? %t\n", l0, Sufficiency(l0), IsCDLeft(l0), l1, IsCTFull(l1))

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
			//logger.L.Infof("in repo.GetHeaderLines resl suf: %q\n", resL)
			return resL, nil
		}

		if Sufficiency(l0) == Insufficient { // on ending part CDInsuf + CRLF + CT + CRLF, on beginning part is impossible
			resL = append(l0, []byte("\r\n")...)
			if IsCTFull(l1) {
				resL = append(resL, l1...)
				resL = append(resL, []byte("\r\n")...)
				//logger.L.Infof("in repo.GetHeaderLines resl: %q\n", resL)
				return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is not full", resL)
			}
		}
		//logger.L.Infof("in repo.GetHeaderLines l0: %q, IsCDLeft(l0)? %t, l1: %q, len(l1) = %d\n", l0, IsCDLeft(l0), l1, len(l1))
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
		//logger.L.Infof("in repo.GetHeaderLines l0: %q, Sufficiency(l0) == %d\n", l0, Sufficiency(l0))
		//logger.L.Infof("in repo.GetHeaderLines l0: %q, EndingOf(GenBoundary(bou)[2:], l0): %t\n", l0, EndingOf(GenBoundary(bou)[2:], l0))
		l1 := b[RepeatedIntex(b, []byte("\r\n"), 1)+2 : RepeatedIntex(b, []byte("\r\n"), 2)]
		//logger.L.Infof("in repo.GetHeaderLines l1: %q, IsCTFull(l1) == %t\n", l1, IsCTFull(l1))
		l2 := b[RepeatedIntex(b, []byte("\r\n"), 2)+2 : RepeatedIntex(b, []byte("\r\n"), 3)]

		//logger.L.Infof("in repo.GetHeaderLines l0: %q,EndingOf(append(GenBoundary(bou)[2:], []byte(\"\r\n\")...)? %t, l1: %q, Sufficiency(l1) == Insufficient? %t, l2: %q, IsCTFull? %t\n", l0, EndingOf(append(GenBoundary(bou)[2:], []byte("\r\n")...), l0), l1, Sufficiency(l1) == Insufficient, l2, IsCTFull(l2))
		if len(l0) >= 0 && EndingOf(GenBoundary(bou)[2:], l0) && (Sufficiency(l1) == Insufficient || Sufficiency(l1) == Sufficient) {
			resL = append(l0, []byte("\r\n")...)
			//logger.L.Infof("in repo.GetHeaderLines resl: %q\n", resL)
			if Sufficiency(l1) == Insufficient {
				resL = append(resL, l1...)
				resL = append(resL, []byte("\r\n")...)
				//logger.L.Infof("in repo.GetHeaderLines resl: %q\n", resL)
				if IsCTFull(l2) {
					resL = append(resL, l2...)
					resL = append(resL, []byte("\r\n\r\n")...)
					//logger.L.Infof("in repo.GetHeaderLines resl: %q\n", resL)
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
			//logger.L.Infof("in repo.GetHeaderLines resl: %q\n", resL)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}
		if len(l2) == 0 { // on ending part CDinsuf + CRLF + CT + 2*CRLF + rand, on beginning part CRLF + CDsuf + 2*CRLF = rand || <-Bound + CRLF + CDsuf + 2*CRLF = rand || <-CDinsuf + CRLF + CT + 2*CRLF
			resL = append(l0, []byte("\r\n")...)
			resL = append(resL, l1...)
			resL = append(resL, []byte("\r\n\r\n")...)
			//logger.L.Infof("in repo.GetHeaderLines resl: %q\n", resL)
			return resL, fmt.Errorf("in repo.GetHeaderLines header \"%s\" is ending part", resL)
		}
		return nil, fmt.Errorf("in repo.GetHeaderLines no header found")

	}
}

// Returns known part of last boundary
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

func IsLastBoundary(p, n []byte, bou Boundary) bool {
	realBoundary := GenBoundary(bou)
	//logger.L.Infof("in repo.IsLastBoundary realBoundary: %q\n", realBoundary)
	combined := append(p, n...)
	//logger.L.Infof("in repo.IsLastBoundary combined: %q\n", combined)
	//logger.L.Infof("in repo.IsLastBoundary len(combined) < len(realBoundary)? %t, !ContainsBouEnding(combined, bou)? %t\n", len(combined) < len(realBoundary), !ContainsBouEnding(combined, bou))

	if len(combined) > len(realBoundary) &&
		bytes.Contains(combined, realBoundary) &&
		combined[len(realBoundary)] != 13 &&
		bytes.Contains(combined[len(combined)-2:], []byte("\r\n")) {
		return true
	}

	return false
}

// Returns CRLF and succeeding line before given index. If line ends with CR (or CRLF) and contains boundary, returns CRLF + line + CR (CRLF)
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

// Returns true if b contains boundary ending
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
