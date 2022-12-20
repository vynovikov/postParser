package repo

import (
	"fmt"
	"postParser/internal/adapters/driven/grpc/pb"
	"strings"
	"time"
)

type Unit struct {
	Name  string
	Value []byte
}

type Meta struct {
	boundary     string
	bodyStartPos int
	streamPart   stream
	boundaryPart boundary
	headerPart   header
}

type stream struct {
	previous *pb.PostParser_MultiPartClient
}

type boundary struct {
	previous string
}
type header struct {
	previos string
}

/*
	type BlockInfo struct {
		TS       time.Time
		Boundary stringS
	}
*/
type FormInfo struct {
	TS       time.Time
	FormName string
}

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
	Part int
	TS   string
	Voc  Vocabulaty
}
type ReceverBody struct {
	B []byte
}
type ReceiverSignal struct {
	Signal string
}

func (rh *ReceiverHeader) SetPart(p int) {
	rh.Part = p
}

func NewReceiverSignal(s string) ReceiverSignal {
	return ReceiverSignal{
		Signal: s,
	}
}

/*
	type EnvelopeOut struct {
		FI        *FormInfo
		Part      int
		FormValue []byte
	}

	type EnvelopeIn struct {
		I    *BlockInfo
		Part int
		B    []byte
	}
*/
type ReceiverUnit struct {
	H ReceiverHeader
	B ReceverBody
	S ReceiverSignal
}

func NewReceiverUnit(h ReceiverHeader, b ReceverBody, s ReceiverSignal) ReceiverUnit {
	return ReceiverUnit{
		H: h,
		B: b,
		S: s,
	}
}

func (r *ReceverBody) SetBytes(buf []byte) {
	r.B = buf
}

func IncPart(h *ReceiverHeader) {
	h.Part++
}
func (r *ReceiverUnit) SetSignal(s ReceiverSignal) {
	r.S = s
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

func NewReceiverHeader(ts string, peaked []byte) ReceiverHeader {

	boundary := FindBoundary(peaked)

	return ReceiverHeader{
		Part: 0,
		TS:   ts,
		Voc:  NewVocabulary(boundary),
	}
}
func NewReceiverBody(n int) ReceverBody {
	return ReceverBody{
		B: make([]byte, n),
	}
}

type SepHeader struct {
	IsBoundary bool
	Lines      []string
}

func NewSepHeader(isBoundary bool, prevBody []string) *SepHeader {
	return &SepHeader{
		IsBoundary: isBoundary,
		Lines:      prevBody,
	}
}

func NewSepHeaderBP(isBoundary bool, prevBody []string) SepHeader {
	return SepHeader{
		IsBoundary: isBoundary,
		Lines:      prevBody,
	}
}

type SepBody struct {
	Line string
}

func NewSepBody(line string) *SepBody {
	return &SepBody{
		Line: line,
	}
}
func NewSepBodyBP(l string) SepBody {
	return SepBody{
		Line: l,
	}
}

type AppFeederHeader struct {
	SepHeader *SepHeader
	SepBody   *SepBody
	PrevPart  int
}

func NewAppFeederHeader(sepHeader *SepHeader, sepBody *SepBody, prevPart int) *AppFeederHeader {
	return &AppFeederHeader{
		SepHeader: sepHeader,
		SepBody:   sepBody,
		PrevPart:  prevPart,
	}
}

type AppFeederHeaderBP struct {
	SepHeader SepHeader
	SepBody   SepBody
	PrevPart  int
}

func NewAppFeederHeaderBP(sh SepHeader, sb SepBody, pp int) AppFeederHeaderBP {
	return AppFeederHeaderBP{
		SepHeader: sh,
		SepBody:   sb,
		PrevPart:  pp,
	}
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

func (b *Boundary) SetBoundaryPrefix(bs []byte) {
	b.Prefix = bs
}
func (b *Boundary) SetBoundaryRoot(bs []byte) {
	b.Root = bs
}
func (afu *AppFeederUnit) SetBody(b []byte) {
	afu.R.B.B = b
}

func (sh *SepHeader) Set(b bool, s []string) {
	sh.IsBoundary = false
	sh.Lines = s
}

//Todo test embedded structs with pointers

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
	//Remove bool
}

func NewStreamKey(ts string, part int) StreamKey {
	return StreamKey{
		TS:   ts,
		Part: part,
	}
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

type action int

const (
	None     action = iota // start new stream or appending to last stream
	StopLast               // stop last stream, create new one
	Stop                   // stop curreant stream after sending this fata
	Finish                 // finish strean section after sending this data

)

type Message struct {
	S          string
	PreAction  action
	PostAction action
}

func NewStreamMessage(m Message) Message {
	return m
}

type StreamData struct {
	SK StreamKey          // Stream identifier
	F  FiFo               // Form and File name
	M  Message            // Control message
	C  CurrentPieceHeader // Current part & ts
}
type CurrentPieceHeader struct {
	TS   string
	Part int
}
type UnaryData struct {
	TS string
	F  FiFo               // Form and File name
	M  Message            // Control message
	C  CurrentPieceHeader // Current part & ts
}
type CloseData struct {
	D []StreamKey // indexes to delete
}

func NewUnary(ts string, f FiFo, m Message) UnaryData {
	return UnaryData{
		TS: ts,
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

func NewStreamData(sk StreamKey, f FiFo, sm Message, cph CurrentPieceHeader) StreamData {
	return StreamData{
		SK: sk,
		F:  f,
		M:  sm,
		C:  cph,
	}
}

func NewAppDistributorHeader(t comm, s StreamData, u UnaryData) AppDistributorHeader {
	return AppDistributorHeader{
		T: t,
		S: s,
		U: u,
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

func NewDistributorUnitStream(ask AppStoreValue, d DataPiece, m Message) AppDistributorUnit {
	//logger.L.Infof("repo.NewDistributorUnitStream invoked by d.GetBody()  = %q\n", d.GetBody(0))
	sk := NewStreamKey(d.TS(), d.Part())
	fifo := NewFiFo(ask.D.FormName, ask.D.FileName)
	sm := m
	cph := CurrentPieceHeader{}

	sd := NewStreamData(sk, fifo, sm, cph)
	//logger.L.Infof("in repo.NewDistributorUnitFromStore sk = %v\n", sk)
	adu := AppDistributorUnit{
		H: NewAppDistributorHeader(ClientStream, sd, UnaryData{}),
		B: NewDistributorBody(d.GetBody(0)),
	}
	//logger.L.Infof("in repo.NewDistributorUnitFromStore adu: %q\n", adu.B.B)

	return adu

}

func NewDistributorUnitStreamEmpty(ask AppStoreValue, d DataPiece, m Message) AppDistributorUnit {
	sk := NewStreamKey(d.TS(), d.Part())
	fifo := NewFiFo("", "")
	sm := NewStreamMessage(m)
	cph := CurrentPieceHeader{}

	sd := NewStreamData(sk, fifo, sm, cph)
	//logger.L.Infof("in repo.NewDistributorUnitFromStore sk = %v\n", sk)

	return AppDistributorUnit{
		H: NewAppDistributorHeader(ClientStream, sd, UnaryData{}),
	}
}

func NewAppDistributorUnitUnary(d DataPiece, bou Boundary, m Message) AppDistributorUnit {
	h, err := d.H(bou)
	if err != nil && !strings.Contains(err.Error(), "is ending part") {
		return AppDistributorUnit{}
	}

	fo, fi := GetFoFi(h)
	fifo := NewFiFo(fo, fi)

	d.BodyCut(len(h))

	return AppDistributorUnit{
		H: NewAppDistributorHeader(Unary, StreamData{}, NewUnary(d.TS(), fifo, m)),
		B: NewDistributorBody(d.GetBody(0)),
	}
}

type CallBoard struct {
	FormName     string
	FileName     string
	InitPart     int
	InitFragment []byte
}

func NewCB(p int, f []byte) *CallBoard {
	return &CallBoard{
		InitPart:     p,
		InitFragment: f,
	}
}

func (c *CallBoard) SetFormMame(f string) {
	c.FormName = f
}

func (c *CallBoard) SetFileMame(f string) {
	c.FileName = f
}

type probability int

const (
	False probability = iota
	True
	Probably
	Last
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
	B    bool        //is begin needed?
	E    probability //is end needed?
}

func NewAppPieceHeader() AppPieceHeader {
	return AppPieceHeader{}
}

func (p *AppPieceHeader) SetTS(ts string) {
	p.TS = ts
}

func (p *AppPieceHeader) SetPart(part int) {
	p.Part = part
}

func (p *AppPieceHeader) SetB(b bool) {
	p.B = b
}
func (p *AppPieceHeader) SetE(e probability) {
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

// may be deleted
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

/*
	func NewAppPieceUnitFromAS(as AppSub) AppPieceUnit {
		return AppPieceUnit{
			APH: AppPieceHeader{
				Part: as.Part,
				TS:   as.TS,
				E:    true,
			},
			APB: AppPieceBody{},
		}
	}
*/
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

// Cuts header fron datapiece, creates Disposition on its base
func NewDispositionFilled(d DataPiece, bou Boundary) (Disposition, error) {

	header, err := d.H(bou)
	d.BodyCut(len(header))
	//logger.L.Infof("in repo.NewDispositionFilled header: %q, err: %v\n", header, err)
	dispo := NewDisposition()
	if err != nil {
		if strings.Contains(err.Error(), "is not full") {
			dispo.SetH(header)
			errCore := err.Error()[len("in repo.GetHeaderLines header \""):strings.Index(err.Error(), "\" is not full")]
			//logger.L.Infof("in store.calcHeader errCore: \"%s\"\n", errCore)
			return dispo, fmt.Errorf("in repo.NewDispositionFilled header \"%s\" is not full", errCore)
		}
	}
	dispo.SetH(header)
	dispo.FormName, dispo.FileName = GetFoFi(header)
	//logger.L.Infof("in repo.NewDispositionFilled dispo header: %q, formName: %q, fileName: %q\n", dispo.H, dispo.FormName, dispo.FileName)
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
			},
			S: true,
		}
	}
	return AppStoreKeyDetailed{SK: StreamKey{
		TS:   d.TS(),
		Part: d.Part(),
	},
		S: false,
	}

}

/*
	func (ask *AppStoreKeyDetailed) SetDisposition(d Disposition) {
		ask.D = d
	}
*/
func (ask AppStoreKeyDetailed) IncPart() AppStoreKeyDetailed {
	return AppStoreKeyDetailed{
		SK: StreamKey{
			Part: ask.SK.Part + 1,
			TS:   ask.SK.TS,
		},
		S: ask.S,
	}
}

func (ask AppStoreKeyDetailed) T() AppStoreKeyDetailed {
	return AppStoreKeyDetailed{
		SK: StreamKey{
			Part: ask.SK.Part,
			TS:   ask.SK.TS,
		},
		S: true,
	}
}

func (ask AppStoreKeyDetailed) F() AppStoreKeyDetailed {
	return AppStoreKeyDetailed{
		SK: StreamKey{
			Part: ask.SK.Part,
			TS:   ask.SK.TS,
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
		//logger.L.Infof("in repo.GetIDs returning %d\n", a.I)
		return a.I
	}
	return make([]int, 0)
}

func NewAppStoreKeyGeneral(d DataPiece) AppStoreKeyGeneral {
	return AppStoreKeyGeneral{
		TS: d.TS(),
	}
}

type AppStoreValue struct {
	D Disposition
	B BeginningData
	E probability
}

type BeginningData struct {
	TS        string
	BeginPart int
}

func NewBeginningData(ts string, bp int) BeginningData {
	return BeginningData{
		TS:        ts,
		BeginPart: bp,
	}
}

type DataPiece interface {
	B() bool
	E() probability
	L(Boundary) ([][]byte, int, error)
	LL() int
	H(Boundary) ([]byte, error)
	Part() int
	TS() string
	IsSub() bool
	//Prepare(string, int)
	BodyCut(int)
	GetBody(int) []byte
	GetHeader() string
	Prepend([]byte)
}

func (a *AppPieceUnit) B() bool {
	return a.APH.B
}

func (a *AppPieceUnit) E() probability {
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

func (a *AppPieceUnit) L(bou Boundary) ([][]byte, int, error) {

	//lines, cut, err := make([][]byte, 0), 0, errors.New("")
	limit := Min(len(a.APB.B), MaxHeaderLimit)

	if a.APH.B {

		//logger.L.Infof("in repo.L trying to get lines from : %q\n", a.APB.B)

		lines, cut, err := GetLinesRightBegin(a.APB.B[:limit], limit, bou)

		return lines, cut, err

	}
	lines, cut, err := GetLinesRightMiddle(a.APB.B[:limit], limit)

	return lines, cut, err
}
func (a *AppPieceUnit) LL() int {
	return len(a.APB.B)
}
func (a *AppPieceUnit) Prepare(ts string, p int) {
	aph := AppPieceHeader{}
	aph.SetTS(ts)
	aph.SetPart(p)

	a.APH = aph

}

func (a *AppPieceUnit) BodyCut(c int) {
	//logger.L.Infof("repo.BodyCut for dataPiece's body %q of length %d invoked with parameters c = %d\n", a.APB.B, len(a.APB.B), c)
	if c < len(a.APB.B)-1 {
		a.APB.B = a.APB.B[c:]
		//logger.L.Infof("in repo.BodyCut body became: %q\n", a.APB.B)

	} else {

		a.APB.B = []byte{}

	}
}

func (a *AppPieceUnit) GetBody(n int) []byte {

	lenb := len(a.APB.B)
	if n > 0 && n < lenb {
		return a.APB.B[:n]
	}
	return a.APB.B
}

func (a *AppPieceUnit) Prepend(b []byte) {
	old := a.APB.B

	a.APB.B = append([]byte{}, b...)
	//	a.APB.B = append(a.APB.B, bou.Prefix...)
	a.APB.B = append(a.APB.B, old...)
}
func (a *AppPieceUnit) GetHeader() string {
	return fmt.Sprintf("%v", a.APH)
}
func (a *AppPieceUnit) H(bou Boundary) ([]byte, error) {
	hbs := make([]byte, 0)

	b := a.GetBody(Min(a.LL(), MaxHeaderLimit))
	//logger.L.Infof("in repo.H b: %q", b)

	hbs = append(hbs)

	return GetHeaderLines(b, bou)

}

func (a *AppSub) B() bool {
	return false
}

func (a *AppSub) E() probability {
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
