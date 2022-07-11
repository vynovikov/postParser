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
	Prefix string
	Root   string
	Suffix string
}

func NewBoundary(p, r, s string) Boundary {
	return Boundary{
		Prefix: p,
		Root:   r,
		Suffix: s,
	}
}

func NewReceiverHeader(ts string, peaked []byte) ReceiverHeader {
	s := string(peaked)
	bPrefix, bRoot := FindBoundary(s)

	boundary := NewBoundary(bPrefix, bRoot, "")

	return ReceiverHeader{
		Part: 0,
		TS:   ts,
		Voc:  NewVocabulary(boundary),
	}
}
func NewReceiverBody() ReceverBody {
	return ReceverBody{
		B: make([]byte, 1024),
	}
}

type SepHeader struct {
	IsBoundary bool
	PrevBody   string
}

func NewSepHeader(isBoundary bool, prevBody string) *SepHeader {
	return &SepHeader{
		IsBoundary: isBoundary,
		PrevBody:   prevBody,
	}
}

type SepBody struct {
	Name   string
	SeqNum int
}

func NewSepBody(name string, seqNum int) *SepBody {
	return &SepBody{
		Name:   name,
		SeqNum: seqNum,
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

type AppFeederUnit struct {
	H *AppFeederHeader
	R ReceiverUnit
}

func NewAppFeaderUnit(h *AppFeederHeader, r ReceiverUnit) AppFeederUnit {
	return AppFeederUnit{
		H: h,
		R: r,
	}
}

func (af *AppFeederUnit) SetReceiverUnit(r ReceiverUnit) {
	af.R = r
}

//Todo test embedded structs with pointers

type MultipartHeader struct {
	SeqNum int
}

type DistributorHeader struct {
	M    MultipartHeader
	Name string
	TS   string
}
type DistributorBody struct {
	B []byte
}

type AppDistributorUnit struct {
	H DistributorHeader
	B DistributorBody
}
