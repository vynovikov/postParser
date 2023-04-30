// Application layer. Central place through witch all adapters interact
package application

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vynovikov/postParser/internal/adapters/driven/rpc"
	"github.com/vynovikov/postParser/internal/adapters/driven/store"
	"github.com/vynovikov/postParser/internal/logger"
	"github.com/vynovikov/postParser/internal/repo"

	"github.com/google/go-cmp/cmp"
)

type AppService struct {
	stopping        bool
	chanInClosed    bool
	transmitterLock sync.Mutex
	appRWLock       sync.RWMutex

	W repo.WaitGroups
	C repo.Channels
}

type PieceKey struct {
	TS   string
	Part int
}

// Spy logger interface for testing
type Logger interface {
	LogStuff(repo.AppUnit)
}

// All adapters combined
type App struct {
	T rpc.Transmitter
	A AppService
	S store.Store
	L Logger // Logger interface is testDouble spy for testing Handle method
}

func NewAppFull(s store.Store, t rpc.Transmitter) (*App, chan struct{}) {
	done := make(chan struct{})

	App := &App{
		T: t,
		A: NewAppService(done),
		S: s,
	}
	return App, done
}

func NewAppEmpty() *App {

	App := &App{
		A: NewAppService(make(chan struct{})),
	}
	return App
}

func NewAppService(done chan struct{}) AppService {
	return AppService{
		W: repo.WaitGroups{
			M: make(map[repo.AppStoreKeyGeneral]*sync.WaitGroup),
		},
		C: repo.Channels{
			ChanIn:  make(chan repo.AppFeederUnit, 10),
			ChanOut: make(chan repo.AppDistributorUnit, 10),
			ChanLog: make(chan string, 10),
			Done:    done,
		},
	}
}

func (a *App) MountLogger(l Logger) {
	a.L = l
}

type Application interface {
	Start()
	AddToFeeder(repo.ReceiverUnit)
	HandleBuffer(repo.AppStoreKeyGeneral, repo.Boundary) ([]repo.AppDistributorUnit, []error)
	SetStopping()
	Stopping() bool
	Stop()
	ChainInClose()
}

func (a *App) Start() {

	a.A.W.Workers.Add(4)
	go a.Work(1)
	go a.Work(2)
	go a.Work(3)
	go a.Work(4)

	a.A.W.Sender.Add(1)
	go a.Send()

	go a.Log()

}

func (a *App) toChanOut(adu repo.AppDistributorUnit) {

	if a.L != nil {
		a.L.LogStuff(adu)
	}
	a.A.C.ChanOut <- adu

}

func (a *App) toChanLog(s string) {
	a.A.C.ChanLog <- s
}

// Work is the central function of whole application.
// Is runiing in four instanses in concurrent way.
// Handles data from receiver, sends results to transmitter
func (a *App) Work(i int) {
	logger.L.Infof("in postparser.application.Work worker %d started\n", i)

	for afu := range a.A.C.ChanIn {
		if len(afu.R.B.B) == 0 {
			continue
		}
		askg := repo.NewAppStoreKeyGeneralFromFeeder(afu)
		w := a.A.W.M[askg]

		// Reading feederUnit bytes chunk, finding to boundary appearance and slicing it into dataPieces
		b, m, e := repo.Slicer(afu.R.B.B, afu.R.H.Bou)

		if !cmp.Equal(b, repo.AppPieceUnit{}) { // Beginning piece is empty only in zero part, leaving it

			b.APH.SetPart(afu.R.H.Part)
			b.APH.SetTS(afu.R.H.TS)

			a.S.Inc(askg, 1)

			w.Add(1)
			go a.Handle(&b, afu.R.H.Bou, w, i)
		}

		a.S.Inc(askg, len(m))
		w.Add(len(m))

		for j := range m {
			m[j].APH.SetPart(afu.R.H.Part)
			m[j].APH.SetTS(afu.R.H.TS)
			go a.Handle(&m[j], afu.R.H.Bou, w, i)
		}

		if !cmp.Equal(e, repo.AppSub{}) { // ending piece isuncertain sub piece
			e.SetPart(afu.R.H.Part)
			e.SetTS(afu.R.H.TS)

			a.S.Inc(askg, 1)
			w.Add(1)
			go a.Handle(&e, afu.R.H.Bou, w, i)

		}

		if afu.R.H.Unblock {
			w.Wait()
			a.S.Unblock(askg)
			a.A.appRWLock.Lock()
			adus, _ := a.HandleBuffer(repo.NewAppStoreKeyGeneralFromFeeder(afu), afu.R.H.Bou)
			a.A.appRWLock.Unlock()

			for _, v := range adus {

				a.toChanLog(fmt.Sprintf("in application.Work extracted from buffer adu header: %v, body %q", v.H, v.B.B))
				a.toChanOut(v)
			}
		}

	}
	a.A.W.Workers.Done()

}

// AddToFeeder updates receiver data and sends it to chanIn
func (a *App) AddToFeeder(in repo.ReceiverUnit) {
	if a.L != nil {
		a.L.LogStuff(in)
	}

	A := repo.NewAppFeederUnit(in)

	askg := repo.NewAppStoreKeyGeneralFromFeeder(A)

	if _, ok := a.A.W.M[askg]; !ok {
		a.A.W.M[askg] = &sync.WaitGroup{}
	}

	a.A.C.ChanIn <- A
}

// Send is running as gourutine. Initiates transmission for any data got from chanOut
func (a *App) Send() {

	for adu := range a.A.C.ChanOut {
		a.A.transmitterLock.Lock()
		go a.T.Transmit(adu, &a.A.transmitterLock)
	}
	a.A.W.Sender.Done()
}

func (a *App) Log() {
	for l := range a.A.C.ChanLog {
		go a.T.Log(l)
	}
}

func NewPieceKeyFromAPU(apu repo.AppPieceUnit) PieceKey {
	return PieceKey{
		Part: apu.APH.Part,
		TS:   apu.APH.TS,
	}
}

// Handle runs as individual goroutine in concurrent way.
// Handles dataPieces depending on its parameters and state of store.
// Tested in application_test.go
func (a *App) Handle(d repo.DataPiece, bou repo.Boundary, w *sync.WaitGroup, i int) {
	prepErrs := make([]error, 0)
	a.toChanLog(fmt.Sprintf("in postparser worker %d invoked application.Handle for dataPiece with header %v, body %q", i, d.GetHeader(), d.GetBody(0)))

	if d.B() == repo.False && d.E() == repo.False { // for unary transmition

		adu := repo.NewAppDistributorUnitUnary(d, bou, repo.Message{})
		a.A.appRWLock.Lock()

		o, err := a.S.Dec(d)

		switch o {
		case repo.FirstAndLast:
			adu.H.U.M.PreAction = repo.Start
			adu.H.U.M.PostAction = repo.Finish
		case repo.First:
			adu.H.U.M.PreAction = repo.Start
		case repo.Last:
			adu.H.U.M.PostAction = repo.Finish
		case repo.Unordered:
			if err != nil && strings.Contains(err.Error(), "further") {
				a.S.BufferAdd(d)
				a.A.appRWLock.Unlock()
				w.Done()
				return
			}
		}

		a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made adu %v", i, d.GetHeader(), d.GetBody(0), adu))

		a.toChanOut(adu)
		a.A.appRWLock.Unlock()
		w.Done()
		return
	}

	// for stream transmition

	adub, header, bErr := CalcBody(d, bou)
	if bErr != nil {
		prepErrs = append(prepErrs, bErr)
	}
	a.A.appRWLock.RLock()
	presence, err := a.S.Presence(d) //may be changed later
	a.A.appRWLock.RUnlock()
	a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made 1 try presense.ASKG %t, presense.ASKD %t, presense.OB %t", i, d.GetHeader(), d.GetBody(0), presence.ASKG, presence.ASKD, presence.OB))
	if err != nil {
		prepErrs = append(prepErrs, err)
	}
	if (!d.IsSub() &&
		(d.B() == repo.True && !presence.ASKD) ||
		(d.B() == repo.False && d.E() == repo.Probably && !presence.OB)) ||
		(d.IsSub() &&
			!presence.OB) { //add case for AppSub with changing presense

		a.A.appRWLock.Lock()
		presence, err = a.S.Presence(d)
		if err != nil {
			prepErrs = append(prepErrs, err)
		}

		if err != nil {
			prepErrs = append(prepErrs, err)
		}
		a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made 2 try presense.ASKG %t, presense.ASKD %t, presense.OB %t", i, d.GetHeader(), d.GetBody(0), presence.ASKG, presence.ASKD, presence.OB))

		sc, scErr := repo.NewStoreChange(d, presence, bou)
		a.toChanLog(fmt.Sprintf("in postparser.application.Handle for dataPiece with header %v, body %q ==> made sc %v, scRrr: %v", d.GetHeader(), d.GetBody(0), sc, scErr))

		if (bErr == nil && scErr != nil) ||
			(bErr != nil && scErr != nil && scErr.Error() != bErr.Error()) {
			prepErrs = append(prepErrs, scErr)
		}

		adus, _ := a.doHandle(d, sc, adub, header, bou, prepErrs)
		for _, v := range adus {
			a.toChanOut(v)
		}
		a.A.appRWLock.Unlock()
		w.Done()
		return

	}
	sc, scErr := repo.NewStoreChange(d, presence, bou)
	a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made sc %v, scRrr: %v", i, d.GetHeader(), d.GetBody(0), sc, scErr))

	if (bErr == nil && scErr != nil) ||
		(bErr != nil && scErr != nil && scErr.Error() != bErr.Error()) {
		prepErrs = append(prepErrs, scErr)
	}

	a.A.appRWLock.Lock()

	adus, _ := a.doHandle(d, sc, adub, header, bou, prepErrs)
	for _, v := range adus {
		a.toChanOut(v)
	}
	a.A.appRWLock.Unlock()

	w.Done()
}

// Helper function for Handle. Always invoked when appRWLock is locked
func (a *App) doHandle(d repo.DataPiece, sc repo.StoreChange, adub repo.AppDistributorBody, header []byte, bou repo.Boundary, prepErrs []error) ([]repo.AppDistributorUnit, []error) {
	var decrementErr error
	adus, errs, o := make([]repo.AppDistributorUnit, 0), prepErrs, repo.Unordered
	if sc.A != repo.Buffer {

		o, decrementErr = a.S.Dec(d)

		if decrementErr != nil {
			sc.A = repo.Buffer
		}
		a.toChanLog(fmt.Sprintf("in postparser.application.doHandle for apu with header %v got o = %d, decrementErr: %v", d.GetHeader(), o, decrementErr))

	}

	a.S.Act(d, sc)

	if len(sc.From[repo.NewAppStoreKeyDetailed(d)]) == 2 && len(header) == 0 { // dataPiece after forked askd
		adub.B = append(sc.From[repo.NewAppStoreKeyDetailed(d)][true].D.H, adub.B...)
	}

	if !d.IsSub() &&
		(sc.A != repo.Buffer && len(adub.B) != 0) { // dataPieces with matched parts

		aduh := CalcHeader(d, sc, o)
		adu := repo.NewAppDistributorUnit(aduh, adub)

		a.toChanLog(fmt.Sprintf("in postparser.application.doHandle for apu with header %v got adu header: %v, body: %q", d.GetHeader(), adu.H, adu.B.B))
		adus = append(adus, adu)
	}
	if repo.IsPartChanged(sc) {

		adusFromBuffer, errsFromBuffer := a.HandleBuffer(repo.NewAppStoreKeyGeneralFromDataPiece(d), bou)
		for _, v := range adusFromBuffer {
			a.toChanLog(fmt.Sprintf("in postparser.application.doHandle apu with header %v extracted from buffer adu with header: %v", d.GetHeader(), v.H))
		}

		adus = append(adus, adusFromBuffer...)
		errs = append(errs, errsFromBuffer...)
	}

	return adus, errs
}

// CalcBody creates body of unit to be transfered. Tested in application_test.go
func CalcBody(d repo.DataPiece, bou repo.Boundary) (repo.AppDistributorBody, []byte, error) {
	var err error
	b := d.GetBody(0)
	adub, header := repo.AppDistributorBody{}, make([]byte, 0)
	if d.IsSub() {
		return adub, b, nil
	}
	header, err = d.H(bou)

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

	return adub, header, nil
}

// CalcHeader creates header of unit to be transfered. Tested in application_test.go
func CalcHeader(d repo.DataPiece, sc repo.StoreChange, o repo.Order) repo.AppDistributorHeader {
	aduh, askd, fo, fi, f, b, pre, post := repo.AppDistributorHeader{}, repo.NewAppStoreKeyDetailed(d), "", "", repo.FiFo{}, repo.BeginningData{}, repo.None, repo.None
	if d.IsSub() {
		return aduh
	}
	askd = askd.IncPart()
	for i := range sc.To {
		fo, fi = sc.To[i][false].D.FormName, sc.To[i][false].D.FileName
		b = sc.To[i][false].B
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
	aduh = repo.NewAppDistributorHeaderStream(d, f, m, b)

	return aduh
}

func (a *App) HandleBuffer(askg repo.AppStoreKeyGeneral, bou repo.Boundary) ([]repo.AppDistributorUnit, []error) {
	return a.S.RegisterBuffer(askg, bou)
}
func (a *App) SetStopping() {

	a.A.stopping = true
}

func (a *App) Stopping() bool {
	return a.A.stopping
}
func (a *App) ChainInClose() {

	if !a.A.chanInClosed {
		a.A.chanInClosed = true
		close(a.A.C.ChanIn)

	}
}

func (a *App) Stop() {

	a.A.W.Workers.Wait()
	close(a.A.C.ChanOut)
	a.toChanLog("postParser is down")
	close(a.A.C.ChanLog)
	a.A.W.Sender.Wait()
	close(a.A.C.Done)
}
