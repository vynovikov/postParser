// gRPC transmitter.
// Transmits data to saver service and logs to logger service.
package rpc

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	tosaver "github.com/vynovikov/postParser/internal/adapters/driven/rpc/tosaver/pb"

	tologger "github.com/vynovikov/postParser/internal/adapters/driven/rpc/tologger/pb"
	"github.com/vynovikov/postParser/internal/logger"
	"github.com/vynovikov/postParser/internal/repo"

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
	Log(string) error
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
	connectStringToLogger = loggerHostName + ":" + "3200"
	connToLogger, err = grpc.Dial(connectStringToLogger, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.L.Error(errs.Wrap(err, "rpc.Transmit.grpc.dial"))
	}

	return &TransmitAdapter{
		saverClient:  tosaver.NewSaverClient(connToSaver),
		loggerClient: tologger.NewLoggerClient(connToLogger),
		M:            make(map[repo.StreamKey]tosaver.Saver_MultiPartClient),
	}
}

func (t *TransmitAdapter) Transmit(adu repo.AppDistributorUnit, mu *sync.Mutex) {

	switch adu.H.T {
	case repo.Unary:
		t.transmitUnary(t.saverClient, adu, mu)
	case repo.ClientStream:
		t.transmitStream(t.saverClient, adu, mu)
	}

}
func (t *TransmitAdapter) Log(s string) error {
	req := &tologger.LogReq{}
	req.Ts = repo.NewTS()
	req.LogString = s

	_, err := t.loggerClient.Log(context.Background(), req)
	if err != nil {
		return err
	}
	return nil
}

// transmitUnary handles unary-type ADUs
func (t *TransmitAdapter) transmitUnary(c tosaver.SaverClient, aduOne repo.AppDistributorUnit, mu *sync.Mutex) []error {
	errs, streamKyes := make([]error, 0), make([]repo.StreamKey, 0)
	if aduOne.H.U.M.PreAction == repo.Start {
		defer mu.Unlock()
	}
	req := t.NewReqUnary(aduOne)

	if aduOne.H.U.M.PostAction == repo.Finish {

		streamKyes, errs = t.Register(aduOne.H)

		for _, v := range streamKyes {

			if aduOne.H.U.M.PreAction == repo.StopLast {
				_, err := t.M[v].CloseAndRecv()
				if err != nil {
					errs = append(errs, err)
				}
			}
			delete(t.M, v)
		}
	}
	if aduOne.H.U.M.PreAction != repo.Start {
		mu.Unlock()
	}
	_, err := c.SinglePart(context.Background(), req)

	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// transmitStream handles stream-type ADUs
// Updates t.M each time
func (t *TransmitAdapter) transmitStream(c tosaver.SaverClient, aduOne repo.AppDistributorUnit, mu *sync.Mutex) []error {
	var (
		stream tosaver.Saver_MultiPartClient
		err    error
	)
	errs := make([]error, 0)

	if aduOne.H.S.M.PreAction == repo.Start {
		defer mu.Unlock()
	}
	req := t.NewReqStream(aduOne)
	streamKeyes, errs := t.Register(aduOne.H)

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

		err = stream.Send(req)

		if err != nil {
			errs = append(errs, err)
		}

	case pre == repo.Continue:
		stream, ok := t.M[streamKeyes[0]]

		if !ok {
			stream, err = t.NewStream(aduOne.H.S.SK.TS, aduOne.H.S.F.FormName, aduOne.H.S.F.FileName, false)
			if err != nil {
				errs = append(errs, err)
			}
			t.M[streamKeyes[0]] = stream
		}
		err = stream.Send(req)
		mu.Unlock()

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
			err = stream.Send(req)
			if aduOne.H.S.M.PreAction == repo.StopLast || aduOne.H.S.M.PreAction == repo.Continue {
				mu.Unlock()
			}

			if err != nil {
				errs = append(errs, err)
			}
		}

		t.M[streamKeyes[1]] = stream

	case repo.Close:

		if stream, ok := t.M[streamKeyes[0]]; ok {
			_, err := stream.CloseAndRecv()
			if err != nil {
				errs = append(errs, err)
			}
			t.Delete(streamKeyes[0])
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
	return errs

}
func (t *TransmitAdapter) Delete(streamKey repo.StreamKey) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.M, streamKey)
}

// Register determines which stream keys for ADU handling should be used.
// Doesn't modify t.M.
// Tested in rpc_test.go
func (t *TransmitAdapter) Register(aduHeader repo.AppDistributorHeader) ([]repo.StreamKey, []error) {
	streamsPost, streamsPre, errs := make([]repo.StreamKey, 0), make([]repo.StreamKey, 0), make([]error, 0)

	switch aduHeader.T {

	case repo.Unary: // unary transmition

		if aduHeader.U.M.PreAction == repo.StopLast { // unary transmition should close previous stream

			SK := repo.StreamKey{TS: aduHeader.U.UK.TS, Part: aduHeader.U.UK.Part - 1}
			if _, ok := t.M[SK]; ok {
				delete(t.M, SK)
			} else {
				errs = append(errs, fmt.Errorf("in parser.Rpc.Register stream \"%v\" not found", SK))
			}
		}
		if aduHeader.U.M.PostAction == repo.Finish { // unary transmition should clear all records of same TS in map and close corresponding streams

			for i := range t.M {
				if i.TS == aduHeader.U.UK.TS {
					streamsPre = append(streamsPre, i)
				}
			}

			for i := range streamsPre {
				old := t.M[streamsPre[i]]
				delete(t.M, streamsPre[i])
				t.M[streamsPre[i]] = old
			}
		}

	default: // stream transmition
		SK := repo.NewStreamKey(aduHeader.S.SK.TS, aduHeader.S.SK.Part, true)

		SKPrev := SK
		SKPost := SK
		if aduHeader.S.SK.Part > 0 {
			SKPrev.Part--
		}
		SKPost.Part++

		if aduHeader.S.M.PreAction == repo.Start {

			streamsPre = append(streamsPre, SK.ToTrue())
		}

		if aduHeader.S.M.PreAction == repo.Open {

			streamsPre = append(streamsPre, SK.ToTrue())
		}

		if aduHeader.S.M.PreAction == repo.StopLast { // last strem should be closed and current data chunk should be added to new stream
			if _, ok := t.M[SK.ToFalse()]; ok {

				streamsPre = append(streamsPre, SK.ToFalse())
			}
		}
		if aduHeader.S.M.PreAction == repo.Continue {

			streamsPre = append(streamsPre, SK.ToFalse())
		}
		if aduHeader.S.M.PostAction == repo.Continue {

			streamsPost = append(streamsPost, SK.IncPart().ToFalse())
		}
		if aduHeader.S.M.PostAction == repo.Close {

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
	err = newStream.Send(reqInit)
	if err != nil {
		return nil, err
	}
	return newStream, nil
}

// NewRqUnary returns request for unary transmission.
// Tested in rpc_test.go
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

// NewReqStream returns request for stream transmission.
// Tested in rpc_test.go
func (t *TransmitAdapter) NewReqStream(aduOne repo.AppDistributorUnit) *tosaver.FileUploadReq {

	FD, I, req := &tosaver.FileData{}, &tosaver.FileUploadReq_FileData{}, &tosaver.FileUploadReq{}

	FD.Ts = aduOne.H.S.SK.TS
	FD.FieldName = aduOne.H.S.F.FormName
	FD.ByteChunk = aduOne.B.B
	FD.Number = uint32(aduOne.H.S.SK.Part - aduOne.H.S.B.Part)

	if aduOne.H.S.M.PostAction == repo.Finish {
		FD.IsLast = true
	}

	I.FileData = FD
	req.Info = I

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
