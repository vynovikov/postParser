package application

import (
	grpc "postParser/internal/adapters/driven/grpc"
	"postParser/internal/adapters/driven/store"
	"postParser/internal/core"
	"postParser/internal/logger"
	"postParser/internal/repo"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
)

type AppService struct {
	chanIn  chan repo.AppFeederUnit
	chanOut chan []repo.AppDistributorUnit
	buf     []repo.AppPieceUnit
	wg      sync.WaitGroup
	mp      map[PieceKey][]repo.AppPieceUnit
	mpLock  sync.Mutex
}

type PieceKey struct {
	TS   string
	Part int
}

/*
	func NewBlockInfo() *repo.BlockInfo {
		return &repo.BlockInfo{
			TS:       time.Now().Local(),
			Boundary: "---xxx",
		}
	}
*/
type App struct {
	T grpc.Transmitter
	C core.Parser
	A AppService
	S store.Store
}

func NewApplication(c core.Parser, s store.Store, t grpc.Transmitter) *App {

	App := &App{
		T: t,
		C: c,
		A: NewAppService(),
		S: s,
	}

	App.Do()
	return App
}

func NewAppService() AppService {
	return AppService{
		chanIn:  make(chan repo.AppFeederUnit, 10),
		chanOut: make(chan []repo.AppDistributorUnit, 10),
		mp:      make(map[PieceKey][]repo.AppPieceUnit, 0),
		wg:      sync.WaitGroup{},
	}
}

type Application interface {
	Do()
	AddToFeeder(repo.ReceiverUnit)
}

func (a *App) Do() {
	go a.Work(1)
	go a.Work(2)
	go a.Work(3)
	go a.Work(4)

	go a.Send()

}

func (a *App) toChanIn(afu repo.AppFeederUnit) {

	a.A.chanIn <- afu
}

func (a *App) toChanOut(adu []repo.AppDistributorUnit) {
	//logger.L.Infof("in application.toChanOut adding to chanOUT: part %v\n", afu.H)

	a.A.chanOut <- adu

}

func (a *App) Work(i int) {
	logger.L.Infof("in application.Work worker %d started\n", i)

	for afu := range a.A.chanIn {

		if len(afu.R.B.B) == 0 {
			continue
		}
		askg := repo.AppStoreKeyGeneral{}
		//logger.L.Infof("in application.Work worker %d got afu with part = %d, body: %q\n", i, afu.R.H.Part, afu.R.B.B)

		//Slicing feederUnit bytes chunk corrsesponding to boundary appearance
		b, m, e := a.C.Slicer(afu)

		if b.TS() != "" {
			askg.TS = b.TS()
		} else if len(m) > 0 {
			askg.TS = m[0].TS()
		}

		if !cmp.Equal(b, repo.AppPieceUnit{}) { // Beginning piece is empty only in zero part, leaving it

			a.S.Inc(askg, 1)

			go a.HandleOrdered(&b, afu.R.H.Voc.Boundary, i)
		}

		a.S.Inc(askg, len(m))
		for j := range m {
			go a.HandleOrdered(&m[j], afu.R.H.Voc.Boundary, i)
		}

		if !cmp.Equal(e, repo.AppSub{}) { // No uncertain sub piece
			a.S.Inc(askg, 1)
			go a.HandleOrdered(&e, afu.R.H.Voc.Boundary, i)
		}
		if afu.R.S.Signal == "EOF" {
			logger.L.Errorf("in application.Work EOF\n")
			//todo send stream stop message
		}

	}

}

func (a *App) AddToFeeder(in repo.ReceiverUnit) {

	A := repo.NewAppFeederUnit(in)
	a.A.chanIn <- A

}

func (a *App) Send() {
	for adu := range a.A.chanOut {

		//	logger.L.Infof("sending to rpc %v\n part %d\n", adu.H.Name, adu.H.TS)

		a.T.Transmit(adu)
	}
}

func (a *App) doHandle(d repo.DataPiece, bou repo.Boundary, i int) ([]repo.AppDistributorUnit, []error) {

	return Execute(a.S, d, bou)

}

func (a *App) HandleUnOrdered(d repo.DataPiece, bou repo.Boundary, i int) {
	logger.L.Infof("in application.HandleUnOrdered worker %d called doHandle with parameters dataPiece header %v, body %q, bou %v\n", i, d.GetHeader(), d.GetBody(0), bou)

	adus, errs := a.doHandle(d, bou, i)

	//logger.L.Infof("in application.HandleUnOrdered worker %d issued adus:\n", i)
	for j, v := range adus {
		logger.L.Infof("in application.HandleUnOrdered worker %d issued adu j = %d, header %v, body %q\n", i, j, v.H, v.B.B)
	}
	//logger.L.Infof("in application.HandleUnOrdered worker %d issued errors:\n", i)
	for j, v := range errs {
		logger.L.Infof("in application.HandleUnOrdered worker %d issued error j = %d, error: %v\n", i, j, v)
	}
	logger.L.Infof("in application.HandleUnOrdered worker %d left counter = %d\n", i, a.S.Counter(repo.NewAppStoreKeyGeneral(d)))
	a.toChanOut(adus)
}

func (a *App) HandleOrdered(d repo.DataPiece, bou repo.Boundary, i int) {
	a.A.mpLock.Lock()
	defer a.A.mpLock.Unlock()
	logger.L.Infof("in application.HandleOrdered worker %d called doHandle with parameters dataPiece header %v, body %q, bou %q\n", i, d.GetHeader(), d.GetBody(0), bou)

	adus, errs := a.doHandle(d, bou, i)
	/*if errE != nil {
		for _, v := range errE {
			logger.L.Errorf("in application.HandleOrdered worker %d made err: %v\n", i, v)
		}

	}
	*/
	//logger.L.Infof("in application.HandleOrdered worker %d issued adus:\n", i)
	if len(adus) > 0 {
		for j, v := range adus {
			logger.L.Infof("in application.HandleOrdered worker %d issued adu %d) header %v, body %q\n", i, j, v.H, v.B.B)
		}
		a.toChanOut(adus)
	}

	//logger.L.Infof("in application.HandleOrdered worker %d issued errors:\n", i)
	for j, v := range errs {
		logger.L.Errorf("in application.HandleOrdered worker %d issued error %d) %v\n", i, j, v)

	}
	logger.L.Infof("in application.HandleOrdered worker %d left counter = %d\n", i, a.S.Counter(repo.NewAppStoreKeyGeneral(d)))

}
func NewPieceKeyFromAPU(apu repo.AppPieceUnit) PieceKey {
	return PieceKey{
		Part: apu.APH.Part,
		TS:   apu.APH.TS,
	}
}

// Making decisions based on AppPieceUnit parameters
func Decide(d repo.DataPiece) []string {
	decision := make([]string, 0)

	if !d.B() && d.E() == repo.False { //no handling needed just unary send
		decision = append(decision, "unary")
		return decision
	}
	decision = append(decision, "register")

	decision = append(decision, "register buffer")

	return decision
}

func Execute(s store.Store, d repo.DataPiece, bou repo.Boundary) ([]repo.AppDistributorUnit, []error) {
	askg := repo.NewAppStoreKeyGeneral(d)
	//	logger.L.Infof("application.Execute invoked with d header: %v, s.C = %d\n", d.GetHeader(), s.Counter(askg))
	out, errs := make([]repo.AppDistributorUnit, 0), make([]error, 0)
	if !d.B() && d.E() == repo.False { //no handling needed just unary send no message

		s.Dec(askg)
		//logger.L.Errorf("in application.Execute dataPiece with body %q decrementing and counter = %d\n", d.GetBody(0), s.Counter(askg))
		out = append(out, repo.NewAppDistributorUnitUnary(d, bou, repo.Message{}))
		if s.Counter(askg) >= 2 {
			return out, errs
		}
		if s.Counter(askg) == 1 {
			adusFromBuffer, errsReg := s.RegisterBuffer(bou)
			out = append(out, adusFromBuffer...)
			errs = append(errs, errsReg...)

			return out, errs
		}

	}
	if !d.B() && d.E() == repo.Last { //no handling needed just unary send with message "last". Need to register
		//logger.L.Infof("in application.Execute dataPiece with body %q is last and counter = %d\n", d.GetBody(0), s.Counter(askg))
		if s.Counter(askg) == 1 {

			adu := repo.NewAppDistributorUnitUnary(d, bou, repo.Message{PostAction: repo.Finish})
			//logger.L.Infof("in application.Execute dataPiece with body %q became the last and made adu with header = %v, body: %q\n", d.GetBody(0), adu.H, adu.B.B)
			out = append(out, adu)
			s.Dec(askg)
			//logger.L.Warnf("in application.Execute for Last piece counter == 1 and after dec c = %d\n", s.Counter())

			return out, errs
		}
		adu, err := s.Register(d, bou)
		//logger.L.Infof("in application.Execute dataPiece with body %q is last and counter = %d have been registered with error %v\n", d.GetBody(0), s.Counter(askg), err)
		if err != nil {
			errs = append(errs, err)
		}
		if !cmp.Equal(adu, repo.AppDistributorUnit{}) {
			out = append(out, adu)
		}
		return out, errs

	}
	// stream piece handling
	adu, err := s.Register(d, bou)
	//logger.L.Infof("in application.Execute Register returned adu: header: %v, body: %q, error: %v, s.C: %v\n", adu.H, adu.B.B, err, s.Counter(askg))
	if !cmp.Equal(adu, repo.AppDistributorUnit{}) ||
		(err != nil &&
			cmp.Equal(adu, repo.AppDistributorUnit{}) &&
			strings.Contains(err.Error(), "double-meaning")) ||
		(err != nil &&
			cmp.Equal(adu, repo.AppDistributorUnit{}) &&
			strings.Contains(err.Error(), "is not full")) {

		if !cmp.Equal(adu, repo.AppDistributorUnit{}) {
			out = append(out, adu)
		}

		s.Dec(askg)
		//logger.L.Errorf("in application.Execute because of %q counter decremented and became %d", adu.B.B, s.Counter(askg))
	}
	if err != nil {
		//logger.L.Errorf("in application.Execute Register returned error: %v, s.Counter(askg) = %d\n", err, s.Counter(askg))
		errs = append(errs, err)
		if (strings.Contains(err.Error(), "finished") && s.Counter(askg) > 1) ||
			strings.Contains(err.Error(), "unknown") ||
			strings.Contains(err.Error(), "added to buffer") ||
			strings.Contains(err.Error(), "but got") {

			//logger.L.Warnf("in application.Execute err.Error %q caused return\n", err.Error())
			return out, errs

		}

		//logger.L.Warnln("in application.Execute checking buffer")

		//logger.L.Errorf("in application.Execute dataPiece with body %q decrementing and counter = %d\n", d.GetBody(0), s.Counter(askg))
		adusFromBuffer, errsReg := s.RegisterBuffer(bou)
		//test not adding empty structs

		if len(adusFromBuffer) > 0 {
			for _, v := range adusFromBuffer {
				if !cmp.Equal(v, repo.AppDistributorUnit{}) ||
					(err != nil &&
						cmp.Equal(v, repo.AppDistributorUnit{}) &&
						strings.Contains(err.Error(), "double-meaning")) {

					if !cmp.Equal(v, repo.AppDistributorUnit{}) {
						out = append(out, v)
					}

					//s.Dec(askg)
					//logger.L.Errorf("in application.Execute because of %q counter decremented and became %d", adu.B.B, s.Counter(askg))
				}
			}

		}

		if len(errsReg) > 0 {
			for _, v := range errsReg {
				if v != nil {
					errs = append(errs, v)
				}
			}

		}

		return out, errs

	}
	//logger.L.Warnln("in application.Execute checking buffer")
	//logger.L.Warnf("in application.Execute dataPiece with body %q decrementing and counter = %d\n", d.GetBody(0), s.Counter(askg))
	adusFromBuffer, errsReg := s.RegisterBuffer(bou)
	//logger.L.Infoln("in application.Execute adusFromBuffer: ")
	/*
		for i, v := range adusFromBuffer {
			logger.L.Infof("in application.Execute %d ---- %v===%q\n", i, v.H, v.B.B)
		}
	*/
	if len(adusFromBuffer) > 0 {
		for _, v := range adusFromBuffer {
			if !cmp.Equal(v, repo.AppDistributorUnit{}) {
				out = append(out, v)
				//s.Dec(askg)
			}
		}

	}

	if len(errsReg) > 0 {
		for _, v := range errsReg {
			if v != nil {
				errs = append(errs, v)
			}
		}

	}
	//logger.L.Infof("in application.Execute after all counter = %d\n", s.Counter(askg))
	//logger.L.Infoln("in application.Execute after all out")
	/*
		for i, v := range out {
			logger.L.Infof("in application.Execute out %d has header %v, body %q\n", i, v.H, v.B)
		}
	*/
	return out, errs

}
