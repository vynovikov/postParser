package application

import (
	grpc "postParser/internal/adapters/driven/grpc"
	"postParser/internal/core"
	"postParser/internal/repo"
	"time"
)

type AppService struct {
	chanIn  chan repo.AppFeederUnit
	chanOut chan repo.AppDistributorUnit
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

		//logger.L.Infof("in application.Work working with %v\n", afu)

		a.C.ParseEnd(afu)
		a.C.ParseBegin(afu)

		//logger.L.Infof("in applicatiom.Work made header %v\n", afu.H)

		ao := newAppDistributorUnit(afu)

		//logger.L.Infof("work done. Made %v\n with part %d\n", ao.H, afu.R.H.Part)

		a.toChanOut(ao)
	}

}

func (a *App) AddToFeeder(in repo.ReceiverUnit) {

	h := repo.NewAppFeederHeader(&repo.SepHeader{}, &repo.SepBody{}, in.H.Part)

	A := newAppFeederUnit(h, in)

	//logger.L.Infof("in application.AddToHeader receive header is %v\n", A.R.H)

	if (in.S == repo.ReceiverSignal{}) {
		//	logger.L.Infof("in application.AddToFeeder sending afu with header %v with sepHeader %v\n", A.H, A.H.SepHeader)
		a.A.chanIn <- A
	}

}

func (a *App) Send() {
	for adu := range a.A.chanOut {

		//	logger.L.Infof("sending to rpc %v\n part %d\n", adu.H.Name, adu.H.TS)

		a.T.Transmit(adu)
	}
}

func newAppFeederUnit(h *repo.AppFeederHeader, u repo.ReceiverUnit) repo.AppFeederUnit {
	return repo.AppFeederUnit{
		H: h,
		R: u,
	}
}

func newAppHeader() *repo.AppFeederHeader {
	return &repo.AppFeederHeader{}
}

func newAppDistributorUnit(afu repo.AppFeederUnit) repo.AppDistributorUnit {
	return repo.AppDistributorUnit{
		B: repo.DistributorBody(afu.R.B),
		H: repo.DistributorHeader{
			Name: "Mashka",
			TS:   afu.R.H.TS,
		},
	}
}
