package core

import (
	"postParser/internal/logger"
	"postParser/internal/repo"

	"github.com/google/go-cmp/cmp"
)

type Core struct{}

type Parser interface {
	Slicer(repo.AppFeederUnit) (repo.AppPieceUnit, []repo.AppPieceUnit, repo.AppPieceUnit)
	ParseBegin(repo.AppPieceUnit, repo.Vocabulaty) ([][]byte, error)
	ParseMiddle(repo.AppPieceUnit, repo.Vocabulaty) ([][]byte, error)
	ParseEnd(repo.AppPieceUnit, repo.Vocabulaty) ([][]byte, error)
}

func NewCore() Parser {
	return &Core{}
}
func (c *Core) Slicer(afu repo.AppFeederUnit) (repo.AppPieceUnit, []repo.AppPieceUnit, repo.AppPieceUnit) {
	boundary := repo.GenBoundary(afu.R.H.Voc.Boundary)

	b, m, e := repo.Slicer(afu.R.B.B, boundary)
	if !cmp.Equal(b, repo.AppPieceUnit{}) {
		b.APH.SetPart(afu.R.H.Part)
		b.APH.SetTS(afu.R.H.TS)
	}

	if !cmp.Equal(e, repo.AppPieceUnit{}) {
		e.APH.SetPart(afu.R.H.Part)
		e.APH.SetTS(afu.R.H.TS)
	}

	for _, v := range m {
		v.APH.SetPart(afu.R.H.Part)
		v.APH.SetTS(afu.R.H.TS)
		logger.L.Infof("in core.Slicer got piece with header %v with body %q\n", v.APH, v.APB.B)
	}
	return b, m, e
}

// Parsing first lines of byte slice, detecting if this lines contain Content-Disposition header part
// ToDo refactor to catch errors
func (c *Core) ParseBegin(apu repo.AppPieceUnit, voc repo.Vocabulaty) ([][]byte, error) {

	//logger.L.Infof("in core.ParseBegin in part %d body: %q\n", apu.APH.Part, apu.APB.B)

	lines, err := repo.GetLinesRight(apu.APB.B, repo.MaxLineLimit, voc)
	if err != nil {
		return nil, err
	}

	//logger.L.Infof("in core.ParseBegin in part %d lines: %q\n", apu.APH.Part, repo.GetLinesRight(apu.APB.B, repo.MaxLineLimit, voc))

	return lines, nil
}

func (c *Core) ParseMiddle(apu repo.AppPieceUnit, voc repo.Vocabulaty) ([][]byte, error) {

	//logger.L.Infof("in core.ParseMiddle in part %d body: %q\n", apu.APH.Part, apu.APB.B)

	lines, err := repo.GetLinesMiddle(apu.APB.B, repo.MaxHeaderLimit, voc)
	if err != nil {
		return nil, err
	}
	//logger.L.Infof("in core.ParseMiddle in part %d lines %q\n", apu.APH.Part, lines)

	return lines, nil
}

// Parsing last lines of byte slice, detecting if lines contain Content-Disposition header part
func (c *Core) ParseEnd(apu repo.AppPieceUnit, voc repo.Vocabulaty) ([][]byte, error) {

	//logger.L.Infof("in core.ParseEnd in part %d body: %q\n", apu.APH.Part, apu.APB.B)

	lines, err := repo.GetLinesLeft(apu.APB.B, repo.MaxHeaderLimit, voc)
	if err != nil {
		return nil, err
	}
	//logger.L.Infof("in core.ParseEnd in part %d lines: %q\n", apu.APH.Part, lines)

	return lines, nil
}
