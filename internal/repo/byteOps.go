package repo

import (
	"bytes"
)

func FindBoundaryB(b []byte) BoundaryB {

	bPrefix, bRoot := []byte{}, []byte{}

	if bytes.Contains(b, []byte(BoundaryField)) {

		start := bytes.Index(b, []byte(BoundaryField)) + len(BoundaryField)

		bRoot = b[start:LineEndPosLimitB(b, start, 70)]

		bb := b[start+1:]

		secBoundaryIndex := bytes.Index(bb, bRoot) - 1

		bPrefix = PrevLineLimit(bb, secBoundaryIndex, MaxLineLimit)
	}

	return BoundaryB{
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

func PrevLineLimit(b []byte, fromIndex, limit int) []byte {
	bb := make([]byte, 0)
	if fromIndex <= 0 {
		return bb
	}
	for i := fromIndex; b[i] != 10 && i > fromIndex-limit; i-- {
		bb = append(bb, b[i])
	}
	bbs := ReverseB(bb)

	return bbs
}

func ReverseB(bb []byte) []byte {

	bbs := make([]byte, 0)

	for i := len(bb) - 1; i >= 0; i-- {
		bbs = append(bbs, bb[i])
	}

	return bbs
}

func Slicer(b []byte, bou BoundaryB) (AppPieceUnit, []AppPieceUnit, AppPieceUnit) {
	boundary := make([]byte, 0)
	boundary = append(boundary, bou.Prefix...)
	boundary = append(boundary, bou.Root...)

	boundNum := bytes.Count(b, boundary)

	switch boundNum {
	case 0:
		pbl := PartlyBoundaryLen(b, bou)
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
		pbl := PartlyBoundaryLen(b, bou)

		aphb := NewAppPieceHeader()
		aphb.SetB(true)
		aphb.SetE(false)

		apbb := NewAppPieceBody(b[:bi-2])
		apub := NewAppPieceUnit(aphb, apbb)

		aphe := NewAppPieceHeader()
		aphe.SetB(false)

		if pbl > 0 {
			aphe.SetE(true)

			apb := NewAppPieceBody(b[:len(b)-pbl-2])
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
		pbl := PartlyBoundaryLen(b, bou)
		m := make([]AppPieceUnit, 0)

		aphb := NewAppPieceHeader()
		aphb.SetB(true)
		aphb.SetE(false)

		apbb := NewAppPieceBody(b[:bi-2])
		apub := NewAppPieceUnit(aphb, apbb)

		b = b[bi:]

		for i := 0; i < boundNum-1; i++ {
			aph := NewAppPieceHeader()
			aph.SetB(false)
			aph.SetE(false)
			ni := bytes.Index(b[1:], boundary) - 1
			apb := NewAppPieceBody(b[:ni])
			apu := NewAppPieceUnit(aph, apb)
			m = append(m, apu)

			b = b[ni+2:]
		}
		aphe := NewAppPieceHeader()
		aphe.SetB(false)
		if pbl > 0 {

			aphx := NewAppPieceHeader()
			aphx.SetB(false)
			aphx.SetE(false)
			apbx := NewAppPieceBody(b[:len(b)-pbl-2])
			apux := NewAppPieceUnit(aphx, apbx)

			m = append(m, apux)

			aphe.SetE(true)
			apbe := NewAppPieceBody(b[len(b)-pbl:])
			apue := NewAppPieceUnit(aphe, apbe)

			return apub, m, apue
		}

		aphe.SetE(true)

		apbe := NewAppPieceBody(b)
		apue := NewAppPieceUnit(aphe, apbe)

		return apub, m, apue
	}
}

func PartlyBoundaryLen(b []byte, bou BoundaryB) int {
	lenb := len(b)
	last := b[lenb-1]
	boundary := genBoundary(bou)
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
	boundaryLen := lenb - boundaryStartPos
	occ = 0

	for i := lenb - 1; i >= boundaryStartPos && i-lenb+1+boundaryLen > 0; i-- {
		if boundary[i-lenb+boundaryLen] != b[i] {
			return -1
		}
		occ++
	}

	return occ

}

func genBoundary(bou BoundaryB) []byte {
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

/*
func Slicer(s string, voc Vocabulaty) []Piece {
	last := len(s) - 1
	pieces := make([]Piece, 0)
	pieces = append(pieces, NewPiece())
	pieces[len(pieces)-1].SetNeedsBegin(true)

	boundNum := strings.Count(s, voc.Boundary.Prefix+voc.Boundary.Root)

	if boundNum == 1 {
		pieces[0].SetContent(s)

		return pieces
	}

	lastIndex := LineStartPosLimit(s, last, len(voc.Boundary.Prefix+voc.Boundary.Root))

	if lastIndex < 0 {
		lastIndex = len(s)
	}
	lastLineLen := 0

	if lastIndex > 0 {
		lastLineLen = len(s[lastIndex:])
	}

	firstPIndex := strings.Index(s, voc.Boundary.Prefix+voc.Boundary.Root)

	pieces[len(pieces)-1].SetContent(s[:firstPIndex-2])

	s = s[firstPIndex:]

	for i := 0; i < boundNum-1; i++ {
		nextPIndex := strings.Index(s[1:], voc.Boundary.Prefix+voc.Boundary.Root) + 1

		pieces = append(pieces, NewPiece())
		pieces[len(pieces)-1].SetNeedsBegin(false)
		pieces[len(pieces)-1].SetNeedsEnd(false)
		pieces[len(pieces)-1].SetContent(s[:nextPIndex-2])

		s = s[nextPIndex:]
	}

	if lastLineLen > 0 &&
		s[LastPrintPosLimit(s, last, len(voc.Boundary.Prefix+voc.Boundary.Root)):] == (voc.Boundary.Prefix + voc.Boundary.Root)[:lastLineLen] {

		pieces = append(pieces, NewPiece())
		pieces[len(pieces)-1].SetNeedsEnd(false)
		pieces[len(pieces)-1].SetNeedsBegin(false)
		pieces[len(pieces)-1].SetContent(s[:lastIndex-2])

		pieces = append(pieces, NewPiece())
		pieces[len(pieces)-1].SetNeedsEnd(true)
		pieces[len(pieces)-1].SetNeedsBegin(false)
		pieces[len(pieces)-1].SetContent(s[lastIndex:])

	} else {
		pieces = append(pieces, NewPiece())
		pieces[len(pieces)-1].SetNeedsBegin(false)
		pieces[len(pieces)-1].SetNeedsEnd(true)
		pieces[len(pieces)-1].SetContent(s)
	}

	return pieces
}
*/
