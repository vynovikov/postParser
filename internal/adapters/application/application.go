package application

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vynovikov/postParser/internal/adapters/driven/rpc"
	"github.com/vynovikov/postParser/internal/adapters/driven/store"
	"github.com/vynovikov/postParser/internal/core"
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

type Logger interface {
	LogStuff(repo.AppUnit)
}

type EmptyLogger struct{}

func NewEmptyLogger() EmptyLogger {
	return EmptyLogger{}
}

func (e *EmptyLogger) LogStuff(repo.AppUnit) {

}

type App struct {
	T rpc.Transmitter
	C core.Parser
	A AppService
	S store.Store
	L Logger
}

func NewAppFull(c core.Parser, s store.Store, t rpc.Transmitter) (*App, chan struct{}) {
	done := make(chan struct{})

	App := &App{
		T: t,
		C: c,
		A: NewAppService(done),
		S: s,
	}

	//App.Do()
	return App, done
}

func NewAppEmpty() *App {

	App := &App{
		A: NewAppService(make(chan struct{})),
	}

	//App.Do()
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
	Do()
	AddToFeeder(repo.ReceiverUnit)
	HandleBuffer(repo.AppStoreKeyGeneral, repo.Boundary) ([]repo.AppDistributorUnit, []error)
	SetStopping()
	Stopping() bool
	Stop()
	ChainInClose()
	//ChanInOpened() bool
}

func (a *App) Do() {

	a.A.W.Workers.Add(4)
	go a.Work(1)
	go a.Work(2)
	go a.Work(3)
	go a.Work(4)

	a.A.W.Sender.Add(1)
	go a.Send()

	go a.Log()

}

func (a *App) toChanIn(afu repo.AppFeederUnit) {

	askg := repo.NewAppStoreKeyGeneralFromFeeder(afu)

	if _, ok := a.A.W.M[askg]; !ok {
		a.A.W.M[askg] = &sync.WaitGroup{}
		//logger.L.Infof("in application.toChanIn A.W.M %v\n", a.A.W.M)
	}

	a.A.C.ChanIn <- afu
}

func (a *App) toChanOut(adu repo.AppDistributorUnit) {
	//logger.L.Infof("in application.toChanOut before a.A.C: %v\n", a.A.C)

	if a.L != nil {
		a.L.LogStuff(adu)
	}
	/*
		if adu.H.S.M.PreAction == repo.Open {
			logger.L.Infof("in application.toChanOut trying to send adu header %v body %q to chanOut\n", adu.H, adu.B.B)
		}
	*/
	a.A.C.ChanOut <- adu
	//logger.L.Infof("in application.toChanOut after a.A.C: %v\n", a.A.C)

}

func (a *App) toChanLog(s string) {
	//logger.L.Infof("in application.toChanLog trying to send %q to chanLog\n", s)
	a.A.C.ChanLog <- s
}

func (a *App) Work(i int) {
	logger.L.Infof("in postparser.application.Work worker %d started\n", i)

	for afu := range a.A.C.ChanIn {
		//logger.L.Infoln("slicing part", afu.R.H.Part)
		if len(afu.R.B.B) == 0 {
			continue
		}
		askg := repo.NewAppStoreKeyGeneralFromFeeder(afu)
		w := a.A.W.M[askg]

		//Slicing feederUnit bytes chunk corrsesponding to boundary appearance
		b, m, e := a.C.Slicer(afu)

		if !cmp.Equal(b, repo.AppPieceUnit{}) { // Beginning piece is empty only in zero part, leaving it

			a.S.Inc(askg, 1)

			w.Add(1)
			go a.Handle(&b, afu.R.H.Bou, w, i)
		}

		a.S.Inc(askg, len(m))
		w.Add(len(m))

		for j := range m {
			go a.Handle(&m[j], afu.R.H.Bou, w, i)
		}

		if !cmp.Equal(e, repo.AppSub{}) { // No uncertain sub piece
			a.S.Inc(askg, 1)
			w.Add(1)
			go a.Handle(&e, afu.R.H.Bou, w, i)

		}

		if afu.R.H.Unblock {
			//logger.L.Errorf("in application.Work worker %d got unblock\n", i)
			w.Wait()
			//logger.L.Errorln("in application.Work unblock action after waiting")
			a.S.Unblock(askg)
			//logger.L.Errorln("in application.Work releasing buffer")
			a.A.appRWLock.Lock()
			adus, _ := a.HandleBuffer(repo.NewAppStoreKeyGeneralFromFeeder(afu), afu.R.H.Bou)
			a.A.appRWLock.Unlock()
			//logger.L.Errorf("in application.Work after unblock extracted %d adus from buffer\n", len(adus))
			/*
				if len(errs) != 0 {
					for _, v := range errs {
						logger.L.Errorf("in application.Work err: %v\n", v)
					}
				}
			*/
			/*
				for i, v := range adus {
					logger.L.Errorf("in application.Work i = %d, adu header: %v, body %q\n", i, v.H, v.B.B)
				}
			*/

			for _, v := range adus {
				//logger.L.Errorf("in application.Work extracted from buffer adu header: %v, body %q\n", v.H, v.B.B)
				a.toChanLog(fmt.Sprintf("in application.Work extracted from buffer adu header: %v, body %q", v.H, v.B.B))
				a.toChanOut(v)
			}

			//a.S.Reset(askg)
		}

	}
	//logger.L.Errorf("in application.Work worker %d left chanIn loop\n", i)
	a.A.W.Workers.Done()

}

func (a *App) AddToFeeder(in repo.ReceiverUnit) {
	//logger.L.Infof("in application.AppToFeeder before a.A.C: %v\n", a.A.C)
	if a.L != nil {
		a.L.LogStuff(in)
	}

	A := repo.NewAppFeederUnit(in)

	askg := repo.NewAppStoreKeyGeneralFromFeeder(A)

	if _, ok := a.A.W.M[askg]; !ok {
		a.A.W.M[askg] = &sync.WaitGroup{}
		//logger.L.Infof("in application.toChanIn A.W.M %v\n", a.A.W.M)
	}
	//logger.L.Infof("in application.AddToFeeder A header %v, value %q\n", A.R.H, A.R.B.B)

	/*
		s := "sdkjfhsdjf"
		logger.L.Infof("in application.toChanIn sending to chanLog %q\n", s)
		a.A.C.ChanLog <- s
	*/
	a.A.C.ChanIn <- A
	//logger.L.Infof("in application.AppToFeeder after a.A.C: %v\n", a.A.C)
}

func (a *App) Send() {

	for adu := range a.A.C.ChanOut {
		//logger.L.Errorf("in application.Send before lock sending adu header %v\n", adu.H)
		a.A.transmitterLock.Lock()
		//logger.L.Errorf("in application.Send after lock sending adu header %v\n", adu.H)
		go a.T.Transmit(adu, &a.A.transmitterLock)
	}
	a.A.W.Sender.Done()
}

func (a *App) Log() {
	for l := range a.A.C.ChanLog {
		//logger.L.Infof("in application.Log got %q, trying to grpc it\n", l)
		go a.T.Log(l)
	}
}

/*
	func (a *App) HandleSmart(d repo.DataPiece, bou repo.Boundary, w *sync.WaitGroup, i int) {
		adus, _ := a.Handle(d, bou, w, i)
		a.toChanOut(adus)
	}
*/
func NewPieceKeyFromAPU(apu repo.AppPieceUnit) PieceKey {
	return PieceKey{
		Part: apu.APH.Part,
		TS:   apu.APH.TS,
	}
}

func (a *App) Handle(d repo.DataPiece, bou repo.Boundary, w *sync.WaitGroup, i int) {
	//defer w.Done()
	prepErrs := make([]error, 0)
	a.toChanLog(fmt.Sprintf("in postparser worker %d invoked application.Handle for dataPiece with header %v, body %q", i, d.GetHeader(), d.GetBody(0)))

	if d.B() == repo.False && d.E() == repo.False { // for unary transmition

		adu := repo.NewAppDistributorUnitUnary(d, bou, repo.Message{})

		//logger.L.Infof("in application.Handle worker %d made adu %v\n", i, adu)

		a.A.appRWLock.Lock()

		o, err := a.S.Dec(d)
		//a.toChanLog(fmt.Sprintf("in application.Handle worker %d for dataPiece with header %v, body %q made o %d, err %v", i, d.GetHeader(), d.GetBody(0), o, err))

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
				//logger.L.Infof("in application.Handle worker %d adds dataPiece header %v, body %q to buffer\n", i, adu.H, adu.B.B)
				a.S.BufferAdd(d)
				a.A.appRWLock.Unlock()
				w.Done()
				return
			}
		}

		//out = append(out, adu)

		a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made adu %v", i, d.GetHeader(), d.GetBody(0), adu))

		a.toChanOut(adu)

		//logger.L.Infof("in application.Handle worker %d for dataPiece with header %v, body %q ==> made adu %v\n", i, d.GetHeader(), d.GetBody(0), adu)

		//a.toChanLog(fmt.Sprintf("in application.Handle worker %d for dataPiece with header %v, body %q ==> made adu %v\n", i, d.GetHeader(), d.GetBody(0), adu))

		a.A.appRWLock.Unlock()
		w.Done()
		return
	}

	// for stream transmition

	adub, header, bErr := CalcBody(d, bou)
	//logger.L.Infof("in application.Handle 1 adub %q, header: %q\n", adub.B, header)
	if bErr != nil {
		//logger.L.Errorf("in application.Handle CalcBody returned err: %v\n", bErr)
		prepErrs = append(prepErrs, bErr)
	}
	a.A.appRWLock.RLock()
	presence, err := a.S.Presence(d) //may be changed later
	a.A.appRWLock.RUnlock()
	/*
		if d.B() == repo.False && d.Part() != 0 {
			logger.L.Infof("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made 1 try presense.ASKG %t, presense.ASKD %t, presense.OB %t\n", i, d.GetHeader(), d.GetBody(0), presence.ASKG, presence.ASKD, presence.OB)
		}
	*/
	a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made 1 try presense.ASKG %t, presense.ASKD %t, presense.OB %t", i, d.GetHeader(), d.GetBody(0), presence.ASKG, presence.ASKD, presence.OB))
	if err != nil {
		prepErrs = append(prepErrs, err)
	}
	/*
		if d.B() == repo.False ||
			(d.B() == repo.True &&
				presense.ASKD) {
			logger.L.Infof("in application.Handle worker %d for dataPiece with header %v ==> made 1 try presense.ASKG %t, presense.ASKD %t, presense.OB %t, err: %v\n", i, d.GetHeader(), presense.ASKG, presense.ASKD, presense.OB, err)
		} else {
			logger.L.Errorf("in application.Handle worker %d for dataPiece with header %v ==> made 1 try presense.ASKG %t, presense.ASKD %t, presense.OB %t, err: %v\n", i, d.GetHeader(), presense.ASKG, presense.ASKD, presense.OB, err)
		}
	*/
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
		/*
			if d.B() == repo.False ||
				(d.B() == repo.True &&
					presense.ASKD) {
				logger.L.Infof("in application.Handle worker %d for dataPiece with header %v ==> made 2 try presense.ASKG %t, presense.ASKD %t, presense.OB %t, err: %v\n", i, d.GetHeader(), presense.ASKG, presense.ASKD, presense.OB, err)
			} else {
				logger.L.Errorf("in application.Handle worker %d for dataPiece with header %v ==> made 2 try presense.ASKG %t, presense.ASKD %t, presense.OB %t, err: %v\n", i, d.GetHeader(), presense.ASKG, presense.ASKD, presense.OB, err)
			}
		*/
		a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made 2 try presense.ASKG %t, presense.ASKD %t, presense.OB %t", i, d.GetHeader(), d.GetBody(0), presence.ASKG, presence.ASKD, presence.OB))

		sc, scErr := repo.NewStoreChange(d, presence, bou)
		//logger.L.Infof("in application.Handle worker %d for dataPiece with header %v ==> sc.A %d\n", i, d.GetHeader(), sc.A)
		a.toChanLog(fmt.Sprintf("in postparser.application.Handle for dataPiece with header %v, body %q ==> made sc %v, scRrr: %v", d.GetHeader(), d.GetBody(0), sc, scErr))

		if (bErr == nil && scErr != nil) ||
			(bErr != nil && scErr != nil && scErr.Error() != bErr.Error()) {
			prepErrs = append(prepErrs, scErr)
		}

		adus, _ := a.doHandle(d, sc, adub, header, bou, prepErrs)
		//logger.L.Errorf("in application.Handle case 1 adus: %v\n", adus)
		for _, v := range adus {
			//logger.L.Errorf("in application.Handle case 1 adu header: %v\n", v.H)
			a.toChanOut(v)
		}
		a.A.appRWLock.Unlock()
		w.Done()
		//for _, v := range errs {
		//logger.L.Errorf("in application.Handle err: %v\n", v)
		//}
		return

	}
	sc, scErr := repo.NewStoreChange(d, presence, bou)
	//logger.L.Infof("in application.Handle worker %d for dataPiece with header %v ==> sc: %v\n", i, d.GetHeader(), sc)

	//logger.L.Infof("in application.Handle sc %v, err: %v\n", sc, scErr)
	a.toChanLog(fmt.Sprintf("in postparser.application.Handle worker %d for dataPiece with header %v, body %q ==> made sc %v, scRrr: %v", i, d.GetHeader(), d.GetBody(0), sc, scErr))

	if (bErr == nil && scErr != nil) ||
		(bErr != nil && scErr != nil && scErr.Error() != bErr.Error()) {
		//logger.L.Errorf("in application.Handle NewStoreChange returned err: %v\n", bErr)
		prepErrs = append(prepErrs, scErr)
	}

	a.A.appRWLock.Lock()

	adus, _ := a.doHandle(d, sc, adub, header, bou, prepErrs)
	//logger.L.Errorf("in application.Handle case 2 %d adus:\n", len(adus))
	for _, v := range adus {
		//	logger.L.Errorf("in application.Handle case 2 adu header: %v, body %q\n", v.H, v.B.B)
		a.toChanOut(v)
	}
	a.A.appRWLock.Unlock()
	//logger.L.Infoln("in application.Handle Done")
	w.Done()
	/*
		for _, v := range errs {
			logger.L.Errorf("in application.Handle err: %v\n", v)
		}
	*/
}

// returns adus and errors based on dataPiece and store states
func (a *App) doHandle(d repo.DataPiece, sc repo.StoreChange, adub repo.AppDistributorBody, header []byte, bou repo.Boundary, prepErrs []error) ([]repo.AppDistributorUnit, []error) {
	var decrementErr error
	adus, errs, o := make([]repo.AppDistributorUnit, 0), prepErrs, repo.Unordered
	//p := d.Part()

	//logger.L.Infof("in application.doHandle handling p = %d\n", p)

	if sc.A != repo.Buffer {

		o, decrementErr = a.S.Dec(d)

		if decrementErr != nil {
			sc.A = repo.Buffer
		}

		a.toChanLog(fmt.Sprintf("in postparser.application.doHandle for apu with header %v got o = %d, decrementErr: %v", d.GetHeader(), o, decrementErr))

		//logger.L.Infof("in application.doHandle for apu with header %v got order = %d\n", d.GetHeader(), o)
	}

	a.S.Act(d, sc)

	if len(sc.From[repo.NewAppStoreKeyDetailed(d)]) == 2 && len(header) == 0 { // dataPiece after forked askd
		adub.B = append(sc.From[repo.NewAppStoreKeyDetailed(d)][true].D.H, adub.B...)
	}

	//logger.L.Infof("in application.doHandle len(adub.B) = %d\n", len(adub.B))

	if !d.IsSub() &&
		(sc.A != repo.Buffer && len(adub.B) != 0) { // dataPieces with matched parts

		aduh := CalcHeader(d, sc, o)
		//logger.L.Infof("in application.doHandle for apu with header %v got aduh = %v\n", d.GetHeader(), aduh)
		adu := repo.NewAppDistributorUnit(aduh, adub)

		a.toChanLog(fmt.Sprintf("in postparser.application.doHandle for apu with header %v got adu header: %v, body: %q", d.GetHeader(), adu.H, adu.B.B))
		adus = append(adus, adu)
	}

	//logger.L.Infof("in application.doHandle sc = %v, IsPartChanged(sc) %t\n", sc, repo.IsPartChanged(sc))
	if repo.IsPartChanged(sc) {
		//logger.L.Infof("in application.doHandle  apu with header %v checking buffer\n", d.GetHeader())
		adusFromBuffer, errsFromBuffer := a.HandleBuffer(repo.NewAppStoreKeyGeneralFromDataPiece(d), bou)
		/*
			for _, v := range errsFromBuffer {
				logger.L.Errorf("in application.doHandle apu with header %v extracted from buffer error: %v\n", d.GetHeader(), v)
			}
		*/
		for _, v := range adusFromBuffer {
			a.toChanLog(fmt.Sprintf("in postparser.application.doHandle apu with header %v extracted from buffer adu with header: %v", d.GetHeader(), v.H))
		}

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
	//logger.L.Infof("in application.CalcHeader header %q, err %v\n", header, err)
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
	aduh, askd, fo, fi, f, b, pre, post := repo.AppDistributorHeader{}, repo.NewAppStoreKeyDetailed(d), "", "", repo.FiFo{}, repo.BeginningData{}, repo.None, repo.None
	if d.IsSub() {
		return aduh
	}
	//logger.L.Infof("in application.CalcHeader from dataPiece header %v, body %q, sc %v, o: %v\n", d.GetHeader(), d.GetBody(0), sc, o)
	//logger.L.Infof("in application.CalcHeader d header %v, sc.To: %v, askd: %v\n", d.GetHeader(), sc.To, askd)
	askd = askd.IncPart()
	//fo, fi := sc.To[askd][false].D.FormName, sc.To[askd][false].D.FileName
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
	//b:=repo.NewBeginningData(sc.To)
	//logger.L.Infof("in application.CalcHeader m: %v\n", m)
	aduh = repo.NewAppDistributorHeaderStream(d, f, m, b)

	return aduh
}

func (a *App) HandleBuffer(askg repo.AppStoreKeyGeneral, bou repo.Boundary) ([]repo.AppDistributorUnit, []error) {
	return a.S.RegisterBuffer(askg, bou)
}
func (a *App) SetStopping() {

	a.A.stopping = true
	//logger.L.Errorf("application.SetStopping stopping = %t\n", a.A.stopping)
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

	//logger.L.Infoln("application.Stop wgWork is done")
	close(a.A.C.ChanOut)

	//logger.L.Infoln("application.Stop chanOut is closed")

	a.A.W.Sender.Wait()

	//logger.L.Infoln("application.Stop wgSend is done")

	close(a.A.C.Done)
}
