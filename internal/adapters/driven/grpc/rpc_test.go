package grpc

import (
	"errors"
	"fmt"
	"postParser/internal/adapters/driven/grpc/pb"
	"postParser/internal/repo"
	"testing"

	"github.com/stretchr/testify/suite"
)

var stream pb.PostParser_MultiPartClient

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
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 0}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{C: repo.CurrentPieceHeader{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.None}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{},
			},
			wantErr: []error{},
			wantSK:  []repo.StreamKey{},
		},

		{
			name: "unary error stream not found",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 0}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{C: repo.CurrentPieceHeader{TS: "qqq", Part: 2}, M: repo.Message{PreAction: repo.StopLast}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 0}: stream,
				},
			},
			wantErr: []error{
				errors.New("in parser.Rpc.Register stream \"{qqq 1 false}\" not found"),
			},
			wantSK: []repo.StreamKey{},
		},
		/*
			{
				name: "unary prevAction == repo.Finish => setting Remove flag",
				T: TransmitAdapter{
					M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
						{TS: "qqq", Part: 0}: stream,
						{TS: "qqq", Part: 1}: stream,
						{TS: "qqq", Part: 2}: stream,
					},
				},
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{C: repo.CurrentPieceHeader{TS: "qqq", Part: 2}, M: repo.Message{PostAction: repo.Finish}}},
				wantT: TransmitAdapter{
					M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
						{TS: "qqq", Part: 0, Remove: true}: stream,
						{TS: "qqq", Part: 1, Remove: true}: stream,
						{TS: "qqq", Part: 2, Remove: true}: stream,
					},
				},
				wantSK: []repo.StreamKey{
					{TS: "qqq", Part: 0, Remove: true},
					{TS: "qqq", Part: 1, Remove: true},
					{TS: "qqq", Part: 2, Remove: true},
				},
				wantErr: []error{},
			},

			{
				name: "tech deleting from map all same TS records",
				T: TransmitAdapter{
					M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
						{TS: "qqq", Part: 0}: stream,
						{TS: "qqq", Part: 1}: stream,
						{TS: "qqq", Part: 2}: stream,
						{TS: "www", Part: 3}: stream,
					},
				},
				H: repo.AppDistributorHeader{T: repo.Tech, C: repo.CloseData{D: []repo.StreamKey{{TS: "qqq", Part: 0}, {TS: "qqq", Part: 1}, {TS: "qqq", Part: 2}}}},
				wantT: TransmitAdapter{
					M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
						{TS: "www", Part: 3}: stream,
					},
				},
				wantErr: []error{},
				wantSK:  []repo.StreamKey{},
			},
		*/
		{
			name: "clientStream prevAction == repo.None, postAction = repo.None, t.M has current stream key",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1},
				{TS: "qqq", Part: 2},
			},
		},

		{
			name: "clientStream prevAction == repo.None, postAction = repo.None, t.M hasn't current stream key",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1},
				{TS: "qqq", Part: 2},
			},
		},

		{
			name: "clientStream prevAction == repo.StopLast, postAction = repo.None",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 2}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.None}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 2}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 2},
				{TS: "qqq", Part: 3},
			},
		},

		{
			name: "clientStream prevAction == repo.StopLast, postAction = repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Finish}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1},
				{TS: "qqq", Part: 1},
			},
		},

		{
			name: "clientStream prevAction == repo.None, postAction = repo.Stop",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1},
				{TS: "qqq", Part: 1},
			},
		},

		{
			name: "clientStream prevAction == repo.None, postAction = repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
					{TS: "www", Part: 0}: stream,
				},
			},
			H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}},
			wantT: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
					{TS: "www", Part: 0}: stream,
				},
			},
			wantErr: []error{},
			wantSK: []repo.StreamKey{
				{TS: "qqq", Part: 1},
				{TS: "qqq", Part: 1},
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
			name: "preAction: repo.None postAction: repo.None, stream exists",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						ByteChunk: []byte("azaza"),
					},
				},
			},
		},

		{
			name: "preAction: repo.None postAction: repo.None, stream not exists",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						ByteChunk: []byte("azaza"),
					},
				},
			},
		},

		{
			name: "preAction: repo.None postAction: repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						ByteChunk: []byte("azaza"),
						IsLast:    true,
					},
				},
			},
		},

		{
			name: "preAction: repo.StopLast postAction: repo.None",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},

			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
						ByteChunk: []byte("azaza"),
					},
				},
			},
		},

		{
			name: "preAction: repo.StopLast postAction: repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Finish}}}},

			wantReq: &pb.FileUploadReq{
				Info: &pb.FileUploadReq_FileData{
					FileData: &pb.FileData{
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
		wantReq *pb.TextFieldsReq
	}{

		{
			name: "preAction: repo.None postAction: repo.None no filename",
			T:    TransmitAdapter{},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldsReq{
				Ts:        "qqq",
				Name:      "alice",
				ByteChunk: []byte("azaza"),
			},
		},

		{
			name: "preAction: repo.None postAction: repo.None filename",
			T:    TransmitAdapter{},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldsReq{
				Ts:        "qqq",
				Name:      "alice",
				Filename:  "short.txt",
				ByteChunk: []byte("azaza"),
			},
		},

		{
			name: "preAction: repo.None postAction: repo.Finish",
			T: TransmitAdapter{
				M: map[repo.StreamKey]pb.PostParser_MultiPartClient{
					{TS: "qqq", Part: 1}: stream,
					{TS: "www", Part: 1}: stream,
				},
			},
			aduOne: repo.AppDistributorUnit{
				H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azaza")},
			},
			wantReq: &pb.TextFieldsReq{
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
