package repo

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Vocabulaty struct {
	Boundary     Boundary
	CDisposition string
	CType        string
	FormName     string
	FileName     string
}

func NewVocabulary(b Boundary) Vocabulaty {
	return Vocabulaty{
		Boundary:     b,
		CDisposition: "Content-Disposition",
		CType:        "Content-Type",
		FormName:     "name=",
		FileName:     "filename=",
	}
}

type ReceiverHeader struct {
	Part    int
	TS      string
	Bou     Boundary
	Unblock bool
}

func NewReceiverHeader(ts string, p int, bou Boundary) ReceiverHeader {

	return ReceiverHeader{
		Part: p,
		TS:   ts,
		Bou:  bou,
	}
}

type ReceiverBody struct {
	B []byte
}

func NewReceiverBody(n int) ReceiverBody {
	return ReceiverBody{
		B: make([]byte, n),
	}
}

type ReceiverSignal struct {
	Signal string
}

func (rh *ReceiverHeader) SetPart(p int) {
	rh.Part = p
}

type ReceiverUnit struct {
	H ReceiverHeader
	B ReceiverBody
}

func (r ReceiverUnit) GetHeader() string {
	return fmt.Sprint(r.H)
}

func (r ReceiverUnit) GetBody() []byte {
	return r.B.B
}

func NewReceiverUnit(h ReceiverHeader, b ReceiverBody) ReceiverUnit {
	return ReceiverUnit{
		H: h,
		B: b,
	}
}

func (r *ReceiverBody) SetBytes(buf []byte) {
	r.B = buf
}

func IncPart(h *ReceiverHeader) {
	h.Part++
}

type Boundary struct {
	Prefix []byte
	Root   []byte
	Suffix []byte
}

func NewBoundary(r []byte) Boundary {
	return Boundary{
		Root: r,
	}
}

type AppUnit interface {
	GetHeader() string
	GetBody() []byte
}

type AppFeederUnit struct {
	R ReceiverUnit
}

func NewAppFeederUnit(r ReceiverUnit) AppFeederUnit {
	return AppFeederUnit{
		R: r,
	}
}

func (afu *AppFeederUnit) SetReceiverUnit(r ReceiverUnit) {
	afu.R = r
}
func (a AppFeederUnit) GetHeader() string {
	return fmt.Sprint(a.R.H)
}
func (a AppFeederUnit) GetBody() []byte {
	return a.R.B.B
}

func (b *Boundary) SetBoundaryPrefix(bs []byte) {
	b.Prefix = bs
}
func (b *Boundary) SetBoundaryRoot(bs []byte) {
	b.Root = bs
}
func (afu *AppFeederUnit) SetBody(b []byte) {
	afu.R.B.B = b
}

type MultipartHeader struct {
	SeqNum int
}

func NewMultipartHeader(n int) MultipartHeader {
	return MultipartHeader{
		SeqNum: n,
	}
}

type StreamKey struct {
	TS   string
	Part int
	N    bool // is created on current part?
}

func NewStreamKey(ts string, part int, n bool) StreamKey {
	return StreamKey{
		TS:   ts,
		Part: part,
		N:    n,
	}
}
func (sk StreamKey) ToTrue() StreamKey {
	sk.N = true
	return sk
}
func (sk StreamKey) ToFalse() StreamKey {
	sk.N = false
	return sk
}
func (sk StreamKey) Reverse() StreamKey {
	sk.N = !sk.N
	return sk
}
func (sk StreamKey) IncPart() StreamKey {
	sk.Part++
	return sk
}
func (sk StreamKey) DecPart() StreamKey {

	if sk.Part > 0 {
		sk.Part--
		return sk
	}
	return sk // sk.Part == 0 remains unchanged
}

type FiFo struct {
	FormName string
	FileName string
}

func NewFiFo(fo, fi string) FiFo {
	return FiFo{
		FormName: fo,
		FileName: fi,
	}
}

type Action int

const (
	None     Action = iota // initisl value
	Start                  // start new stream and mark stream header as first
	Open                   // start new stream, don't mart
	Continue               // put data chunk into existing stream increment stream part
	StopLast               // stop last stream, create new one
	Close                  // stop curreant stream after sending this fata
	Finish                 // after sending data, mark data chunk as last, finish strean group

)

type Message struct {
	S          string
	PreAction  Action
	PostAction Action
}

func NewStreamMessage(m Message) Message {
	return m
}
func NewMessage(pre, post Action) Message {
	return Message{
		PreAction:  pre,
		PostAction: post,
	}
}

type StreamData struct {
	SK StreamKey     // Stream identifier
	F  FiFo          // Form and File name
	M  Message       // Control message
	B  BeginningData // Where stream begins
}
type CurrentPieceHeader struct {
	TS   string
	Part int
}
type UnaryKey struct {
	TS   string
	Part int
}

func NewUnaryKey(ts string, p int) UnaryKey {
	return UnaryKey{
		TS:   ts,
		Part: p,
	}
}

type UnaryData struct {
	UK UnaryKey
	F  FiFo    // Form and File name
	M  Message // Control message
}
type CloseData struct {
	TS string // indexes to delete
}

func NewUnary(uk UnaryKey, f FiFo, m Message) UnaryData {
	return UnaryData{
		UK: uk,
		F:  f,
		M:  m,
	}
}

type AppDistributorHeader struct {
	// Transfer type
	T comm

	// Unary info
	U UnaryData

	// Stream info
	S StreamData

	// Close info
	C CloseData
}

func NewCurrentPieceHeader(ts string, p int) CurrentPieceHeader {
	return CurrentPieceHeader{
		TS:   ts,
		Part: p,
	}
}

func NewStreamData(sk StreamKey, f FiFo, sm Message, b BeginningData) StreamData {
	return StreamData{
		SK: sk,
		F:  f,
		M:  sm,
		B:  b,
	}
}

func NewUnaryData(uk UnaryKey, f FiFo, m Message) UnaryData {
	return UnaryData{
		UK: uk,
		F:  f,
		M:  m,
	}
}

func NewAppDistributorHeader(t comm, s StreamData, u UnaryData) AppDistributorHeader {
	return AppDistributorHeader{
		T: t,
		S: s,
		U: u,
	}
}

func NewAppDistributorHeaderStream(d DataPiece, f FiFo, m Message, b BeginningData) AppDistributorHeader {
	return AppDistributorHeader{
		T: ClientStream,
		S: StreamData{
			SK: StreamKey{
				TS:   d.TS(),
				Part: d.Part(),
			},
			F: f,
			M: m,
			B: b,
		},
	}
}

type AppDistributorBody struct {
	B []byte
}

func NewDistributorBody(b []byte) AppDistributorBody {
	return AppDistributorBody{
		B: b,
	}
}
func (b *AppDistributorBody) SetBody(body []byte) {
	b.B = body
}

type AppDistributorUnit struct {
	H AppDistributorHeader
	B AppDistributorBody
}

func NewAppDistributorUnit(h AppDistributorHeader, b AppDistributorBody) AppDistributorUnit {
	return AppDistributorUnit{
		H: h,
		B: b,
	}
}
func (adu AppDistributorUnit) GetHeader() string {
	return fmt.Sprint(adu.H)
}
func (adu AppDistributorUnit) GetBody() []byte {
	return adu.B.B
}

func NewDistributorUnitStream(ask AppStoreValue, d DataPiece, m Message) AppDistributorUnit {
	sk := NewStreamKey(d.TS(), d.Part(), false)
	fifo := NewFiFo(ask.D.FormName, ask.D.FileName)
	b := ask.B
	sm := m

	sd := NewStreamData(sk, fifo, sm, b)
	adu := AppDistributorUnit{
		H: NewAppDistributorHeader(ClientStream, sd, UnaryData{}),
		B: NewDistributorBody(d.GetBody(0)),
	}

	return adu

}

func NewDistributorUnitStreamEmpty(ask AppStoreValue, d DataPiece, m Message) AppDistributorUnit {
	sk := NewStreamKey(d.TS(), d.Part(), false)
	fifo := NewFiFo("", "")
	sm := NewStreamMessage(m)

	sd := NewStreamData(sk, fifo, sm, BeginningData{})

	return AppDistributorUnit{
		H: NewAppDistributorHeader(ClientStream, sd, UnaryData{}),
	}
}

func NewAppDistributorUnitUnary(d DataPiece, bou Boundary, m Message) AppDistributorUnit {
	h, err := d.H(bou)
	if err != nil &&
		!strings.Contains(err.Error(), "is ending part") {
		return AppDistributorUnit{}
	}
	if len(h) >= len(d.GetBody(0)) {
		return AppDistributorUnit{}
	}

	fo, fi := GetFoFi(h)
	fifo := NewFiFo(fo, fi)

	uk := NewUnaryKey(d.TS(), d.Part())

	return AppDistributorUnit{
		H: NewAppDistributorHeader(Unary, StreamData{}, NewUnary(uk, fifo, m)),
		B: NewDistributorBody(d.GetBody(0)[len(h):]),
	}
}

func NewAppDistributorUnitUnaryComposed(ask AppStoreValue, d DataPiece, m Message) AppDistributorUnit {
	uk := NewUnaryKey(d.TS(), d.Part())
	fifo := NewFiFo(ask.D.FormName, ask.D.FileName)
	sm := m

	ud := NewUnaryData(uk, fifo, sm)
	adu := AppDistributorUnit{
		H: NewAppDistributorHeader(Unary, StreamData{}, ud),
		B: NewDistributorBody(d.GetBody(0)),
	}

	return adu
}

type disposition int

const (
	False disposition = iota
	True
	Probably
)

type comm int

const (
	Tech comm = iota
	Unary
	ClientStream
)

type sufficiency int

const (
	Incomplete sufficiency = iota
	Sufficient
	Insufficient
)

type AppPieceHeader struct {
	Part int
	TS   string
	B    disposition //is begin needed?
	E    disposition //is end needed?
}

func NewAppPieceHeader() AppPieceHeader {
	aph := AppPieceHeader{}
	return aph
}

func (p *AppPieceHeader) SetTS(ts string) {
	p.TS = ts
}

func (p *AppPieceHeader) SetPart(part int) {
	p.Part = part
}

func (p *AppPieceHeader) SetB(b disposition) {
	p.B = b
}
func (p *AppPieceHeader) SetE(e disposition) {
	p.E = e
}

type AppPieceBody struct {
	B []byte
}

func NewAppPieceBodyEmpty() AppPieceBody {
	return AppPieceBody{
		B: make([]byte, 0),
	}
}

func NewAppPieceBodyFilled(b []byte) AppPieceBody {
	return AppPieceBody{
		B: b,
	}
}
func (apb *AppPieceBody) SetBody(b []byte) {
	apb.B = b
}

type AppPieceUnit struct {
	APH AppPieceHeader
	APB AppPieceBody
}

func NewAppPieceUnit(aph AppPieceHeader, apb AppPieceBody) AppPieceUnit {
	return AppPieceUnit{
		APH: aph,
		APB: apb,
	}
}
func NewAppPieceUnitEmpty() AppPieceUnit {
	return AppPieceUnit{}
}
func NewAppPieceUnitCompose(p int, ts string) AppPieceUnit {
	return AppPieceUnit{
		APH: AppPieceHeader{
			Part: p,
			TS:   ts,
		},
		APB: AppPieceBody{},
	}
}
func (apu *AppPieceUnit) SetB(b []byte) {
	apu.APB.B = b
}

func (apu *AppPieceUnit) SetAPH(aph AppPieceHeader) {
	apu.APH = aph
}
func (apu *AppPieceUnit) SetAPB(apb AppPieceBody) {
	apu.APB = apb
}

type AppSubHeader struct {
	Part int
	TS   string
}

func NewASH(p int, ts string) AppSubHeader {
	return AppSubHeader{
		Part: p,
		TS:   ts,
	}
}

type AppSubBody struct {
	B []byte
}

type AppSub struct {
	ASH AppSubHeader
	ASB AppSubBody
}

func NewAppSub() AppSub {
	return AppSub{}
}

func (a *AppSub) SetPart(p int) {
	a.ASH.Part = p
}

func (a *AppSub) SetTS(ts string) {
	a.ASH.TS = ts
}

func (a *AppSub) SetB(b []byte) {
	a.ASB.B = b
}

type Disposition struct {
	FormName string
	FileName string
	H        []byte
}

func NewDisposition() Disposition {
	return Disposition{}
}

// NewDispositionFilled cuts header fron dataPiece, creates Disposition on its base
func NewDispositionFilled(d DataPiece, bou Boundary) (Disposition, error) {

	header, err := d.H(bou)
	d.BodyCut(len(header))
	dispo := NewDisposition()
	if err != nil {
		if strings.Contains(err.Error(), "is not full") {
			dispo.SetH(header)
			errCore := err.Error()[len("in repo.GetHeaderLines header \""):strings.Index(err.Error(), "\" is not full")]
			return dispo, fmt.Errorf("in repo.NewDispositionFilled header \"%s\" is not full", errCore)
		}
	}
	dispo.SetH(header)
	dispo.FormName, dispo.FileName = GetFoFi(header)

	return dispo, nil
}

func (d *Disposition) SetFormMame(fo string) {
	d.FormName = fo
}
func (d *Disposition) SetFileMame(fi string) {
	d.FileName = fi
}
func (d *Disposition) SetH(h []byte) {
	d.H = h
}

type AppStoreKeyDetailed struct {
	SK StreamKey
	S  bool //is created by appSub, mutating not allowed
}

func NewAppStoreKeyDetailed(d DataPiece) AppStoreKeyDetailed {
	if d.IsSub() {

		return AppStoreKeyDetailed{
			SK: StreamKey{
				TS:   d.TS(),
				Part: d.Part(),
				//N:    true,
			},
			S: true,
		}
	}
	return AppStoreKeyDetailed{
		SK: StreamKey{
			TS:   d.TS(),
			Part: d.Part(),
			//N:    true,
		},
		S: false,
	}

}
func (ask AppStoreKeyDetailed) DecPart() AppStoreKeyDetailed {
	return AppStoreKeyDetailed{
		SK: StreamKey{
			Part: ask.SK.Part - 1,
			TS:   ask.SK.TS,
			N:    ask.SK.N,
		},
		S: ask.S,
	}
}

func (ask AppStoreKeyDetailed) IncPart() AppStoreKeyDetailed {
	return AppStoreKeyDetailed{
		SK: StreamKey{
			Part: ask.SK.Part + 1,
			TS:   ask.SK.TS,
			N:    ask.SK.N,
		},
		S: ask.S,
	}
}

func (ask AppStoreKeyDetailed) T() AppStoreKeyDetailed {
	return AppStoreKeyDetailed{
		SK: StreamKey{
			Part: ask.SK.Part,
			TS:   ask.SK.TS,
			N:    ask.SK.N,
		},
		S: true,
	}
}

func (ask AppStoreKeyDetailed) F() AppStoreKeyDetailed {
	return AppStoreKeyDetailed{
		SK: StreamKey{
			Part: ask.SK.Part,
			TS:   ask.SK.TS,
			N:    ask.SK.N,
		},
		S: false,
	}
}

type AppStoreKeyGeneral struct {
	TS string
}
type AppStoreBufferID struct {
	ASKG AppStoreKeyGeneral
	I    int
}

func NewAppStoreBufferID(askg AppStoreKeyGeneral, i int) AppStoreBufferID {
	return AppStoreBufferID{
		ASKG: askg,
		I:    i,
	}
}

type AppStoreBufferIDs struct {
	ASKG AppStoreKeyGeneral
	I    []int
}

func NewAppStoreBufferIDs() *AppStoreBufferIDs {
	return &AppStoreBufferIDs{}
}

// Add appends new element to slice of ASBI.
// Tested in models_test.go
func (a *AppStoreBufferIDs) Add(asbi AppStoreBufferID) {
	if a.ASKG.TS == "" {
		a.ASKG = asbi.ASKG
		a.I = append([]int{}, asbi.I)
		return
	}
	if asbi.ASKG.TS == a.ASKG.TS {
		a.I = append(a.I, asbi.I)
		return
	}

}
func (a *AppStoreBufferIDs) GetIDs(askg AppStoreKeyGeneral) []int {
	if a.ASKG.TS == askg.TS {
		return a.I
	}
	return make([]int, 0)
}

func NewAppStoreKeyGeneralFromDataPiece(d DataPiece) AppStoreKeyGeneral {
	return AppStoreKeyGeneral{
		TS: d.TS(),
	}
}
func NewAppStoreKeyGeneralFromASKD(askd AppStoreKeyDetailed) AppStoreKeyGeneral {
	return AppStoreKeyGeneral{
		TS: askd.SK.TS,
	}
}

func NewAppStoreKeyGeneralFromFeeder(afu AppFeederUnit) AppStoreKeyGeneral {
	return AppStoreKeyGeneral{
		TS: afu.R.H.TS,
	}
}

type AppStoreValue struct {
	D Disposition
	B BeginningData
	E disposition
}

// NewAppStoreValue creates ASV based on dataPiece and boundary parameters.
// Tested in models_test.go
func NewAppStoreValue(d DataPiece, bou Boundary) (AppStoreValue, error) {
	asv := AppStoreValue{}

	header, err := d.H(bou)
	asv.B = NewBeginningData(d.Part())
	asv.E = d.E()
	if err != nil {

		if !strings.Contains(err.Error(), "is not full") &&
			!strings.Contains(err.Error(), "is ending part") {
			return asv, err
		}
		if strings.Contains(err.Error(), "is not full") {
			asv.D.H = header
			return asv, err
		}
	}
	asv.D.H = header
	asv.D.FormName, asv.D.FileName = GetFoFi(asv.D.H)
	return asv, nil
}

// CompleteAppStoreValue completes given ASB based on dataPiece and boundary parameters.
// Tested in models_test.go
func CompleteAppStoreValue(asv AppStoreValue, d DataPiece, bou Boundary) (AppStoreValue, error) {
	ci := 0
	header, err := d.H(bou)
	if err != nil {
		if !strings.Contains(err.Error(), "is not full") &&
			!strings.Contains(err.Error(), "is ending part") &&
			!strings.Contains(err.Error(), "no header found") {
			return asv, err
		}
		if strings.Contains(err.Error(), "no header found") {
			return AppStoreValue{}, err
		}
	}
	ci = bytes.Index(header, []byte("Content-Disposition"))

	if ci > 0 {

		if IsBoundary(asv.D.H, header, bou) {

			raw := append(asv.D.H, header...)

			asv.D.H = raw[bytes.Index(raw, []byte("Content-Disposition")):]
			asv.D.FormName, asv.D.FileName = GetFoFi(asv.D.H)
			asv.E = d.E()

			return asv, nil
		}

		return asv, err
	}
	asv.D.H = append(asv.D.H, header...)
	asv.D.FormName, asv.D.FileName = GetFoFi(asv.D.H)

	return asv, err
}

type BeginningData struct {
	Part int
}

func NewBeginningData(bp int) BeginningData {
	return BeginningData{
		Part: bp,
	}
}

type DataPiece interface {
	B() disposition
	E() disposition
	LL() int
	H(Boundary) ([]byte, error)
	Part() int
	TS() string
	IsSub() bool
	BodyCut(int)
	GetBody(int) []byte
	GetHeader() string
	Prepend([]byte)
}

func (a *AppPieceUnit) B() disposition {
	return a.APH.B
}

func (a *AppPieceUnit) E() disposition {
	return a.APH.E
}

func (a *AppPieceUnit) Body() []byte {
	return a.APB.B
}

func (a *AppPieceUnit) Part() int {
	return a.APH.Part
}

func (a *AppPieceUnit) TS() string {
	return a.APH.TS
}

func (a *AppPieceUnit) IsSub() bool {
	return false
}

// LL returns length of dataPiece's body
func (a *AppPieceUnit) LL() int {
	return len(a.APB.B)
}

// BodyCut cuts fisrt c bytes from dataPirce's body.
// Tested in models_test.go
func (a *AppPieceUnit) BodyCut(c int) {
	if c < len(a.APB.B)-1 {
		a.APB.B = a.APB.B[c:]
	} else {
		a.APB.B = []byte{}
	}
}

// GetBody returns first n bytes of dataPiece's body
func (a *AppPieceUnit) GetBody(n int) []byte {

	lenb := len(a.APB.B)
	if n > 0 && n < lenb {
		return a.APB.B[:n]
	}
	return a.APB.B
}

// Prepend adds b to body, placing addition in front of old content
func (a *AppPieceUnit) Prepend(b []byte) {
	old := a.APB.B

	a.APB.B = append([]byte{}, b...)
	a.APB.B = append(a.APB.B, old...)
}

// GetHeader returns header of dataPiece as string
func (a *AppPieceUnit) GetHeader() string {
	return fmt.Sprintf("%v", a.APH)
}

// H return header lines found in the beginning of the body and error
// Tested in models_test.go
func (a *AppPieceUnit) H(bou Boundary) ([]byte, error) {

	b := a.GetBody(Min(a.LL(), MaxHeaderLimit))

	return GetHeaderLines(b, bou)

}

func (a *AppSub) B() disposition {
	return False
}

func (a *AppSub) E() disposition {
	return Probably
}
func (a *AppSub) L(bou Boundary) ([][]byte, int, error) {
	lines := make([][]byte, 0)
	return lines, 0, nil
}
func (a *AppSub) LL() int {
	return len(a.ASB.B)
}
func (a *AppSub) Part() int {
	return a.ASH.Part
}
func (a *AppSub) TS() string {
	return a.ASH.TS
}
func (a *AppSub) IsSub() bool {
	return true
}
func (a *AppSub) BodyCut(int) {
}
func (a *AppSub) GetBody(int) []byte {
	return a.ASB.B
}
func (a *AppSub) Prepend([]byte) {
}

func (a *AppSub) SetBody(b []byte) {
	a.ASB.B = b
}
func (a *AppSub) GetHeader() string {
	return fmt.Sprintf("%v", a.ASH)
}
func (a *AppSub) SetHeader(ash AppSubHeader) {
	a.ASH = ash
}

func (a *AppSub) H(bou Boundary) ([]byte, error) {
	return a.ASB.B, nil
}

type Presense struct {
	ASKG bool                                           // is askg met
	ASKD bool                                           // is askd met
	OB   bool                                           // is true branch met
	GR   map[AppStoreKeyDetailed]map[bool]AppStoreValue // general key record
}

func NewPresense(askg, askd, tb bool, gr map[AppStoreKeyDetailed]map[bool]AppStoreValue) Presense {
	return Presense{
		ASKG: askg,
		ASKD: askd,
		OB:   tb,
		GR:   gr,
	}
}

type DetailedRecord struct {
	ASKD AppStoreKeyDetailed
	DR   map[bool]AppStoreValue
}

// NewStoreChange calculates how store.R should be changed due to dataPiece.
// Tested in models_test.go
func NewStoreChange(d DataPiece, p Presense, bou Boundary) (StoreChange, error) {
	sc, _, asv, dr, gr, askd, ok, err := StoreChange{}, make([]AppStoreKeyDetailed, 0), AppStoreValue{}, make(map[bool]AppStoreValue), make(map[AppStoreKeyDetailed]map[bool]AppStoreValue), NewAppStoreKeyDetailed(d), false, errors.New("")

	if d.B() == True {
		if !p.ASKG {
			sc.A = Buffer
			return sc, fmt.Errorf("in repo.NewStoreChange TS \"%s\" is unknown", d.TS())
		}
		if !p.ASKD {
			sc.A = Buffer
			return sc, fmt.Errorf("in repo.NewStoreChange for given TS \"%s\", Part \"%d\" is unexpected", d.TS(), d.Part())
		}
		sc.From = p.GR
		if asv, ok = p.GR[askd][false]; ok {

			if asv.D.FormName == "" {
				asv, err = CompleteAppStoreValue(asv, d, bou)
				if err != nil {
					if !strings.Contains(err.Error(), "is not full") &&
						!strings.Contains(err.Error(), "is ending part") {
						return sc, err
					}
					if strings.Contains(err.Error(), "is ending part") {
						asv.E = d.E()

						dr[false] = asv
						gr[askd.IncPart().F()] = dr

						sc.A = Change
						sc.To = gr

						return sc, nil
					}

				}
			}
			asv.E = d.E()
			dr[false] = asv

			switch len(p.GR) {
			case 1:
				if len(p.GR[askd]) == 2 {

					sc.A = Change

					m3t := p.GR[askd][true]
					asv, err = CompleteAppStoreValue(m3t, d, bou)

					if err != nil {
						if strings.Contains(err.Error(), "no header found") { // no new asv

							gr[askd.IncPart().F()] = dr
							sc.To = gr

							return sc, err
						}
					}
					// got new asv

					dr[false] = asv
					gr[askd.IncPart().F()] = dr
					sc.To = gr

					return sc, nil
				}

				switch d.E() {
				case False:
					sc.A = Change
					gr[askd.F()] = dr
					sc.To = gr

					return sc, nil

				case True:
					sc.A = Change
					gr[askd.IncPart().F()] = dr
					sc.To = gr

					return sc, nil

				case Probably:

					gr[askd.F()] = dr

					sc.To = gr
					sc.A = Change

					return sc, nil
				}

			case 2: // detailed map has 2 records
				sc.A = Change
				m3t := p.GR[askd.T()]
				dr[true] = m3t[true]

				if len(p.GR[askd]) == 2 {

					m3f := p.GR[askd.F()]

					asv, err = CompleteAppStoreValue(m3t[true], d, bou)

					if err != nil {

						if strings.Contains(err.Error(), "no header found") { // no new asv

							dr[false] = m3f[false]
							gr[askd.IncPart().F()] = dr
							sc.To = gr

							return sc, err
						}
					}
					dr[false] = asv
					gr[askd.IncPart().F()] = dr
					sc.To = gr

					return sc, nil

				}
				gr[askd.IncPart().F()] = dr
				sc.To = gr

				return sc, nil
			}
		}
	}

	// d.B == False

	asv.B.Part = d.Part()
	asv.E = d.E()
	sc.A = Change
	sc.From = p.GR

	if d.IsSub() { // dataPiece is AppSub

		asv.D.H = d.GetBody(0)
		dr[true] = asv

		if m3f, ok := p.GR[askd.F()]; ok {
			dr[false] = m3f[false]
			gr[askd.F().IncPart()] = dr
			sc.To = gr
			return sc, nil
		}

		gr[askd.T()] = dr
		sc.To = gr
		return sc, nil

	}
	// dataPiece is AppPieceUnit

	asv, err = NewAppStoreValue(d, bou)
	if err != nil {
		if !strings.Contains(err.Error(), "is not full") {
			return sc, err
		}
	}
	dr[false] = asv

	switch d.E() {
	case Probably:
		if p.OB {

			dr[true] = p.GR[askd.T()][true]
			gr[askd.IncPart()] = dr
			sc.To = gr
		} else {

			gr[askd] = dr
			sc.To = gr

		}
	case True:
		gr[askd.F().IncPart()] = dr
		sc.To = gr
		if err != nil {
			if !strings.Contains(err.Error(), "is not full") {
				return sc, err
			}
		}
		return sc, err

	case False:
	}
	return sc, nil
}

type StoreAction int

const (
	Buffer StoreAction = iota // store dataPiece in Buffer
	Change                    // mutate store record
	Delete                    // delete detailed record
	Remove                    // delete general record. Should be seg an application.Handle based on s.Counter == 0

)

type StoreChange struct {
	A    StoreAction
	From map[AppStoreKeyDetailed]map[bool]AppStoreValue // deteiled map record to be changed
	To   map[AppStoreKeyDetailed]map[bool]AppStoreValue // changed deteiled map record
}

type Counter struct {
	Max     int
	Cur     int
	Started bool
	Blocked bool
}

func NewCounter() Counter {
	return Counter{}
}

type Order int

const (
	Unordered    Order = iota // tech value if max was decrenebted
	First                     // first adu of the TS
	Last                      // last adu of the TS
	FirstAndLast              // first and last simultaneously
	Intermediate              // not first and not last
)

// IsPartChanged returns true of part of ASKD to be changed due to sc.
// Tested in models_test.go
func IsPartChanged(sc StoreChange) bool {
	pFrom, pTo := 0, 0
	if len(sc.From) == 0 && len(sc.To) == 1 {
		for i := range sc.To {
			if sc.To[i][i.S].E == Probably {
				return false
			}
		}
		return true
	}
	for i := range sc.From {
		pFrom = i.SK.Part
		break
	}
	for i := range sc.To {
		pTo = i.SK.Part
		break
	}
	return pTo == pFrom+1
}

type rType int

type GRequest struct {
	ts        string
	FieldName string
	FileInfo  bool
	FileData  bool
	FileName  string
	ByteChunk []byte
	IsFirst   bool
	IsLast    bool
	RType     rType
}

const (
	U rType = iota
	S
)

type GId struct {
	wantIDs []int
	gotIDs  []int
}

func NewGId() *GId {
	return &GId{}
}
func (g *GId) AddGIds(i, j int) {
	g.gotIDs = append(g.gotIDs, i)
	g.wantIDs = append(g.wantIDs, j)
}

// IsOk returns true if got data corresponds to want.
// Tested in models_test.go
func IsOk(name string, want, got []GRequest) bool {

	//logger.L.Infof("in repo IsOk len(got) = %d\n", len(got))

	found, ids, nlast := false, NewGId(), 0

	if len(want) == 0 || len(got) == 0 || len(want) != len(got) {
		return false
	}
	if len(want) == 1 && len(got) == 1 && got[0].IsFirst && got[0].IsLast {
		return true
	}

	// free to call x & y by indexes

	if !got[0].IsFirst {
		return false
	}
	if !got[len(got)-1].IsLast {
		return false
	}

	// comparing x and y

	for i, v := range got {

		if v.RType == U { // unary comparision

			for _, w := range want {
				if v.FieldName == w.FieldName &&
					v.FileName == w.FileName &&
					len(v.ByteChunk) == len(w.ByteChunk) &&
					bytes.Contains(v.ByteChunk, w.ByteChunk) {

					found = true

					if len(want) == 1 && len(got) == 1 {
						return true
					}
					break
				}
			}
			if !found {
				return false
			}
			found = false
		}
		if v.RType == S && v.FileInfo { // stream comparision
			found = true
			for j, w := range want {
				found = false

				if v.RType == w.RType && v.FieldName == w.FieldName && v.FileName == w.FileName {

					nlast = 0
					m := j + 1

					for m < len(want) && want[m].FileData { // looping through want

						found = false
						n := Max(i+1, nlast+1)

						for n < len(got) {

							if len(want[m].ByteChunk) == len(got[n].ByteChunk) && bytes.Contains(want[m].ByteChunk, got[n].ByteChunk) {

								nlast = n
								found = true
								ids.AddGIds(n, m)

								break
							}
							n++
						}
						if !found {
							return false
						}
						m++
					}
				}
			}
		}
	}

	return true
}

type WaitGroups struct {
	M       map[AppStoreKeyGeneral]*sync.WaitGroup
	Workers sync.WaitGroup
	Sender  sync.WaitGroup
}

type Channels struct {
	ChanIn  chan AppFeederUnit
	ChanOut chan AppDistributorUnit
	ChanLog chan string
	Done    chan struct{}
}
