package rpc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/vynovikov/postParser/internal/adapters/driven/rpc/tosaver/pb"
	"github.com/vynovikov/postParser/internal/repo"

	"github.com/stretchr/testify/suite"
)

var stream pb.Saver_MultiPartClient

type rpcSuite struct {
	suite.Suite
}

func TestRpcSuite(t *testing.T) {
	suite.Run(t, new(rpcSuite))
}
func (s *rpcSuite) SetupTest() {

}

func (s *rpcSuite) TestRegister() {
	tt := []struct {
		name    string
		T       TransmitAdapter
		H       repo.AppDistributorHeader
		wantT   TransmitAdapter
		wantErr []error
		wantSK  []repo.StreamKey
	}{
		{
			name: "unary stopLast",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 0}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{UK: repo.UnaryKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.None}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{},
			},
			wantErr: []error{},
			wantSK:  []repo.StreamKey{},
		},

		{
			name: "unary error stream not found",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{UK: repo.UnaryKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			wantErr: []error{
				errors.New("in parser.Rpc.Register stream \"{qqq 0 false}\" not found"),
			},
			wantSK: []repo.StreamKey{},
		},

		{
			name: "clientStream prevAction == repo.Open, postAction = repo.Continue, t.M countains adu's part stream => returning keys to open and continue",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.Open, PostAction: repo.Continue}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1, N: true},
				{TS: "qqq", Part: 2, N: false},
			},
		},

		{
			name: "clientStream prevAction == repo.Start, postAction = repo.Continue => returning keys to open and continue",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.Open, PostAction: repo.Continue}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1, N: true},
				{TS: "qqq", Part: 2, N: false},
			},
		},

		{
			name: "clientStream prevAction == repo.Continue, postAction = repo.Continue => returning keys to open and continue",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1, N: false},
				{TS: "qqq", Part: 2, N: false},
			},
		},

		{
			name: "clientStream prevAction == repo.StopLast, postAction = repo.Continue",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 2, N: false}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Continue}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 2, N: false}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 2, N: false},
				{TS: "qqq", Part: 3, N: false},
			},
		},

		{
			name: "clientStream prevAction == repo.StopLast, postAction = repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Finish}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1, N: false},
				{TS: "qqq", Part: 1, N: false},
			},
		},

		{
			name: "clientStream prevAction == repo.Continue, postAction = repo.Close",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Close}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1, N: false},
				{TS: "qqq", Part: 1, N: false},
			},
		},

		{
			name: "clientStream prevAction == repo.Continue, postAction = repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
					{TS: "www", Part: 0, N: false}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Finish}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1, N: false}: stream,
					{TS: "www", Part: 0, N: false}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1, N: false},
				{TS: "qqq", Part: 1, N: false},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotSK, gotErr := v.T.Register(v.H)
			s.Equal(v.wantT, v.T)
			s.Equal(v.wantSK, gotSK)
			s.Equal(v.wantErr, gotErr)
		})
	}
}

func (s *rpcSuite) TestNewReqStream() {
	tt := []struct {
		name    string
		T       TransmitAdapter
		aduOne  repo.AppDistributorUnit
		wantReq *pb.FileUploadReq
	}{
		{
			name: "preAction: repo.Start postAction: repo.Continue, S.SK.Part - S.B.Part = 0",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Start, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						Ts:        "qqq",
						FieldName: "alice",
						ByteChunk: []byte("azaza"),
					},
				},
			},
		},
		{
			name: "preAction: repo.Start postAction: repo.Continue, S.SK.Part - S.B.Part = 1",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Start, PostAction: repo.Continue}, B: repo.BeginningData{Part: 1}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						Ts:        "qqq",
						FieldName: "alice",
						ByteChunk: []byte("azaza"),
						Number:    1,
					},
				},
			},
		},

		{
			name: "preAction: repo.Continue postAction: repo.Continue, stream exists",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						Ts:        "qqq",
						FieldName: "alice",
						Number:    1,
						ByteChunk: []byte("azaza"),
					},
				},
			},
		},

		{
			name: "preAction: repo.Open postAction: repo.Continue, stream not exists",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Open, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						Ts:        "qqq",
						FieldName: "alice",
						Number:    1,
						ByteChunk: []byte("azaza"),
					},
				},
			},
		},
		{
			name: "preAction: repo.Continue postAction: repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						Ts:        "qqq",
						FieldName: "alice",
						Number:    1,
						ByteChunk: []byte("azaza"),
						IsLast:    true,
					},
				},
			},
		},

		{
			name: "preAction: repo.StopLast postAction: repo.Continue",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},

			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						Ts:        "qqq",
						Number:    1,
						ByteChunk: []byte("azaza"),
					},
				},
			},
		},

		{
			name: "preAction: repo.StopLast postAction: repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Finish}}}},

			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						Ts:     "qqq",
						Number: 1,
						IsLast: true,
					},
				},
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {

			gotReq := v.T.NewReqStream(v.aduOne)
			s.Equal(fmt.Sprint(v.wantReq), fmt.Sprint(gotReq))

		})
	}
}
func (s *rpcSuite) TestNewReqUnary() {
	tt := []struct {
		name    string
		T       TransmitAdapter
		aduOne  repo.AppDistributorUnit
		wantReq *pb.TextFieldReq
	}{

		{
			name: "preAction: repo.None postAction: repo.None no filename",
			T:    TransmitAdapter{},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{UK: repo.UnaryKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldReq{
				Ts:        "qqq",
				Name:      "alice",
				ByteChunk: []byte("azaza"),
			},
		},

		{
			name: "preAction: repo.None postAction: repo.None filename",
			T:    TransmitAdapter{},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{UK: repo.UnaryKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldReq{
				Ts:        "qqq",
				Name:      "alice",
				Filename:  "short.txt",
				ByteChunk: []byte("azaza"),
			},
		},

		{
			name: "preAction: repo.Start postAction: repo.None",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
					{TS: "www", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{UK: repo.UnaryKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Start, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldReq{
				Ts:        "qqq",
				Name:      "alice",
				Filename:  "short.txt",
				ByteChunk: []byte("azaza"),
				IsFirst:   true,
			},
		},

		{
			name: "preAction: repo.Start postAction: repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
					{TS: "www", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{UK: repo.UnaryKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Start, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldReq{
				Ts:        "qqq",
				Name:      "alice",
				Filename:  "short.txt",
				ByteChunk: []byte("azaza"),
				IsFirst:   true,
				IsLast:    true,
			},
		},

		{
			name: "preAction: repo.Continue postAction: repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.Saver_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
					{TS: "www", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{UK: repo.UnaryKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldReq{
				Ts:        "qqq",
				Name:      "alice",
				Filename:  "short.txt",
				ByteChunk: []byte("azaza"),
				IsLast:    true,
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {

			gotReq := v.T.NewReqUnary(v.aduOne)
			s.Equal(fmt.Sprint(v.wantReq), fmt.Sprint(gotReq))

		})
	}
}
