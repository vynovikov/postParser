package application

import (
	grpc "postParser/internal/adapters/driven/grpc"
	"postParser/internal/core"
	"postParser/internal/logger"
	"postParser/internal/repo"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
)

type AppService struct {
	chanIn  chan repo.AppFeederUnit
	chanOut chan repo.AppDistributorUnit
	mp      map[PieceKey][]repo.AppPieceUnit
	mpLock  sync.Mutex
}

type PieceKey struct {
	TS   string
	Part int
	B    bool
}

func NewBlockInfo() *repo.BlockInfo {
	return &repo.BlockInfo{
		TS:       time.Now().Local(),
		Boundary: "---xxx",
	}
}

type App struct {
	T grpc.Transmitter
	C core.Parser
	A AppService
}

func NewApplication(c core.Parser, t grpc.Transmitter) *App {

	App := &App{
		T: t,
		C: c,
		A: NewAppService(),
	}

	App.Do()
	return App
}

func NewAppService() AppService {
	return AppService{
		chanIn:  make(chan repo.AppFeederUnit, 10),
		chanOut: make(chan repo.AppDistributorUnit, 10),
	}
}

type Application interface {
	Do()
	AddToFeeder(repo.ReceiverUnit)
}

func (a *App) Do() {
	go a.Work()

	go a.Send()

}

func (a *App) toChanIn(afu repo.AppFeederUnit) {

	a.A.chanIn <- afu
}

func (a *App) toChanOut(afu repo.AppDistributorUnit) {
	//logger.L.Infof("in application.toChanOut adding to chanOUT: part %v\n", afu.H)

	a.A.chanOut <- afu

}

func (a *App) Work() {

	for afu := range a.A.chanIn {

		//Slicing feederUnit bytes chunk corrsesponding to boundary appearance
		b, m, e := a.C.Slicer(afu)

		if !cmp.Equal(b, repo.AppPieceUnit{}) { // Beginning piece is empty only in zero part, leaving it
			go a.HandleBegin(b, afu.R.H.Voc)
		}

		if len(m) > 0 {
			for _, v := range m {
				go a.HandleMiddle(v, afu.R.H.Voc)
			}
		}

		if !cmp.Equal(e, repo.AppPieceUnit{}) { // Ending piece is empty only after last boundary, leaving it
			go a.HandleEnd(e, afu.R.H.Voc)
		}
	}

}

func (a *App) AddToFeeder(in repo.ReceiverUnit) {

	A := repo.NewAppFeederUnit(in)

	if (in.S == repo.ReceiverSignal{}) {

		a.A.chanIn <- A
	}

}

func (a *App) Send() {
	for adu := range a.A.chanOut {

		//	logger.L.Infof("sending to rpc %v\n part %d\n", adu.H.Name, adu.H.TS)

		a.T.Transmit(adu)
	}
}
func (a *App) HandleBegin(apu repo.AppPieceUnit, voc repo.Vocabulaty) {
	lines, err := a.C.ParseBegin(apu, voc)
	if err != nil {
		logger.L.Errorf("in application.HandleBegin error: %v\n", err)
	}
	logger.L.Infof("in application.HandleBegin lines: %q\n", lines)
}

func (a *App) HandleMiddle(apu repo.AppPieceUnit, voc repo.Vocabulaty) {
	lines, err := a.C.ParseMiddle(apu, voc)
	if err != nil {
		logger.L.Errorf("in application.HandleMiddle error: %v\n", err)
	}
	logger.L.Infof("in application.HandleMiddle lines: %q\n", lines)
}

func (a *App) HandleEnd(apu repo.AppPieceUnit, voc repo.Vocabulaty) {
	lines, err := a.C.ParseEnd(apu, voc)
	if err != nil {
		logger.L.Errorf("in application.HandleEnd error: %v\n", err)
	}
	logger.L.Infof("in application.HandleEnd lines: %q\n", lines)
}
