package grpc

import (
	"context"
	"fmt"
	pb "postParser/internal/adapters/driven/grpc/pb"
	"postParser/internal/logger"
	"postParser/internal/repo"
	"sort"

	errs "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Stream pb.PostParser_MultiPartClient
	req    *pb.FileUploadReq
	err    error
)

type Transmitter interface {
	Transmit([]repo.AppDistributorUnit)
}

type TransmitAdapter struct {
	c      pb.PostParserClient
	M      map[repo.StreamKey]pb.PostParser_MultiPartClient // stream map
	LastSK repo.StreamKey
	CurSK  repo.StreamKey
	CurReq *pb.FileUploadReq
}

func NewTransmitter() Transmitter {
	connectString := "localhost:3100"
	conn, err := grpc.Dial(connectString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.L.Error(errs.Wrap(err, "rpc.Transmit.grpc.dial"))
	}
	client := pb.NewPostParserClient(conn)

	return &TransmitAdapter{
		c: client,
		M: make(map[repo.StreamKey]pb.PostParser_MultiPartClient),
	}
}
func (t *TransmitAdapter) Transmit(adu []repo.AppDistributorUnit) {
	//connectString := os.Getenv("HOST") + ":" + os.Getenv("PORT")

	for _, v := range adu {
		switch v.H.T {
		case repo.Unary:
			errs := t.transmitUnary(t.c, v)

			for _, rng := range errs {
				logger.L.Errorf("in grpc.Transmit err: %v\n", rng)
			}

		case repo.ClientStream:

			errs := t.transmitStream(t.c, v)
			for _, rng := range errs {
				logger.L.Errorf("in grpc.Transmit err: %v\n", rng)
			}

		}
	}
}

// update proto msg to use last flag
func (t *TransmitAdapter) transmitUnary(c pb.PostParserClient, aduOne repo.AppDistributorUnit) []error {
	errs, streamKyes := make([]error, 0), make([]repo.StreamKey, 0)
	req := t.NewReqUnary(aduOne)

	//logger.L.Infof("in grpc.transmitUnary for aduOne header %v, body %q, made req: %v\n", aduOne.H, aduOne.B.B, req)

	if aduOne.H.U.M.PostAction == repo.Finish {
		streamKyes, errs = t.Register(aduOne.H)

		for _, v := range streamKyes {

			_, err := t.M[v].CloseAndRecv()

			if err != nil {
				errs = append(errs, err)
			}

			delete(t.M, v)

		}
	}

	_, err := c.SinglePart(context.Background(), req)

	if err != nil {
		errs = append(errs, err)
	}
	//logger.L.Infof("in grpc.transmitUnary got res: %v\n", res)

	return errs
}

// handles stream ADUs, reads ans writes t.M
func (t *TransmitAdapter) transmitStream(c pb.PostParserClient, aduOne repo.AppDistributorUnit) []error {
	//logger.L.Warnf("grpc.transmitStream invoked with adu header %v, body %q, message %q, t.M: %v\n", aduOne.H, aduOne.B.B, aduOne.H.S.M, t.M)
	var (
		stream pb.PostParser_MultiPartClient
		err    error
	)
	errs := make([]error, 0)
	req := t.NewReqStream(aduOne)
	//techAduHeader := repo.AppDistributorHeader{T: repo.Tech, C: repo.CloseData{D: []repo.StreamKey{}}}

	streamKeyes, errs := t.Register(aduOne.H)
	/*
		logger.L.Infof("in grpc.transmitStream for aduOne with header %v and body %q and t.M %v called Register which returned streamKeys (%d) and errors (%d):\n", aduOne.H, aduOne.B.B, t.M, len(streamKeyes), len(errs))
		logger.L.Warnf("in grpc.transmitStream aduOne.H.S.M.PreAction = %d, aduOne.H.S.M.PostAction = %d\n", aduOne.H.S.M.PreAction, aduOne.H.S.M.PostAction)

		if len(streamKeyes) < 2 {
			errs = append(errs, fmt.Errorf("in parser.grpc.transmitStream len(streamKeys) < 2: %v", streamKeyes))
			return errs
		}

		for i, v := range streamKeyes {
			logger.L.Infof("in grpc.transmitStream streamKey i = %d, v: %v\n", i, v)
		}

		for i, v := range errs {
			logger.L.Infof("in grpc.transmitStream error i = %d, v: %v\n", i, v)
			if v != nil {
				errs = append(errs, v)
			}
		}
	*/
	switch pre := aduOne.H.S.M.PreAction; {
	case pre == repo.Start || pre == repo.Open:
		if pre == repo.Start {
			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, true)
		} else {
			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, false)
		}
		if err != nil {
			//	logger.L.Errorf("in grpc.transmitStream error: %v\n", err)
			errs = append(errs, err)
		}
		t.M[streamKeyes[0]] = stream

		//logger.L.Infof("in grpc.transmitStream preAction Start req: %v\n", req)
		err = stream.Send(req)

		if err != nil {
			errs = append(errs, err)
		}

	case pre == repo.Continue:
		//logger.L.Infof("in grpc.transmitStream in preAction none t.M: %v\n", t.M)
		stream, ok := t.M[streamKeyes[0]]
		if !ok {
			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, false)
			if err != nil {
				errs = append(errs, err)
			}
			t.M[streamKeyes[0]] = stream
		}
		//logger.L.Infof("in grpc.transmitStream preAction none req: %v\n", req)
		err = stream.Send(req)

		if err != nil {
			errs = append(errs, err)
		}

	case pre == repo.StopLast:

		if stream, ok := t.M[streamKeyes[0]]; ok {
			if aduOne.H.S.M.PostAction == repo.Finish {
				err := stream.Send(req)
				if err != nil {
					errs = append(errs, err)
				}
			}
			_, err = stream.CloseAndRecv()
			if err != nil {
				errs = append(errs, err)
			}
			delete(t.M, streamKeyes[0])
		}

	}
	//logger.L.Warnf("in grpc.transmitStream after PreAction t.M: %v\n", t.M)

	switch aduOne.H.S.M.PostAction {

	case repo.Continue:

		stream, ok := t.M[streamKeyes[0]]

		if ok {
			delete(t.M, streamKeyes[0])
		} else {

			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, false)

			if err != nil {
				errs = append(errs, err)
			}
			//logger.L.Infof("in grpc.transmitStream PostAction none req: %v\n", req)
			err = stream.Send(req)

			if err != nil {
				errs = append(errs, err)
			}
		}

		t.M[streamKeyes[1]] = stream

		//logger.L.Infof("in grpc.transmitStream after PostAction t.M: %v\n", t.M)

	case repo.Close:

		if stream, ok := t.M[streamKeyes[0]]; ok {
			_, err := stream.CloseAndRecv()
			if err != nil {
				errs = append(errs, err)
			}
			delete(t.M, streamKeyes[0])

		}

	case repo.Finish:
		for i := range t.M {
			if i.TS == aduOne.H.S.SK.TS {

				_, err := t.M[i].CloseAndRecv()

				if err != nil {
					errs = append(errs, err)
				}
				delete(t.M, i)
			}
		}
	}

	//logger.L.Infof("in grpc.transmitStream after handling t.M: %v\n", t.M)

	return errs

}

// Returns stream keys for handling ADUs based on ADU header, doesn't edit t.M
func (t *TransmitAdapter) Register(aduHeader repo.AppDistributorHeader) ([]repo.StreamKey, []error) {
	streamsPost, streamsPre, errs := make([]repo.StreamKey, 0), make([]repo.StreamKey, 0), make([]error, 0)
	//logger.L.Infof("rpc.Register invoked with aduHeader = %v, t.M = %v\n", aduHeader, t.M)

	switch aduHeader.T {
	case repo.Tech: // non-transmition handling
		if len(aduHeader.C.D) > 0 { //
			//logger.L.Infof("in rpc.Register before deleting t.M : %v\n", t.M)
			for _, v := range aduHeader.C.D {
				//logger.L.Infof("in rpc.Register trying to delete key : %v\n", v)
				delete(t.M, v)
			}
			//logger.L.Infof("in rpc.Register after deleting t.M : %v\n", t.M)
		}

	case repo.Unary: // unary transmition
		if aduHeader.U.M.PreAction == repo.StopLast { // unary transmition should close previous stream
			//logger.L.Infof("in rpc.Register M =%v\n", t.M)

			SK := repo.StreamKey{TS: aduHeader.U.UK.TS, Part: aduHeader.U.UK.Part - 1}
			//logger.L.Infof("in rpc.Register t.M = %v, SK = %v\n", t.M, SK)
			if _, ok := t.M[SK]; ok {
				delete(t.M, SK)
			} else {
				errs = append(errs, fmt.Errorf("in parser.Rpc.Register stream \"%v\" not found", SK))
			}
			// Stream to be closed is absent in map
		}
		if aduHeader.U.M.PostAction == repo.Finish { // unary transmition should clear all records of same TS in map and close corresponding streams
			//logger.L.Infof("in rpc.Register case = Finish\n")

			for i := range t.M {
				//logger.L.Infof("in rpc.Register i = %v\n", i)
				if i.TS == aduHeader.U.UK.TS {
					//logger.L.Infof("in rpc.Register t.M index = %v\n", i)
					streamsPre = append(streamsPre, i)
				}
			}

			for i := range streamsPre {

				old := t.M[streamsPre[i]]
				delete(t.M, streamsPre[i])
				//	streamsPre[i].Remove = true
				//logger.L.Infof("in rpc.Register deleteKey index = %d, value = %v\n", i, deleteKeys[i])
				t.M[streamsPre[i]] = old
			}

		}

	default: // stream transmition
		SK := repo.NewStreamKey(aduHeader.S.SK.TS, aduHeader.S.SK.Part, true)
		//SKCur := repo.StreamKey{TS: aduHeader.S.SK.TS, Part: aduHeader.S.SK.Part}
		SKPrev := SK
		SKPost := SK
		if aduHeader.S.SK.Part > 0 {
			SKPrev.Part--
		}
		SKPost.Part++
		//difference between none and stopLast
		if aduHeader.S.M.PreAction == repo.Start {
			streamsPre = append(streamsPre, SK.ToTrue())
		}

		if aduHeader.S.M.PreAction == repo.Open {
			//logger.L.Infof("in rpc.Register Preaction == none case t.M : %v, SKPrev: %v,  SKCur: %v\n", t.M, SKPrev, SKCur)

			streamsPre = append(streamsPre, SK.ToTrue())
			//logger.L.Infof("in grpc.Register none preaction %v added to streamsPre\n", SKCur)
		}

		if aduHeader.S.M.PreAction == repo.StopLast { // last strem should be closed and current data chunk should be added to new stream
			//logger.L.Infof("in rpc.Register stopLast case t.M : %v, SK: %v\n", t.M, SK.ToFalse())
			if _, ok := t.M[SK.ToFalse()]; ok {
				streamsPre = append(streamsPre, SK.ToFalse())
				//logger.L.Infof("in rpc.Register %v added to streamsPre\n", SKCur)

			}
		}
		if aduHeader.S.M.PreAction == repo.Continue {
			streamsPre = append(streamsPre, SK.ToFalse())
		}
		if aduHeader.S.M.PostAction == repo.Continue {
			//logger.L.Infof("in rpc.Register none case t.M : %v, SKPrev:%v,  SKCur: %v, SKPost: %v\n", t.M, SKPrev, SKCur, SKPost)

			streamsPost = append(streamsPost, SK.IncPart().ToFalse())
		}
		if aduHeader.S.M.PostAction == repo.Close {

			//logger.L.Infof("in rpc.Register repo.Stop case SKCur = %v\n", SKCur)

			streamsPost = append(streamsPost, SK.ToFalse())

		}
		if aduHeader.S.M.PostAction == repo.Finish {

			for i := range t.M {
				if i.TS == aduHeader.S.SK.TS && !i.N {
					streamsPost = append(streamsPost, i)
				}
			}
		}

	}
	sort.SliceStable(streamsPre, func(i, j int) bool { return streamsPre[i].Part < streamsPre[j].Part })
	streamsPost = append(streamsPre, streamsPost...)

	return streamsPost, errs

}

func (t *TransmitAdapter) NewStream(ts, fo, fi string, f bool) (pb.PostParser_MultiPartClient, error) {
	reqInit := &pb.FileUploadReq{
		Info: &pb.FileUploadReq_FileInfo{
			FileInfo: &pb.FileInfo{
				Ts:        ts,
				IsFirst:   f,
				FieldName: fo,
				FileName:  fi,
			},
		},
	}
	newStream, err := t.c.MultiPart(context.Background())
	if err != nil {
		return nil, err
	}
	//logger.L.Infof("in grpc reqInit %v\n", reqInit)
	err = newStream.Send(reqInit)
	if err != nil {
		return nil, err
	}
	return newStream, nil
}

func (t *TransmitAdapter) NewReqUnary(aduOne repo.AppDistributorUnit) *pb.TextFieldReq {
	req := &pb.TextFieldReq{}

	req.Ts = aduOne.H.U.UK.TS
	req.Name = aduOne.H.U.F.FormName
	req.ByteChunk = aduOne.B.B

	if aduOne.H.U.F.FileName != "" {
		req.Filename = aduOne.H.U.F.FileName
	}
	if aduOne.H.U.M.PreAction == repo.Start {
		req.IsFirst = true
	}

	if aduOne.H.U.M.PostAction == repo.Finish {
		req.IsLast = true
	}
	return req
}

func (t *TransmitAdapter) NewReqStream(aduOne repo.AppDistributorUnit) *pb.FileUploadReq {

	FD, I, req := &pb.FileData{}, &pb.FileUploadReq_FileData{}, &pb.FileUploadReq{}

	FD.ByteChunk = aduOne.B.B

	if aduOne.H.S.M.PostAction == repo.Finish {
		FD.IsLast = true
	}

	I.FileData = FD
	req.Info = I

	return req
}
