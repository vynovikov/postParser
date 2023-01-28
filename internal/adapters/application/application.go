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
	chanIn   chan repo.AppFeederUnit
	chanOut  chan []repo.AppDistributorUnit
	mpLock   sync.Mutex
	mpRWLock sync.RWMutex
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
	}
}

type Application interface {
	Do()
	AddToFeeder(repo.ReceiverUnit)
	//Compute(repo.DataPiece, repo.Presense, repo.Boundary) (repo.AppDistributorUnit, repo.StoreChange, error)
	HandleBuffer(repo.DataPiece, repo.Boundary) ([]repo.AppDistributorUnit, []error)
	//	CalcHeader(repo.DataPiece, repo.StoreChange) repo.AppDistributorHeader
	//CalcBody(repo.DataPiece, repo.Boundary) (repo.AppDistributorBody, error)
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

			//go a.HandleOrdered(&b, afu.R.H.Voc.Boundary, i)
			go a.HandleSmart(&b, afu.R.H.Voc.Boundary, i)
		}

		a.S.Inc(askg, len(m))
		for j := range m {
			//go a.HandleOrdered(&m[j], afu.R.H.Voc.Boundary, i)
			go a.HandleSmart(&m[j], afu.R.H.Voc.Boundary, i)
		}

		if !cmp.Equal(e, repo.AppSub{}) { // No uncertain sub piece
			a.S.Inc(askg, 1)
			go a.HandleSmart(&e, afu.R.H.Voc.Boundary, i)
			//go a.Handle(&e, afu.R.H.Voc.Boundary, i)
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

func (a *App) HandleSmart(d repo.DataPiece, bou repo.Boundary, i int) {
	logger.L.Infof("in application.HandleSmart worker %d called Handle with parameters dataPiece header %v, body %q, bou %v\n", i, d.GetHeader(), d.GetBody(0), bou)

	adus, _ := a.Handle(d, bou, i)

	//logger.L.Infof("in application.HandleUnOrdered worker %d issued adus:\n", i)
	/*	for j, v := range adus {
				logger.L.Warnf("in application.HandleSmart worker %d issued adu j = %d, header %v, body %q\n", i, j, v.H, v.B.B)
			}
			//logger.L.Infof("in application.HandleUnOrdered worker %d issued errors:\n", i)

		for j, v := range errs {
			logger.L.Infof("in application.HandleSmart worker %d issued error j = %d, error: %v\n", i, j, v)
		}
	*/
	//	logger.L.Warnf("in application.HandleSmart worker %d left counter = %d\n", i, a.S.Counter(repo.NewAppStoreKeyGeneralFromDataPiece(d)))
	a.toChanOut(adus)
}

func NewPieceKeyFromAPU(apu repo.AppPieceUnit) PieceKey {
	return PieceKey{
		Part: apu.APH.Part,
		TS:   apu.APH.TS,
	}
}

func (a *App) Handle(d repo.DataPiece, bou repo.Boundary, i int) ([]repo.AppDistributorUnit, []error) {
	_, out, prepErrs, _, p, isSub := repo.AppDistributorUnit{}, make([]repo.AppDistributorUnit, 0), make([]error, 0), repo.Unordered, d.Part(), d.IsSub()
	logger.L.Infof("worker %d invoked application.Handle for dataPiece with Part = %d which is sub %t\n", i, p, isSub)

	if d.B() == repo.False && d.E() == repo.False { // for unary transmition

		adu := repo.NewAppDistributorUnitUnary(d, bou, repo.Message{})

		//logger.L.Infof("in application.Handle worker %d made out %v\n", i, out)

		a.A.mpRWLock.Lock()
		defer a.A.mpRWLock.Unlock()

		switch a.S.Dec(d) {
		case repo.FirstAndLast:
			adu.H.U.M.PreAction = repo.Start
			adu.H.U.M.PostAction = repo.Finish
		case repo.First:
			adu.H.U.M.PreAction = repo.Start
		case repo.Last:
			adu.H.U.M.PostAction = repo.Finish
		}

		out = append(out, adu)

		logger.L.Infof("in application.Handle worker %d for dataPiece with header %v, body %q ==> made out %v\n", i, d.GetHeader(), d.GetBody(0), out)

		return out, prepErrs

	}

	// for stream transmition

	adub, header, bErr := CalcBody(d, bou)
	//logger.L.Infof("in application.Handle 1 adub %q, header: %q\n", adub.B, header)
	if bErr != nil {
		//logger.L.Errorf("in application.Handle CalcBody returned err: %v\n", bErr)
		prepErrs = append(prepErrs, bErr)
	}
	a.A.mpRWLock.RLock()
	presense := a.S.Presense(d) //may be changed later
	a.A.mpRWLock.RUnlock()
	logger.L.Infof("in application.Handle worker %d for dataPiece with header %v, body %q ==> made 1 try presense.ASKG %t, presense.ASKD %t, presense.OB %t\n", i, d.GetHeader(), d.GetBody(0), presense.ASKG, presense.ASKD, presense.OB)

	if (!d.IsSub() &&
		(d.B() == repo.True && !presense.ASKD) ||
		(d.B() == repo.False && d.E() == repo.Probably && !presense.OB)) ||
		(d.IsSub() &&
			!presense.OB) { //add case for AppSub with changing presense

		a.A.mpRWLock.Lock()
		presense = a.S.Presense(d)
		logger.L.Infof("in application.Handle worker %d for dataPiece with header %v, body %q ==> made 2 try presense.ASKG %t, presense.ASKD %t, presense.OB %t\n", i, d.GetHeader(), d.GetBody(0), presense.ASKG, presense.ASKD, presense.OB)

		sc, scErr := repo.NewStoreChange(d, presense, bou)
		logger.L.Infof("in application.Handle for dataPiece with header %v, body %q ==> made sc %v, scRrr: %v\n", d.GetHeader(), d.GetBody(0), sc, scErr)

		if (bErr == nil && scErr != nil) ||
			(bErr != nil && scErr != nil && scErr.Error() != bErr.Error()) {
			prepErrs = append(prepErrs, scErr)
		}

		defer a.A.mpRWLock.Unlock()
		return a.doHandle(d, sc, adub, header, bou, prepErrs)

	}
	sc, scErr := repo.NewStoreChange(d, presense, bou)
	//logger.L.Infof("in application.Handle sc %v, err: %v\n", sc, scErr)
	logger.L.Infof("in application.Handle worker %d for dataPiece with header %v, body %q ==> made sc %v, scRrr: %v\n", i, d.GetHeader(), d.GetBody(0), sc, scErr)

	if (bErr == nil && scErr != nil) ||
		(bErr != nil && scErr != nil && scErr.Error() != bErr.Error()) {
		//logger.L.Errorf("in application.Handle NewStoreChange returned err: %v\n", bErr)
		prepErrs = append(prepErrs, scErr)
	}

	a.A.mpRWLock.Lock()
	defer a.A.mpRWLock.Unlock()

	return a.doHandle(d, sc, adub, header, bou, prepErrs)
}

// returns adus and errors based on dataPiece and store states
func (a *App) doHandle(d repo.DataPiece, sc repo.StoreChange, adub repo.AppDistributorBody, header []byte, bou repo.Boundary, prepErrs []error) ([]repo.AppDistributorUnit, []error) {

	adus, errs, o := make([]repo.AppDistributorUnit, 0), prepErrs, repo.Unordered
	p := d.Part()

	logger.L.Infof("in application.doHandle handling p = %d\n", p)

	if sc.A != repo.Buffer {

		o = a.S.Dec(d)
		logger.L.Infof("in application.doHandle for apu with header %v got order = %d\n", d.GetHeader(), o)
	}

	a.S.Act(d, sc)

	if len(sc.From[repo.NewAppStoreKeyDetailed(d)]) == 2 && len(header) == 0 { // dataPiece after forked askd
		adub.B = append(sc.From[repo.NewAppStoreKeyDetailed(d)][true].D.H, adub.B...)
	}

	logger.L.Infof("in application.doHandle len(adub.B) = %d\n", len(adub.B))

	if !d.IsSub() &&
		(sc.A != repo.Buffer && len(adub.B) != 0) { // dataPieces with matched parts

		aduh := CalcHeader(d, sc, o)
		adu := repo.NewAppDistributorUnit(aduh, adub)
		logger.L.Infof("in application.doHandle for apu with header %v got adu header: %v, body: %q\n", d.GetHeader(), adu.H, adu.B.B)
		adus = append(adus, adu)
	}

	//logger.L.Infof("in application.doHandle sc = %v, IsPartChanged(sc) %t\n", sc, repo.IsPartChanged(sc))
	if repo.IsPartChanged(sc) {
		//logger.L.Infoln("in application.doHandle got into buffer")
		adusFromBuffer, errsFromBuffer := a.HandleBuffer(d, bou)
		adus = append(adus, adusFromBuffer...)
		errs = append(errs, errsFromBuffer...)
	}

	return adus, errs
}

func CalcBody(d repo.DataPiece, bou repo.Boundary) (repo.AppDistributorBody, []byte, error) {
	var err error
	b := d.GetBody(0)
	adub, header := repo.AppDistributorBody{}, make([]byte, 0)
	if d.IsSub() {
		return adub, b, nil
	}
	header, err = d.H(bou)
	logger.L.Infof("in application.CalcHeader header %q, err %v\n", header, err)
	if err != nil {
		if !strings.Contains(err.Error(), "is ending part") &&
			!strings.Contains(err.Error(), "no header found") {
			return repo.AppDistributorBody{}, d.GetBody(repo.MaxHeaderLimit), err
		}
	}

	adub = repo.AppDistributorBody{B: b}

	if len(header) > 0 && len(header) < len(b) {
		adub = repo.AppDistributorBody{B: d.GetBody(0)[len(header):]}
		return adub, header, nil
	}
	//logger.L.Infof("in application.CalcHeader body %q\n", d.GetBody(0))

	return adub, header, nil
}

func CalcHeader(d repo.DataPiece, sc repo.StoreChange, o repo.Order) repo.AppDistributorHeader {
	aduh, askd, fo, fi, f, pre, post := repo.AppDistributorHeader{}, repo.NewAppStoreKeyDetailed(d), "", "", repo.FiFo{}, repo.None, repo.None
	if d.IsSub() {
		return aduh
	}
	//logger.L.Infof("in application.CalcHeader from dataPiece header %v, body %q, sc %v, o: %v\n", d.GetHeader(), d.GetBody(0), sc, o)
	//logger.L.Infof("in application.CalcHeader d header %v, sc.To: %v, askd: %v\n", d.GetHeader(), sc.To, askd)
	askd = askd.IncPart()
	//fo, fi := sc.To[askd][false].D.FormName, sc.To[askd][false].D.FileName
	for i := range sc.To {
		fo, fi = sc.To[i][false].D.FormName, sc.To[i][false].D.FileName
		break
	}
	f = repo.NewFiFo(fo, fi)

	if sc.A == repo.Change && d.B() == repo.False && o == repo.First {

		pre = repo.Start
	}

	if sc.A == repo.Change && d.B() == repo.False && o != repo.First {

		pre = repo.Open
	}

	if sc.A == repo.Change && d.B() == repo.True {

		pre = repo.Continue
	}

	if sc.A == repo.Change && d.E() == repo.False && o == repo.Last {

		post = repo.Finish
	}

	if sc.A == repo.Change && d.E() == repo.False && o != repo.Last {

		post = repo.Close
	}
	if sc.A == repo.Change && d.E() != repo.False {

		post = repo.Continue

	}

	m := repo.NewMessage(pre, post)
	//logger.L.Infof("in application.CalcHeader m: %v\n", m)
	aduh = repo.NewAppDistributorHeaderStream(d, f, m)

	return aduh
}

func (a *App) HandleBuffer(d repo.DataPiece, bou repo.Boundary) ([]repo.AppDistributorUnit, []error) {
	return a.S.RegisterBuffer(d, bou)
}
