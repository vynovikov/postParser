package rpc

import (
	"context"
	"fmt"
	"os"
	tosaver "postParser/internal/adapters/driven/rpc/tosaver/pb"
	"sort"
	"sync"

	tologger "postParser/internal/adapters/driven/rpc/tologger/pb"
	"postParser/internal/logger"
	"postParser/internal/repo"

	errs "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var (
	Stream tosaver.Saver_MultiPartClient
	req    *tosaver.FileUploadReq
	err    error
)

type Transmitter interface {
	Transmit(repo.AppDistributorUnit, *sync.Mutex)
	Log(string) []error
}

type TransmitAdapter struct {
	saverClient  tosaver.SaverClient
	loggerClient tologger.LoggerClient
	M            map[repo.StreamKey]tosaver.Saver_MultiPartClient // stream map
	LastSK       repo.StreamKey
	CurSK        repo.StreamKey
	CurReq       *tosaver.FileUploadReq
	lock         sync.Mutex
}

func NewTransmitter(lis *bufconn.Listener) *TransmitAdapter {
	//connectStringToSaver := "localhost:3100"
	var (
		connectStringToSaver  string
		connectStringToLogger string
		connToSaver           *grpc.ClientConn
		connToLogger          *grpc.ClientConn
		err                   error
	)

	saverHostName, loggerHostName := os.Getenv(("SAVER_HOSTNAME")), os.Getenv(("LOGGER_HOSTNAME"))
	if len(saverHostName) == 0 {
		saverHostName = "localhost"
	}
	if len(loggerHostName) == 0 {
		loggerHostName = "localhost"
	}

	connectStringToSaver = saverHostName + ":" + "3100"
	connToSaver, err = grpc.Dial(connectStringToSaver, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.L.Error(errs.Wrap(err, "rpc.Transmit.grpc.dial"))
	}
	connectStringToLogger = "postlogger" + ":" + "3200"
	connToLogger, err = grpc.Dial(connectStringToLogger, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.L.Error(errs.Wrap(err, "rpc.Transmit.grpc.dial"))
	}

	//client := tosaver.NewPostParserClient(conn)

	return &TransmitAdapter{
		saverClient:  tosaver.NewSaverClient(connToSaver),
		loggerClient: tologger.NewLoggerClient(connToLogger),
		M:            make(map[repo.StreamKey]tosaver.Saver_MultiPartClient),
	}
}

func (t *TransmitAdapter) Transmit(adu repo.AppDistributorUnit, mu *sync.Mutex) {
	//connectString := os.Getenv("HOST") + ":" + os.Getenv("PORT")

	switch adu.H.T {
	case repo.Unary:
		//logger.L.Warnf("in grpc.Transmit trying to send via unary adu header %v, body %q\n", adu.H, adu.B.B)
		t.transmitUnary(t.saverClient, adu, mu)
		/*
			for _, rng := range errs {
				logger.L.Errorf("in grpc.Transmit err: %v\n", rng)
			}
		*/
	case repo.ClientStream:

		//logger.L.Warnf("in grpc.Transmit trying to send via stream adu header %v, body %q\n", adu.H, adu.B.B)
		t.transmitStream(t.saverClient, adu, mu)
		/*
			for _, rng := range errs {
				logger.L.Errorf("in grpc.Transmit err: %v\n", rng)
			}
		*/
		//logger.L.Errorf("in grpc.Transmit adu header %v, body %q was sent\n", adu.H, adu.B.B)
		//time.Sleep(time.Millisecond * 10)

	}

}
func (t *TransmitAdapter) Log(s string) []error {

	errs := make([]error, 0)
	req := &tologger.LogReq{}

	//logger.L.Infof("in rpc.Log trying to log %q\n", s)
	req.Ts = repo.NewTS()
	req.LogString = s

	_, err := t.loggerClient.Log(context.Background(), req)
	if err != nil {
		logger.L.Errorf("in rpc.Log error %v\n", err)
		errs = append(errs, err)
	}
	return errs
}

// update proto msg to use last flag
func (t *TransmitAdapter) transmitUnary(c tosaver.SaverClient, aduOne repo.AppDistributorUnit, mu *sync.Mutex) []error {
	//logger.L.Infof("grpc.transmitUnary invoked for aduOne header %v, body %q \n", aduOne.H, aduOne.GetBody())
	errs, streamKyes := make([]error, 0), make([]repo.StreamKey, 0)
	if aduOne.H.U.M.PreAction == repo.Start {
		defer mu.Unlock()
	}
	req := t.NewReqUnary(aduOne)

	//logger.L.Infof("in grpc.transmitUnary aduOne header %v, body %q made req %v\n", aduOne.H, aduOne.GetBody(), req)

	if aduOne.H.U.M.PostAction == repo.Finish {

		streamKyes, errs = t.Register(aduOne.H)
		//logger.L.Infof("in grpc.transmitUnary aduOne header %v, body %q during t.M %v got streamKeys: %v, errs %v\n", aduOne.H, aduOne.GetBody(), t.M, streamKyes, errs)

		for _, v := range streamKyes {

			if aduOne.H.U.M.PreAction == repo.StopLast {

				//logger.L.Infof("in grpc.transmitUnary aduOne header %v, body %q trying to close stream by %v\n", aduOne.H, aduOne.GetBody(), v)
				_, err := t.M[v].CloseAndRecv()
				//logger.L.Infoln("in grpc.transmitUnary success")
				if err != nil {
					errs = append(errs, err)
				}
			}

			delete(t.M, v)

		}
		//logger.L.Infof("in grpc.transmitUnary aduOne header %v, body %q left t.M: %v\n", aduOne.H, aduOne.GetBody(), t.M)
	}
	//logger.L.Infof("in grpc.transmitUnary aduOne header %v, body %q unlocking\n", aduOne.H, aduOne.GetBody())
	if aduOne.H.U.M.PreAction != repo.Start {
		//logger.L.Errorf("in grpc.transmitUnary aduOne with header: %v, unlocks mutex\n", aduOne.H)
		mu.Unlock()
	}

	//logger.L.Infof("in grpc.transmitUnary aduOne header %v, body %q req %v sending\n", aduOne.H, aduOne.GetBody(), req)
	_, err := c.SinglePart(context.Background(), req)

	if err != nil {
		errs = append(errs, err)
	}
	//logger.L.Infof("in grpc.transmitUnary got res: %v\n", res)

	//logger.L.Warnf("in grpc.transmitUnary for aduOne header %v after handling t.M: %v\n", aduOne.H, t.M)

	return errs
}

// handles stream ADUs, reads ans writes t.M
func (t *TransmitAdapter) transmitStream(c tosaver.SaverClient, aduOne repo.AppDistributorUnit, mu *sync.Mutex) []error {
	//logger.L.Infof("grpc.transmitStream invoked with adu header %v, message %q, t.M: %v\n", aduOne.H, aduOne.H.S.M, t.M)
	var (
		stream tosaver.Saver_MultiPartClient
		err    error
	)
	errs := make([]error, 0)

	if aduOne.H.S.M.PreAction == repo.Start {
		defer mu.Unlock()
	}
	req := t.NewReqStream(aduOne)
	//techAduHeader := repo.AppDistributorHeader{T: repo.Tech, C: repo.CloseData{D: []repo.StreamKey{}}}

	streamKeyes, errs := t.Register(aduOne.H)
	//logger.L.Infof("in grpc.transmitStream for adu header %v, streamKeys are %v\n", aduOne.H, streamKeyes)

	switch pre := aduOne.H.S.M.PreAction; {
	case pre == repo.Start || pre == repo.Open:

		if pre == repo.Start {
			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, true)
		} else {
			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, false)
		}
		if err != nil {
			logger.L.Errorf("in grpc.transmitStream error: %v\n", err)
			errs = append(errs, err)
		}
		t.M[streamKeyes[0]] = stream

		if aduOne.H.S.M.PreAction != repo.Start {
			mu.Unlock()
		}
		//logger.L.Infof("in grpc.transmitStream for adu header %v preAction = repo.Start or repo.Open sending req: %v\n", aduOne.H, req)
		err = stream.Send(req)
		//logger.L.Infof("in grpc.transmitStream for adu header %v preAction = repo.Start or repo.Open req: %v has been sent to SK %v, t.M became %v\n", aduOne.H, req, streamKeyes[0], t.M)

		if err != nil {
			errs = append(errs, err)
		}

	case pre == repo.Continue:
		//logger.L.Infof("in grpc.transmitStream in during handling adu header %v case preAction continue t.M: %v\n", aduOne.H, t.M)
		stream, ok := t.M[streamKeyes[0]]
		//logger.L.Infof("in grpc.transmitStream for adu header %v preAction = continue req: %v has been sent to SK %v, t.M became %v\n", aduOne.H, req, streamKeyes[0], t.M)
		if !ok {
			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, false)
			if err != nil {
				errs = append(errs, err)
			}
			t.M[streamKeyes[0]] = stream
		}

		//logger.L.Infof("in grpc.transmitStream adu header %v preAction continue sending req: %v\n", aduOne.H, req)

		err = stream.Send(req)
		mu.Unlock()

		if err != nil {
			errs = append(errs, err)
			//logger.L.Errorf("in grpc.transmitStream unable to send req: %v\n", err)
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
	//logger.L.Infof("in grpc.transmitStream after PreAction t.M: %v\n", t.M)
	//logger.L.Infof("in grpc.transmitStream adu header %v has postaction repo.Continue? %t\n", aduOne.H, aduOne.H.S.M.PostAction == repo.Continue)

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
			//logger.L.Infoln("in grpc.transmitStream adu header %v checking for preaction \n", aduOne.H)
			if aduOne.H.S.M.PreAction == repo.StopLast || aduOne.H.S.M.PreAction == repo.Continue {
				//logger.L.Infof("in grpc.transmitStream adu header %v unlocks \n", aduOne.H)
				mu.Unlock()
			}

			if err != nil {
				errs = append(errs, err)
			}
		}

		t.M[streamKeyes[1]] = stream

		//logger.L.Infof("in grpc.transmitStream for adu header %v after PostAction == Continue t.M: %v\n", aduOne.H, t.M)

	case repo.Close:
		//logger.L.Infof("in grpc.transmitStream for adu header %v, streamKeys are %v\n", aduOne.H, streamKeyes)

		if stream, ok := t.M[streamKeyes[0]]; ok {
			_, err := stream.CloseAndRecv()
			if err != nil {
				errs = append(errs, err)
			}
			t.Delete(streamKeyes[0])
			//delete(t.M, streamKeyes[0])

		}

	case repo.Finish:
		t.lock.Lock()
		for i := range t.M {
			if i.TS == aduOne.H.S.SK.TS {

				_, err := t.M[i].CloseAndRecv()

				if err != nil {
					errs = append(errs, err)
				}
				delete(t.M, i)
			}
		}
		t.lock.Unlock()
	}

	//logger.L.Infof("in grpc.transmitStream after handling adu header %v t.M: %v\n", aduOne.H, t.M)

	return errs

}
func (t *TransmitAdapter) Delete(streamKey repo.StreamKey) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.M, streamKey)
}

// Returns stream keys for handling ADUs based on ADU header, doesn't edit t.M
func (t *TransmitAdapter) Register(aduHeader repo.AppDistributorHeader) ([]repo.StreamKey, []error) {
	streamsPost, streamsPre, errs := make([]repo.StreamKey, 0), make([]repo.StreamKey, 0), make([]error, 0)
	//logger.L.Infof("rpc.Register invoked with aduHeader = %v, t.M = %v\n", aduHeader, t.M)

	switch aduHeader.T {
	/*	case repo.Tech: // non-transmition handling
		if len(aduHeader.C.D) > 0 { //
			//logger.L.Infof("in rpc.Register before deleting t.M : %v\n", t.M)
			for _, v := range aduHeader.C.D {
				//logger.L.Infof("in rpc.Register trying to delete key : %v\n", v)
				delete(t.M, v)
			}
			//logger.L.Infof("in rpc.Register after deleting t.M : %v\n", t.M)
		}
	*/
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
			//logger.L.Infof("in grpc.Register start preaction %v added to streamsPre\n", SKCur)
			//logger.L.Infof("in grpc.Register adu header %v have pre = start\n", aduHeader)
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

func (t *TransmitAdapter) NewStream(ts, fo, fi string, f bool) (tosaver.Saver_MultiPartClient, error) {
	reqInit := &tosaver.FileUploadReq{
		Info: &tosaver.FileUploadReq_FileInfo{
			FileInfo: &tosaver.FileInfo{
				Ts:        ts,
				IsFirst:   f,
				FieldName: fo,
				FileName:  fi,
			},
		},
	}
	newStream, err := t.saverClient.MultiPart(context.Background())
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

func (t *TransmitAdapter) NewReqUnary(aduOne repo.AppDistributorUnit) *tosaver.TextFieldReq {
	req := &tosaver.TextFieldReq{}

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

func (t *TransmitAdapter) NewReqStream(aduOne repo.AppDistributorUnit) *tosaver.FileUploadReq {

	FD, I, req := &tosaver.FileData{}, &tosaver.FileUploadReq_FileData{}, &tosaver.FileUploadReq{}

	FD.Ts = aduOne.H.S.SK.TS
	FD.FieldName = aduOne.H.S.F.FormName
	FD.ByteChunk = aduOne.B.B
	//logger.L.Infof("in rpc.NewReqStream adu header %v has SK.Part = %d, SK.B,Part = %d\n", aduOne.H, aduOne.H.S.SK.Part, aduOne.H.S.B.Part)
	FD.Number = uint32(aduOne.H.S.SK.Part - aduOne.H.S.B.Part)

	if aduOne.H.S.M.PostAction == repo.Finish {
		FD.IsLast = true
	}

	I.FileData = FD
	req.Info = I

	//logger.L.Infof("in rpc.NewReqStream adu header %v body %q converted into req number %d\n", aduOne.H, aduOne.B.B, req.GetFileData().Number)

	return req
}

func DecodeUnaryReq(r *tosaver.TextFieldReq) repo.GRequest {
	res := repo.GRequest{}

	res.FieldName = r.Name
	res.FileName = r.Filename
	res.ByteChunk = r.ByteChunk
	res.IsFirst = r.IsFirst
	res.IsLast = r.IsLast
	res.RType = repo.U

	return res
}

func DecodeStreamInfoReq(r *tosaver.FileUploadReq) repo.GRequest {
	res := repo.GRequest{}

	info := r.GetFileInfo()
	res.FieldName = info.FieldName
	res.FileInfo = true
	res.FileName = info.FileName
	res.IsFirst = info.IsFirst
	res.RType = repo.S
	//logger.L.Infof("in rpc.DecodeStreamInfoReq %v decoded into  %v\n", r, res)

	return res
}

func DecodeStreamDataReq(r *tosaver.FileUploadReq, info *tosaver.FileInfo) repo.GRequest {
	res := repo.GRequest{}

	res.FileData = true
	data := r.GetFileData()
	res.FieldName = info.FieldName
	res.ByteChunk = data.ByteChunk
	res.IsLast = data.IsLast
	res.RType = repo.S

	return res
}
