package repo

import (
	"postParser/internal/adapters/driven/grpc/pb"
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

type BlockInfo struct {
	TS       time.Time
	Boundary string
}
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

type AppDistributorHeader struct {
	M        MultipartHeader
	FormName string
	FileName string
	TS       string
}

func NewAppDistributorHeader(m MultipartHeader, ts, fo, fi string) AppDistributorHeader {
	return AppDistributorHeader{
		M:        m,
		FormName: fo,
		FileName: fi,
		TS:       ts,
	}
}

func (h *AppDistributorHeader) SetFormMame(fo string) {
	h.FormName = fo
}

func (h *AppDistributorHeader) SetFileName(fi string) {
	h.FileName = fi
}

/*
func NewDistributorHeader(afu AppFeederUnit, header string) DistributorHeader {
	h := make([]string, 0)

	for _, v := range afu.H.SepHeader.PrevBody {
		h = append(h, v)
	}
	h = append(h, header)

	formName := ""
	//h[strings.Index(h, afu.R.H.Voc.FormName)+len(afu.R.H.Voc.FormName)+1 : FindNext(h, "\"", strings.Index(h, afu.R.H.Voc.FormName)+len(afu.R.H.Voc.FormName))]
	fileName := ""
	for _, v := range h {
		if strings.Contains(v, afu.R.H.Voc.FormName) {
			formName = v[strings.Index(v, afu.R.H.Voc.FormName)+len(afu.R.H.Voc.FormName)]
		}

		if strings.Contains(v, afu.R.H.Voc.FileName) {
			fileName = v[strings.Index(h, afu.R.H.Voc.FileName)+len(afu.R.H.Voc.FileName)+1 : FindNext(h, "\"", strings.Index(h, afu.R.H.Voc.FileName)+len(afu.R.H.Voc.FileName))]
		}
	}

	return DistributorHeader{
		FormName: formName,
		FileName: fileName,
		TS:       afu.R.H.TS,
	}
}
*/

type DistributorBody struct {
	B []byte
}

func NewDistributorBody(b []byte) DistributorBody {
	return DistributorBody{
		B: b,
	}
}
func (b *DistributorBody) SetBody(body []byte) {
	b.B = body
}

type AppDistributorUnit struct {
	H AppDistributorHeader
	B DistributorBody
}

func NewAppDistributorUnit(h AppDistributorHeader, b DistributorBody) AppDistributorUnit {
	return AppDistributorUnit{
		H: h,
		B: b,
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

type AppPieceHeader struct {
	Part int
	TS   string
	B    bool //is begin needed?
	E    bool //is end needed?
	CB   *CallBoard
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
func (p *AppPieceHeader) SetE(e bool) {
	p.E = e
}

type AppPieceBody struct {
	B []byte
}

func NewAppPieceBody(b []byte) AppPieceBody {
	return AppPieceBody{
		B: b,
	}
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

func (apu *AppPieceUnit) SetAPH(aph AppPieceHeader) {
	apu.APH = aph
}
func (apu *AppPieceUnit) SetAPB(apb AppPieceBody) {
	apu.APB = apb
}
