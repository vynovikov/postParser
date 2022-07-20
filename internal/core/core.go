package core

import (
	"errors"
	"postParser/internal/logger"
	"postParser/internal/repo"
)

type Core struct{}

type Parser interface {
	ParseBegin(repo.AppFeederUnit) (repo.AppDistributorUnit, repo.AppFeederHeaderBP, error)
	ParseEnd(repo.AppFeederUnit)
}

func NewCore() Parser {
	return &Core{}
}

func (c *Core) ParseEnd(afu repo.AppFeederUnit) {
	//"/r"=13 "/n"=10
	//|BBBBBBBBBBBBBBB/r| - separated body
	//|BBBBBBBBBBBBBBB/r/n| - separated body
	//|BBBBBBBBBBBBBBB/r/n/-| - separared boundary
	//|BBBBBBBBBBBBBBB/r/n/----| - separared boundary, end of sending
	//|BBBBBBBBBBBBBBB/r/n/----xxx| - separared boundary, end of sending
	//|BBBBBBBBBBBBBBB/r/n/-----xxxx/r| - separared boundary, end of sending
	//|BBBBBBBBBBBBBBB/r/n/-----xxxx/r/n| - separared boundary, end of sending
	//|BBBBBBBBBBBBBBB/r/n/-----xxxx/r/n/C| - separared Header

	s := ""
	endingLen := len(afu.R.H.Voc.Boundary.Prefix) + len(afu.R.H.Voc.Boundary.Root) + len(repo.Disposition) + len(repo.CType) + 40
	endingLimit := len(repo.Disposition) + 30

	if len(afu.R.B.B) > endingLen {
		s = string(afu.R.B.B[len(afu.R.B.B)-endingLen:])
	} else {
		s = string(afu.R.B.B)
	}

	lens := len(s)
	if lens == 0 {
		return
	}

	lines := repo.GetLinesRev(s, endingLimit, afu.R.H.Voc)
	if len(lines) > 1 {
		afu.H.SepHeader.IsBoundary = false
	}
	if len(lines) == 1 {
		afu.H.SepHeader.IsBoundary = true
	}
	afu.H.PrevPart = afu.R.H.Part
	afu.H.SepHeader.Lines = lines
	logger.L.Infof("in core.ParseEnd afu PrevPart %v\n", afu.H.PrevPart)
	logger.L.Infof("in core.ParseEnd afu SepHeader %v\n", afu.H.SepHeader)

}
func (c *Core) ParseBegin(afu repo.AppFeederUnit) (repo.AppDistributorUnit, repo.AppFeederHeaderBP, error) {
	s, fo, fi, l, firstBodyPos, lastBodyPos := string(afu.R.B.B), "", "", "", 0, 0
	head := repo.NewLines(make([]string, 0), false)

	lens := len(s)
	if lens == 0 {
		return repo.AppDistributorUnit{}, repo.AppFeederHeaderBP{}, errors.New("unable to get afu content")
	}

	beginningLen := len(afu.R.H.Voc.Boundary.Prefix) + len(afu.R.H.Voc.Boundary.Root) + len(repo.Disposition) + len(repo.CType) + 40

	if len(s) > beginningLen {
		s = string(afu.R.B.B)[:beginningLen]
	} else {
		s = string(afu.R.B.B)
	}

	logger.L.Infof("in core.ParseBegin prevLines: %q\n", afu.H.SepHeader.Lines)

	if lens > beginningLen {
		head = repo.GetLinesFw(s[:beginningLen], afu.H.SepHeader.Lines, repo.MaxLineLimit, afu.R.H.Voc)
	} else {
		head = repo.GetLinesFw(s, afu.H.SepHeader.Lines, repo.MaxLineLimit, afu.R.H.Voc)
	}
	logger.L.Infof("in core.ParseBegin curLines: %q\n", head.CurLines)

	lines := repo.JoinLines(afu.H.SepHeader.Lines, head)
	logger.L.Infof("in core.ParseBegin lines: %q\n", lines)

	if len(lines) == 3 {
		fo, fi = repo.FindForm(lines[1], afu.R.H.Voc)
		firstBodyPos = repo.FindFirstBodyPos(s, afu.H.SepHeader.Lines, head.CurLines)
		lastBodyPos, l = repo.FindLastBodyPos(s, firstBodyPos, afu.R.H.Voc.Boundary)
	}
	logger.L.Infof("in core.ParseBegin firstBodyPos: %d and s: %q\n", firstBodyPos, s[firstBodyPos:])
	logger.L.Infof("in core.ParseBegin lastBodyPos: %d\n", lastBodyPos)

	sepB1 := repo.NewSepBodyBP(l)
	afhBlueprint := repo.NewAppFeederHeaderBP(repo.SepHeader{}, sepB1, afu.R.H.Part)

	m := repo.NewMultipartHeader(0)
	dh := repo.NewAppDistributorHeader(m, afu.R.H.TS, fo, fi)
	db := repo.NewDistributorBody(afu.R.B.B[firstBodyPos:lastBodyPos])

	adu := repo.NewAppDistributorUnit(dh, db)

	//logger.L.Infof("in core.ParseBegin adu: %v\n", adu)

	return adu, afhBlueprint, nil

}
