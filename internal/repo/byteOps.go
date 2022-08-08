package repo

import (
	"bytes"
	"errors"
	"postParser/internal/logger"
	"regexp"
	"unicode"
)

const (
	BoundaryField  = "boundary="
	Sep            = "\r\n"
	MaxLineLimit   = 100
	MaxHeaderLimit = 210
)

func FindBoundary(b []byte) Boundary {

	bPrefix, bRoot := []byte{}, []byte{}

	if bytes.Contains(b, []byte(BoundaryField)) {

		start := bytes.Index(b, []byte(BoundaryField)) + len(BoundaryField)

		bRoot = b[start:LineEndPosLimitB(b, start, 70)]

		bb := b[start+1:]

		secBoundaryIndex := bytes.Index(bb, bRoot) - 1

		bPrefix = LineLeftLimit(bb, secBoundaryIndex, MaxLineLimit)
	}

	return Boundary{
		Prefix: bPrefix,
		Root:   bRoot,
	}

}
func LineEndPosLimitB(b []byte, fromIndex, limit int) int {
	i := fromIndex

	for i < len(b)-1 && i < fromIndex+limit {
		i++

		if b[i] == 13 {
			return i
		}

	}
	return -1
}

func LineLeftLimit(b []byte, fromIndex, limit int) []byte {
	bb := make([]byte, 0)

	if fromIndex <= 0 {
		return bb
	}
	for i := fromIndex; b[i] != 10 && i > fromIndex-limit; i-- {
		bb = append(bb, b[i])
	}
	bbs := Reverse(bb)

	return bbs
}

func LineRightLimit(b []byte, fromIndex, limit int) []byte {
	bb := make([]byte, 0)

	if fromIndex < 0 {
		return bb
	}
	ffi := fromIndex
	for (b[ffi] == 13 || b[ffi] == 10) &&
		ffi < fromIndex+limit {
		ffi++
	}

	for i := ffi; b[i] != 13 && i < fromIndex+limit; i++ {
		bb = append(bb, b[i])
	}
	return bb
}

// Searches for first printable character index in first line of byte slice with limit. Returns error if limit exceeded.
func CurrentLineFirstPrintIndexRight(b []byte, limit int) (int, error) {

	p := 0 // statring number of printable character is -1 in order of correct work of incremention
	if len(b) == 0 {
		return -1, errors.New("passed byte slice with zero length")
	}

	for i := 0; (b[i] == 13 || b[i] == 10) && i < limit; i++ {
		p++
	}
	if p == limit {
		return -1, errors.New("no actual characters before limit")
	}

	return p, nil
}
func GetCurrentLineRight(b []byte, fromIndex, limit int) ([]byte, error) {

	bb, i, lenb := make([]byte, 0), fromIndex, len(b)

	if lenb == 0 {
		return bb, errors.New("passed byte slice with zero length")
	}
	// checking if line is ending part of last boundary
	//if lenb < limit &&
	//	(!bytes.Contains(b, []byte(CD)) &&
	//		!bytes.Contains(b, []byte(CType))) {
	if lenb < limit {
		for i < lenb {
			bb = append(bb, b[i])
			i++
		}
		return bb, nil
	}

	for i < limit-fromIndex && b[i] != 13 {

		bb = append(bb, b[i])

		if i == limit-fromIndex {
			return nil, errors.New("line limit exceeded. No separator met")
		}

		i++
	}

	return bb, nil
}
func SingleLineRightUnchanged(b []byte, limit int) ([]byte, error) {

	return GetCurrentLineRight(b, 0, limit)
}

func SingleLineRightTrimmed(b []byte, limit int) ([]byte, error) {

	index, err := CurrentLineFirstPrintIndexRight(b, limit)
	if err != nil {
		return nil, err
	}
	//logger.L.Infof("in repo.SingleLineRightTrimmed first printed index %d\n", index)
	return GetCurrentLineRight(b, index, limit-index)
}

// Returns index of first printable character index to the left of byte slice end with limit, or error if limit exceeded
func CurrentLineFirstPrintIndexLeft(b []byte, limit int) (int, error) {
	lenb := len(b)
	i := lenb - 1
	//logger.L.Infof("in repo.CurrentLineFirstPrintIndexLeft before loop b: %q, lenb = %d, i = %d\n", b, lenb, i)
	if lenb == 0 {
		return -1, errors.New("passed byte slice with zero length")
	}

	//logger.L.Infof("in repo.CurrentLineFirstPrintIndexLeft before loop i = %d b[i] = %d\n", i, b[i])

	for (b[i] == 13 || b[i] == 10) &&
		i > lenb-1-limit {
		//logger.L.Infof("in repo.CurrentLineFirstPrintIndexLeft in loop i = %d b[i] = %d\n", i, b[i])
		i--
	}
	if i == lenb-1-limit {
		return -1, errors.New("no actual characters before limit")
	}

	return i, nil
}

// Checks byte slice to the left, returns line before separator or error if limit exceeded
func GetCurrentLineLeft(b []byte, fromIndex, limit int) ([]byte, error) {
	//logger.L.Infof("repo.GetCurrentLineLeft invoked with b = %q, fromindex = %d limit = %d\n", b, fromIndex, limit)
	bb, i, lenb := make([]byte, 0), fromIndex, len(b)
	if lenb == 0 {
		return bb, errors.New("passed byte slice with zero length")
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
func Slicer(b, boundary []byte) (AppPieceUnit, []AppPieceUnit, AppPieceUnit) {
	//	logger.L.Infof("in repo.Slicer body: %q and bounddary = %q\n", b, boundary)
	boundNum := bytes.Count(b, boundary)

	//logger.L.Infof("in repo.Slicer body: %q and boundNum = %d\n", b, boundNum)

	switch boundNum {
	case 0:
		pbl := PartlyBoundaryLenLeft(b, boundary)
		ph := NewAppPieceHeader()
		ph.SetB(true)

		if pbl > 0 {

			ph.SetE(false)
			pb := NewAppPieceBody(b[:len(b)-pbl-2])
			pub := NewAppPieceUnit(ph, pb)

			phe := NewAppPieceHeader()
			phe.SetB(false)
			phe.SetE(true)
			pbe := NewAppPieceBody(b[len(b)-pbl:])
			pue := NewAppPieceUnit(phe, pbe)

			return pub, nil, pue

		}

		ph.SetE(true)
		pb := NewAppPieceBody(b)
		apu := NewAppPieceUnit(ph, pb)

		return apu, nil, apu

	case 1:

		bi := bytes.Index(b, boundary)

		// pre-boundary part of byte slice
		// todo trimmer function (there are 2 Separators before last boundary)
		pb := b[:bi-3*len(Sep)]

		pbl := PartlyBoundaryLenLeft(b, boundary)
		apub := NewAppPieceUnitEmpty()

		aphb := NewAppPieceHeader()
		aphb.SetB(true)

		aphb.SetE(false)

		apbb := NewAppPieceBody(pb)

		// if current part != 0 (no occurence of BoundaryField), slicing and setting up, otherwise leaving it empty
		if !bytes.Contains(pb, []byte(BoundaryField)) {
			apub.SetAPH(aphb)
			apub.SetAPB(apbb)
		}
		logger.L.Infof("in repo.Slicer case 1 begin piece header %v,body: %q", apub.APH, apub.APB.B)

		b = b[bi+len(boundary)+2:] //trim byte slice to exclude boundary itself
		aphe := NewAppPieceHeader()
		aphe.SetB(false)

		be, err := GetCurrentLineRight(b, bi+len(boundary), MaxLineLimit)
		if err != nil {
			logger.L.Error((err))
		}
		if be != nil {

			return apub, nil, AppPieceUnit{}
		}

		if pbl > 0 {
			aphe.SetE(true)

			apb := NewAppPieceBody(b[len(boundary)+2 : len(b)-pbl-2])
			apu := NewAppPieceUnit(aphe, apb)

			m := make([]AppPieceUnit, 0)
			m = append(m, apu)

			aphe := NewAppPieceHeader()
			aphe.SetB(false)
			aphe.SetE(true)
			apbe := NewAppPieceBody(b[len(b)-pbl:])
			apue := NewAppPieceUnit(aphe, apbe)

			return apub, m, apue

		}
		aphe.SetE(true)
		apbe := NewAppPieceBody(b[bi:])

		apue := NewAppPieceUnit(aphe, apbe)

		return apub, nil, apue

	default:
		bi := bytes.Index(b, boundary)
		pbl := PartlyBoundaryLenLeft(b, boundary)
		apub := NewAppPieceUnitEmpty()
		m := make([]AppPieceUnit, 0)

		aphb := NewAppPieceHeader()
		aphb.SetB(true)
		aphb.SetE(false)

		apbb := NewAppPieceBody(b[:bi-2])
		if !bytes.Contains(b, []byte(BoundaryField)) {
			apub.SetAPH(aphb)
			apub.SetAPB(apbb)

		}
		//apub := NewAppPieceUnit(aphb, apbb)

		b = b[bi+len(boundary)+2:]

		for i := 0; i < boundNum-1; i++ {
			aph := NewAppPieceHeader()
			aph.SetB(false)
			aph.SetE(false)
			ni := bytes.Index(b[1:], boundary) - 1
			//logger.L.Infof("in repo.Slicer in loop b: %q\n", b[:ni])
			apb := NewAppPieceBody(b[:ni])
			apu := NewAppPieceUnit(aph, apb)
			logger.L.Infof("in repo.Slicer in loop b: %q\n", apu.APB.B)
			m = append(m, apu)

			//logger.L.Infof("in repo.Slicer in loop b: %q, \n", b)
			b = b[len(Sep)+ni+len(boundary)+len(Sep):]
		}

		ni := bytes.Index(b[1:], boundary) - 1
		b = b[ni+2:]

		aphe := NewAppPieceHeader()
		aphe.SetB(false)
		if pbl > 0 {

			aphx := NewAppPieceHeader()
			aphx.SetB(false)
			aphx.SetE(false)
			//logger.L.Infof("in repo.Slicer in if b: %q\n", b[:len(b)-pbl-2])

			apbx := NewAppPieceBody(b[:len(b)-pbl-3*len(Sep)])
			apux := NewAppPieceUnit(aphx, apbx)

			logger.L.Infof("in repo.Slicer last in mid chunk with body: %q\n", apux.APB.B)
			m = append(m, apux)

			aphe.SetE(true)
			apbe := NewAppPieceBody(b[len(b)-pbl:])
			apue := NewAppPieceUnit(aphe, apbe)
			logger.L.Infof("in repo.Slicer e with body: %q\n", apue.APB.B)

			return apub, m, apue
		}

		aphe.SetE(true)

		apbe := NewAppPieceBody(b)
		apue := NewAppPieceUnit(aphe, apbe)

		return apub, m, apue
	}
}

func PartlyBoundaryLenLeft(b, boundary []byte) int {
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
	boundaryStartPos := LineStartPosLimitB(b, lenb-1, len(boundary))
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

// returrns true if byte slice has all printable charackters/
func IsPrintable(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if !unicode.IsPrint(rune(b[i])) {
			return false
		}
	}
	return true
}

//generates boundary based on given Boundary struct
func GenBoundary(bou Boundary) []byte {
	boundary := make([]byte, 0)
	boundary = append(boundary, bou.Prefix...)
	boundary = append(boundary, bou.Root...)
	return boundary
}

func LineStartPosLimitB(b []byte, fromIndex, limit int) int {

	i := fromIndex

	for i > fromIndex-limit {
		if b[i] == 10 {
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

// returns header lines on current data chunk which are hyphennated from previous data chunk
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
		l, err := SingleLineRightTrimmed(b[:limit], limit)
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
