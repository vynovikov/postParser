package core

import (
	"postParser/internal/logger"
	"postParser/internal/repo"
)

type Core struct{}

type Parser interface {
	ParseBegin(repo.AppFeederUnit)
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
	endingLen := len(afu.R.H.Voc.Boundary.Prefix) + len(afu.R.H.Voc.Boundary.Root) + len(repo.Disposition) + 30
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

	lines := repo.GetLines(s, endingLimit, afu.R.H.Voc)
	if len(lines) > 1 {
		afu.H.SepHeader.IsBoundary = false
	}
	if len(lines) == 1 {
		afu.H.SepHeader.IsBoundary = true
	}
	afu.H.PrevPart = afu.R.H.Part
	for _, v := range lines {
		afu.H.SepHeader.PrevBody += v
	}
	logger.L.Infof("in core.ParseEnd afu PrevPart %v\n", afu.H.PrevPart)
	logger.L.Infof("in core.ParseEnd afu SepHeader %v\n", afu.H.SepHeader)

}
func (c *Core) ParseBegin(afu repo.AppFeederUnit) {

}
