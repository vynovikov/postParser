package repo

import (
	"bytes"
	"errors"
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
	N    bool // is created on current part
	//Remove bool
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
	SK StreamKey // Stream identifier
	F  FiFo      // Form and File name
	M  Message   // Control message
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
	D []StreamKey // indexes to delete
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

func NewStreamData(sk StreamKey, f FiFo, sm Message) StreamData {
	return StreamData{
		SK: sk,
		F:  f,
		M:  sm,
	}
}

func NewAppDistributorHeader(t comm, s StreamData, u UnaryData) AppDistributorHeader {
	return AppDistributorHeader{
		T: t,
		S: s,
		U: u,
	}
}

func NewAppDistributorHeaderStream(d DataPiece, f FiFo, m Message) AppDistributorHeader {
	return AppDistributorHeader{
		T: ClientStream,
		S: StreamData{
			SK: StreamKey{
				TS:   d.TS(),
				Part: d.Part(),
			},
			F: f,
			M: m,
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

func NewDistributorUnitStream(ask AppStoreValue, d DataPiece, m Message) AppDistributorUnit {
	//logger.L.Infof("repo.NewDistributorUnitStream invoked by d.GetBody()  = %q\n", d.GetBody(0))
	sk := NewStreamKey(d.TS(), d.Part(), false)
	fifo := NewFiFo(ask.D.FormName, ask.D.FileName)
	sm := m

	sd := NewStreamData(sk, fifo, sm)
	//logger.L.Infof("in repo.NewDistributorUnitFromStore sk = %v\n", sk)
	adu := AppDistributorUnit{
		H: NewAppDistributorHeader(ClientStream, sd, UnaryData{}),
		B: NewDistributorBody(d.GetBody(0)),
	}
	//logger.L.Infof("in repo.NewDistributorUnitFromStore adu: %q\n", adu.B.B)

	return adu

}

func NewDistributorUnitStreamEmpty(ask AppStoreValue, d DataPiece, m Message) AppDistributorUnit {
	sk := NewStreamKey(d.TS(), d.Part(), false)
	fifo := NewFiFo("", "")
	sm := NewStreamMessage(m)

	sd := NewStreamData(sk, fifo, sm)
	//logger.L.Infof("in repo.NewDistributorUnitFromStore sk = %v\n", sk)

	return AppDistributorUnit{
		H: NewAppDistributorHeader(ClientStream, sd, UnaryData{}),
	}
}

func NewAppDistributorUnitUnary(d DataPiece, bou Boundary, m Message) AppDistributorUnit {
	//logger.L.Infof("repo.NewAppDistributorUnitUnary invoked with dataPiece header: %v, body: %q, bou: %q, message: %v\n", d.GetHeader(), d.GetBody(0), bou, m)
	//b:=make([]byte,0)
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

	//d.BodyCut(len(h))

	//logger.L.Infof("in repo.NewAppDistributorUnitUnary fifo: %v, m: %v, body: %q\n", fifo, m, d.GetBody(0))

	return AppDistributorUnit{
		H: NewAppDistributorHeader(Unary, StreamData{}, NewUnary(uk, fifo, m)),
		B: NewDistributorBody(d.GetBody(0)[len(h):]),
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
	return AppPieceHeader{}
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
				//N:    true,
			},
			S: true,
		}
	}
	return AppStoreKeyDetailed{SK: StreamKey{
		TS:   d.TS(),
		Part: d.Part(),
		//N:    true,
	},
		S: false,
	}

}

/*
	func (ask *AppStoreKeyDetailed) SetDisposition(d Disposition) {
		ask.D = d
	}
*/
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

type AppStoreValue struct {
	D Disposition
	B BeginningData
	E disposition
}

func NewAppStoreValue(d DataPiece, bou Boundary) (AppStoreValue, error) {
	//logger.L.Infof("repo.NewAppStoreValue invoked with datapiece header: %v, body %q, bou %q\n", d.GetHeader(), d.GetBody(0), bou)
	asv := AppStoreValue{}

	header, err := d.H(bou)
	//logger.L.Infof("in repo.NewAppStoreValue header: %q\n", header)
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

func CompleteAppStoreValue(asv AppStoreValue, d DataPiece, bou Boundary) (AppStoreValue, error) {
	ci := 0
	header, err := d.H(bou)
	//logger.L.Infof("in repo.CompleteAppStoreValue initial asv: %q, header: %q, err: %v\n", asv, header, err)
	if err != nil {
		/*
			logger.L.Errorf("in repo.CompleteAppStoreValue err: %v, contains? %t\n", err, strings.Contains(err.Error(), "no header found"))

			logger.L.Errorf("in repo.CompleteAppStoreValue first %t second %t third %t\n",
				!strings.Contains(err.Error(), "is not full"),
				!strings.Contains(err.Error(), "is ending part"),
				!strings.Contains(err.Error(), "no header found"))
		*/
		if !strings.Contains(err.Error(), "is not full") &&
			!strings.Contains(err.Error(), "is ending part") &&
			!strings.Contains(err.Error(), "no header found") {
			return asv, err
		}
		if strings.Contains(err.Error(), "no header found") {
			//logger.L.Infoln("in repo.CompleteAppStoreValue no header found")
			return AppStoreValue{}, err
		}
	}
	ci = bytes.Index(header, []byte("Content-Disposition"))
	//logger.L.Infof("in repo.CompleteAppStoreValue ci = %d\n", ci)

	if ci > 0 {

		if IsBoundary(asv.D.H, header, bou) {

			//logger.L.Infof("in repo.CompleteAppStoreValue header[ci:]: %q\n", header[ci:])

			raw := append(asv.D.H, header...)

			//			logger.L.Infof("in repo.CompleteAppStoreValue asv.D.H: %q\n",raw)
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

func (a *AppPieceUnit) L(bou Boundary) ([][]byte, int, error) {

	//lines, cut, err := make([][]byte, 0), 0, errors.New("")
	limit := Min(len(a.APB.B), MaxHeaderLimit)

	if a.APH.B == True {

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

// Cuts fisrt c bytes from dataPirce's body
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

	b := a.GetBody(Min(a.LL(), MaxHeaderLimit))
	//logger.L.Infof("in repo.H b: %q", b)

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
				//logger.L.Infof("in repo.NewStoreChange asv: %q, error: %v", asv, err)
				if err != nil {
					if !strings.Contains(err.Error(), "is not full") &&
						!strings.Contains(err.Error(), "is ending part") {
						return sc, err
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
					//logger.L.Infof("in repo.NewStoreChange asv: %q, error: %v", asv, err)
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
				//logger.L.Infoln("strait case")

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

					//logger.L.Infof("in repo.NewStoreChange true case sc: %v\n", sc)

					return sc, nil

				case Probably:
					//logger.L.Warnln("in repo.NewStoreChange Probably case")

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
			//logger.L.Infof("in repo.NewStoreChange dr: %v\n", dr)
			gr[askd.F().IncPart()] = dr
			sc.To = gr
			return sc, nil
		}

		gr[askd.T()] = dr
		sc.To = gr
		//logger.L.Infof("in repo.NewStoreChange sc: %v\n", sc)
		return sc, nil

	}
	// dataPiece is AppPieceUnit

	asv, err = NewAppStoreValue(d, bou)
	//logger.L.Infof("in repo.NewStoreChange asv: %v, err: %v\n", asv, err)
	if err != nil {
		//logger.L.Errorf("in repo.NewStoreChange err: %v\n", err)
		if !strings.Contains(err.Error(), "is not full") {
			return sc, err
		}
	}
	//logger.L.Infof("in repo.NewStoreChange asv = %v\n", asv)
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

		//logger.L.Infof("in repo.NewStoreChange d.B == False, True case A = %d\n", sc.A)

		//logger.L.Infof("in repo.NewStoreChange d.B == True, True case dr = %v\n", dr)
		gr[askd.F().IncPart()] = dr
		sc.To = gr
		if err != nil {
			if !strings.Contains(err.Error(), "is not full") {
				return sc, err
			}
		}
		//logger.L.Infof("in repo.NewStoreChange d.B == False, True case sc = %v\n", sc)
		return sc, err

		//sc.ASKD = NewAppStoreKeyDetailed(d)

	case False:
	}
	//logger.L.Infof("in repo.NewStoreChange d.B == False, sc = %v\n", sc)
	return sc, nil
}

/*
// get slice of AppStoreKeyDetailed from gr, putting S == false at first
func getFrom(gr map[AppStoreKeyDetailed]map[bool]AppStoreValue) map[AppStoreKeyDetailed]map[bool]AppStoreValue {

		n, askd,from := 0, AppStoreKeyDetailed{}

		if len(gr) == 0 {
			return askds
		}
		if len(gr) == 1 {
			for i := range gr {
				askds = append(askds, i)
			}
			return askds
		}
		for i := range gr {

			if n == 0 && i.S {
				askd = i
				n++
				continue
			}
			if n == 1 && !i.S {
				askds = append(askds, i)
				askds = append(askds, askd)
				return askds
			}

			askds = append(askds, i)
			n++
		}

		//logger.L.Infof("in repo.getFrom askds: %v\n", askds)

		return askds
	}
*/
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
	Max int
	Cur int
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
