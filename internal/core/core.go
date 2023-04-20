package core

import (
	"github.com/vynovikov/postParser/internal/logger"
	"github.com/vynovikov/postParser/internal/repo"

	"github.com/google/go-cmp/cmp"
)

type Core struct{}

type Parser interface {
	Slicer(repo.AppFeederUnit) (repo.AppPieceUnit, []repo.AppPieceUnit, repo.AppSub)
	ParseBegin(repo.AppPieceUnit, repo.Vocabulaty) ([][]byte, error)
	ParseMiddle(repo.AppPieceUnit, repo.Vocabulaty) ([][]byte, error)
	ParseEnd(repo.AppSub, repo.Vocabulaty) ([][]byte, error)
}

func NewCore() Parser {
	return &Core{}
}
func (c *Core) Slicer(afu repo.AppFeederUnit) (repo.AppPieceUnit, []repo.AppPieceUnit, repo.AppSub) {

	b, m, e := repo.Slicer(afu.R.B.B, afu.R.H.Bou)
	if !cmp.Equal(b, repo.AppPieceUnit{}) {
		b.APH.SetPart(afu.R.H.Part)
		b.APH.SetTS(afu.R.H.TS)
		//	logger.L.Warnf("in core.Slicer in part %d got begin piece with body: %q\n", afu.R.H.Part, b.APB.B)
	}

	if !cmp.Equal(e, repo.AppSub{}) {
		e.SetPart(afu.R.H.Part)
		e.SetTS(afu.R.H.TS)
		//	logger.L.Warnf("in core.Slicer in part %d got subPiece with body: %q\n", afu.R.H.Part, e.B)
	}

	for i := range m {
		m[i].APH.SetPart(afu.R.H.Part)
		m[i].APH.SetTS(afu.R.H.TS)
		//	logger.L.Warnf("in core.Slicer in part %d got middle piece with body: %q\n", afu.R.H.Part, m[i].APB.B)
	}
	return b, m, e
}

// Parsing first lines of byte slice, detecting if this lines contain Content-Disposition header part
// ToDo refactor to catch errors
func (c *Core) ParseBegin(apu repo.AppPieceUnit, voc repo.Vocabulaty) ([][]byte, error) {
	var (
		lines [][]byte
		//		err   error
	)

	logger.L.Infof("in core.ParseBegin in part %d body: %q\n", apu.APH.Part, apu.APB.B)

	/*
		lines, err := repo.GetLinesRight(apu.APB.B, repo.MaxLineLimit, voc)
		if err != nil {
			return nil, err
		}

		logger.L.Infof("in core.ParseBegin in part %d lines: %q\n", apu.APH.Part, repo.GetLinesRight(apu.APB.B, repo.MaxLineLimit, voc))
	*/

	return lines, nil
}

func (c *Core) ParseMiddle(apu repo.AppPieceUnit, voc repo.Vocabulaty) ([][]byte, error) {

	var (
		lines [][]byte
		//		err   error
	)

	logger.L.Infof("in core.ParseMiddle in part %d body: %q\n", apu.APH.Part, apu.APB.B)

	/*
		lines, err := repo.GetLinesMiddle(apu.APB.B, repo.MaxHeaderLimit, voc)
		if err != nil {
			return nil, err
		}
		logger.L.Infof("in core.ParseMiddle in part %d lines %q\n", apu.APH.Part, lines)
	*/

	return lines, nil
}

// Parsing last lines of byte slice, detecting if lines contain Content-Disposition header part
func (c *Core) ParseEnd(as repo.AppSub, voc repo.Vocabulaty) ([][]byte, error) {
	var (
		lines [][]byte
		//		err   error
	)

	logger.L.Infof("in core.ParseEnd in part %d body: %q\n", as.Part, as.B)
	/*	if len(as.B) > 0 {

			lines, err = repo.GetLinesMiddle(apu.APB.B, repo.MaxLineLimit, voc)
			if err != nil {
				return nil, err
			}

			return lines, err

		}

		lines, err = repo.GetLinesLeft(apu.APH.H, repo.MaxHeaderLimit, voc)
		if err != nil {
			return nil, err
		}
	*/
	//logger.L.Infof("in core.ParseEnd in part %d lines: %q\n", apu.APH.Part, lines)

	return lines, nil
}
