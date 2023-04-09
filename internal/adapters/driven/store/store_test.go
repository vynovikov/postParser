package store

import (
	"errors"
	"postParser/internal/repo"
	"testing"

	"github.com/stretchr/testify/suite"
)

type storeSuite struct {
	suite.Suite
}

func TestStore(t *testing.T) {
	suite.Run(t, new(storeSuite))
}

func (s *storeSuite) TestRegister() {
	tt := []struct {
		name      string
		store     Store
		d         repo.DataPiece
		bou       repo.Boundary
		wantStore Store
		wantErr   error
		wantADU   repo.AppDistributorUnit
	}{
		{
			name: "B, Part is matched, E() == repo.False => creating ADU, deleting ASKD, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Close}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, D.H not full => fulfilling D_1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte"),
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("nt-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, D.H not full => fulfilling D_2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n"),
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, D.H not full => fulfilling D_3",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, part is match, E()==repo.True => incrementing detailed key part",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, part is not match => adding datapiece to buffer",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
				},
			},
			wantErr: errors.New("in store.Register dataPiece's Part for given TS \"qqq\" should be \"5\" but got \"6\""),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "B, E == Probably => detailed part remains",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, TS is not match => add to buffer",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "www", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "www"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
				},
			},
			wantErr: errors.New("in store.Register dataPiece's TS \"www\" is unknown"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "B, E == repo.False => deleting map record",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register dataPiece group with TS \"qqq\" and Part \"3\" is finished"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Close}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E == repo.Probably => detailed part unchanged",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E == repo.Probably, AppSub present => detailed part incrementing",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, while appSub present, 2 different ASKDs, 1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
					}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
					}},

				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E()==repo.Probably, while appSub present, join new and old ASKD, 1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
					}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							},
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
						},
					},
				},

				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E()==repo.Probably, 2 different askd with 1 branch in each => joining askd",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
					}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("bzbzbz")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							},
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							}},
					},
				},

				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzbz")}},
		},

		{
			name: "Registering dataPiece next to appSub. Header lines in the beginning of body",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 7}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("---------0123456789\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazazazazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------------0123456789")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 8}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register new dataPiece group with TS \"qqq\" and Part = 5"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 7}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, B: repo.BeginningData{Part: 5}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azazazazazaza")}},
		},

		{
			name: "Registering dataPiece next to appSub. No header lines in the beginning of body",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H: []byte("\r\n--------"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("--------azazaza\r\nbzbzbzbzb")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------------0123456789")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 7}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 6}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 3}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("\r\n----------------azazaza\r\nbzbzbzbzb")}},
		},

		{
			name: "Registering dataPiece next to appSub. Body contains ending of false branch header => fulfilling false brach disposition",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------------0123456789")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 7}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register new dataPiece group with TS \"qqq\" and Part = 5"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 6}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 5}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			b, err := v.store.Register(v.d, v.bou)
			if v.wantErr != nil {
				//logger.L.Errorf("in store.TestRegister err = %v\n", err)
				s.Equal(v.wantErr, err)
			}
			//logger.L.Infof("in store.TestRegister wantStore = %v\n", v.wantStore)
			//logger.L.Infof("in store.TestRegister v.store = %v\n", v.store)
			s.Equal(v.wantStore, v.store)
			//logger.L.Infof("in store.TestRegister wanB = %v\n", v.wantB)
			s.Equal(v.wantADU, b)

		})
	}

}

func (s *storeSuite) TestRegisterBuffer() {

	tt := []struct {
		name      string
		store     Store
		askg      repo.AppStoreKeyGeneral
		bou       repo.Boundary
		wantStore Store
		wantADUs  []repo.AppDistributorUnit
		wantErr   []error
	}{
		{
			name: "No elements",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 5},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
			},
			wantADUs: []repo.AppDistributorUnit{},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "Empty ASKG map",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.False,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					}},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 1}},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.False,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					}},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 1}},
			},
			wantADUs: []repo.AppDistributorUnit{},
			wantErr:  []error{},
		},

		{
			name: "len(B) = 1, part and TS matched => releasing",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					}},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 5, Started: true, Blocked: true}},
			},

			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 4, Started: true, Blocked: true}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "real case, buffer element has E() == repo.Probably, OB is registered, after releasing s.R[askg][askd] should increment its part",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{Part: 2},
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								E: repo.Probably,
							},
						},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azazaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczczcz")}},
					}},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 5, Started: true, Blocked: true}},
			},

			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 3, Started: true, Blocked: true}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("\r\nczczczcz")}},
			},
			wantErr: []error{},
		},

		{
			name: "Buffer contains dataPieces with wanted and unwanted Part, releasing one that wanted, leaving unwanted",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 3, Blocked: true}},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 2, Started: true, Blocked: true}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "len(B)>0, all E() == repo.True, TS == dataPiece's TS, parts are matched, all actions are repo.None",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 5, Started: true, Blocked: true}},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 2, Started: true, Blocked: true}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("czczc")}},
			},
			wantErr: []error{},
		},

		{
			name: "len(B)>0, all E() == repo.True, TS == dataPiece's TS, parts are matched, dataPiece with part 5 remains",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 3, Started: true, Blocked: true}},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{{TS: "qqq"}: {
					{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
						false: repo.AppStoreValue{
							B: repo.BeginningData{Part: 1},
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							E: repo.True,
						},
					}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 1, Started: true, Blocked: true},
				},
			},

			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has single element and current counter == 1 and blocked"),
			},
		},

		{
			name: "len(B)>0, element with different TS should be ignored",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
					{TS: "www"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 5, Started: true, Blocked: true}},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "www"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 3, Started: true, Blocked: true}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
			},
			wantErr: []error{},
		},

		{
			name: "len(B)>0, has E == repo.Probably",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 5, Started: true, Blocked: true}},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 6, Cur: 3, Started: true, Blocked: true}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
			},
		},

		{
			name: "len(B)==1, has current counter == 1 && blocked",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: true},
				},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: true},
				},
			},
			wantADUs: []repo.AppDistributorUnit{},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has single element and current counter == 1 and blocked"),
			},
		},

		{
			name: "len(B)==3, has current counter == 2 && blocked",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azazazaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzbzbzb")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczczczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 3, Blocked: true},
				},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczczczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: true},
				},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azazazaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzbzbzb")}},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has single element and current counter == 1 and blocked"),
			},
		},

		{
			name: "len(B)==1, has current counter == 1 && !blocked => resetting store",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azazazaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: false},
				},
			},
			askg: repo.AppStoreKeyGeneral{TS: "qqq"},
			bou:  repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, B: repo.BeginningData{Part: 1}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azazazaza")}},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" and Part \"1\" is finished"),
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {

			gotADUs, gotErr := v.store.RegisterBuffer(v.askg, v.bou)

			s.Equal(v.wantStore, v.store)
			s.Equal(v.wantADUs, gotADUs)
			s.Equal(v.wantErr, gotErr)
		})
	}
}

func (s *storeSuite) TestPresense() {

	tt := []struct {
		name         string
		store        Store
		d            repo.DataPiece
		wantPresense repo.Presense
		wantError    error
	}{
		{
			name: "No ASKG",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d:            &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{},
			wantError:    nil,
		},

		{
			name: "ASKG met, no ASKD",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true},
			wantError: nil,
		},

		{
			name: "ASKG met, wrong ASKD",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true},
			wantError: nil,
		},

		{
			name: "AppSub, no ASKG",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d:            &repo.AppSub{ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")}},
			wantPresense: repo.Presense{},
			wantError:    nil,
		},

		{
			name: "AppSub, ASKG met, no opposite detailed branch",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.True,
							},
						},
					},
				},
			},
			d: &repo.AppSub{ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")}},
			wantPresense: repo.Presense{
				ASKG: true,
			},
			wantError: nil,
		},
		{
			name: "AppSub, ASKG met, opposite detailed branch met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppSub{ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.Probably,
						},
					},
				},
			},
			wantError: nil,
		},

		{
			name: "B() == repo.False && E() == repo.Probably, no askg met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d:            &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")}},
			wantPresense: repo.Presense{},
		},

		{
			name: "B() == repo.False && E() == repo.Probably, askg met but no OB met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						}},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.False, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
			},
			wantError: nil,
		},

		{
			name: "B() == repo.False && E() == repo.Probably, askg met, OB met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.False, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
					},
				},
			},
			wantError: nil,
		},

		{
			name: "B() == repo.True && E() == repo.Probably, askd met, askd.T() met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
					},
				},
			},
			wantError: nil,
		},

		{
			name: "E() == repo.False, askd met, askd.T() met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				OB:   false,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
				},
			},
			wantError: nil,
		},

		{
			name: "E() == repo.Probably, askd met, oppoaite branch met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 1},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 1},
							E: repo.True,
						},
					},
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
					},
				},
			},
			wantError: nil,
		},

		{
			name: "ASKD met, B() == repo.True, E() == repo.False, Cur == 1 && Fuse == true => enpty Presense",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.True,
							},
						},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 1, Blocked: true},
				},
			},
			d:            &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{},
			wantError:    errors.New("in store.Presense matched but current counter == 1 && Fuse is on"),
		},

		{
			name: "ASKD met, B() == repo.True, E() == repo.False, Cur == 1 && Fuse == false => all trues",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.True,
							},
						},
					},
				},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 1, Blocked: false},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
				},
			},

			wantError: nil,
		},

		{
			name: "ASKD met, 2 specific branches, E() == repo.True",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.Probably,
							},
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								B: repo.BeginningData{Part: 3},
								E: repo.True,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 3, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				OB:   false,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.Probably,
						},
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
					},
				},
			},
			wantError: nil,
		},

		{
			name: "ASKD met, 2 detailed branches, 2 specific branches, E() == repo.Probably",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.Probably,
							},
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 3},
								E: repo.Probably,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								B: repo.BeginningData{Part: 5},
								E: repo.True,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 3, B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.Probably,
						},
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
					},
					{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{Part: 3},
							E: repo.Probably,
						},
					},
				},
			},
			wantError: nil,
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			gotPresense, gotError := v.store.Presence(v.d)
			if v.wantError != nil {
				s.Equal(v.wantError, gotError)
			}

			//logger.L.Errorf("in store.TestPresense gotError %v\n", gotError)

			s.Equal(v.wantPresense, gotPresense)
		})
	}
}
func (s *storeSuite) TestDec() {
	tt := []struct {
		name      string
		store     Store
		d         repo.DataPiece
		wantO     repo.Order
		wantStore Store
		wantError error
	}{
		{
			name: "ASKG not found",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true},
				},
			},
			d:     &repo.AppSub{ASH: repo.AppSubHeader{TS: "www", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")}},
			wantO: repo.Unordered,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true},
				},
			},
			wantError: errors.New("in store.Dec askg \"{www}\" not found"),
		},

		{
			name: "AppSub => Max --, Cur --, Started remains",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true},
				},
			},
			d:     &repo.AppSub{ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")}},
			wantO: repo.Unordered,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true},
				},
			},
		},

		{
			name: "First, Cur > 1",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Conte")},
			},
			wantO: repo.First,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true},
				},
			},
		},
		{
			name: "First, Max == 1 && Cur = 1, B() == repo.False, E() == repo.True, Blocked => decrement to 0 with no error",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 1, Cur: 1, Started: false, Blocked: true},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Conte")},
			},
			wantO: repo.First,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 1, Cur: 0, Started: true, Blocked: true},
				},
			},
		},

		{
			name: "cannot dec further_1",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 1, Cur: 1, Started: false, Blocked: true},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantO: repo.Unordered,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 1, Cur: 1, Started: false, Blocked: true},
				},
			},
			wantError: errors.New("in store.Dec cannot dec further"),
		},

		{
			name: "cannot dec further_2",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantO: repo.Unordered,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true},
				},
			},
			wantError: errors.New("in store.Dec cannot dec further"),
		},

		{
			name: "Intermediate",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("nt-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantO: repo.Intermediate,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true},
				},
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			gotO, gotError := v.store.Dec(v.d)

			if v.wantError != nil {
				s.Equal(v.wantError, gotError)
			}

			s.Equal(v.wantO, gotO)
			s.Equal(v.wantStore, v.store)

		})
	}
}

/*
	func (s *storeSuite) TestUpdate() {
		tt := []struct {
			name      string
			store     Store
			dr        repo.DetailedRecord
			wantStore Store
		}{
			{
				name: "askg, askd present, updating askd",
				store: &StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.True,
								},
							},
						},
					},
				},
				dr: repo.DetailedRecord{
					ASKD: repo.AppStoreKeyDetailed{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false},
					DR: map[bool]repo.AppStoreValue{
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
				},
				wantStore: &StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.True,
								},
							},
						},
					},
				},
			},

			{
				name: "create askg, update askd",
				store: &StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				},
				dr: repo.DetailedRecord{
					ASKD: repo.AppStoreKeyDetailed{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false},
					DR: map[bool]repo.AppStoreValue{
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
				},
				wantStore: &StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.True,
								},
							},
						},
					},
				},
			},
		}

		for _, v := range tt {
			s.Run(v.name, func() {
				v.store.Update(v.dr)

				s.Equal(v.wantStore, v.store)
			})
		}
	}
*/
func (s *storeSuite) TestAct() {
	tt := []struct {
		name      string
		store     Store
		d         repo.DataPiece
		sc        repo.StoreChange
		wantStore Store
	}{

		{
			name: "B() == repo.False & E() == repo.True, header present completely, store is empty",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.True,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
		},

		{
			name: "B() == repo.False & E() == repo.True, header present completely, store is not empty",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{
									Part: 2,
								},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.True,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{
									Part: 2,
								},
								E: repo.Probably,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
		},

		{
			name: "B() == repo.False & E() == repo.True, header present partly",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: repo.Disposition{
								H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain"),
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.True,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain"),
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
		},

		{
			name: "B() == repo.False & E() == repo.True, header present partly && ASKG met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							}}},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 3, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
						false: {
							D: repo.Disposition{
								H: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain"),
							},
							B: repo.BeginningData{
								Part: 3,
							},
							E: repo.True,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{Part: 0},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							}},

						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: {
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain"),
								},
								B: repo.BeginningData{
									Part: 3,
								},
								E: repo.True,
							},
						},
					},
				},
			},
		},

		{
			name: "AppSub, no ASKG met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{
								Part: 1,
							},
							E: repo.Probably,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{
									Part: 1,
								},
								E: repo.Probably,
							},
						},
					},
				},
			},
		},

		{
			name: "AppSub, ASKG met, no OD branch met_0",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 2}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{
								Part: 2,
							},
							E: repo.Probably,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{
									Part: 2,
								},
								E: repo.Probably,
							},
						},
					},
				},
			},
		},

		{
			name: "AppSub, ASKG met, no OD branch met_1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{
								Part: 1,
							},
							E: repo.Probably,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{
									Part: 1,
								},
								E: repo.Probably,
							},
						},
					},
				},
			},
		},

		{
			name: "AppSub, ASKG met, OD branch met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				From: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.Probably,
						},
					},
				},
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.Probably,
						},
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{
								Part: 1,
							},
							E: repo.Probably,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.Probably,
							},
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{
									Part: 1,
								},
								E: repo.Probably,
							},
						},
					},
				},
			},
		},

		{
			name: "B() == repo.False & E() == repo.Probably",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{Part: 0},
								E: repo.Probably,
							},
							true: {
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								B: repo.BeginningData{Part: 1},
								E: repo.Probably,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: {
								D: repo.Disposition{
									H: []byte("\r\nb"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("bPrefixbRoot\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				From: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.Probably,
						},
						true: {
							D: repo.Disposition{
								H: []byte("\r\n"),
							},
							B: repo.BeginningData{Part: 1},
							E: repo.Probably,
						},
					},
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: repo.Disposition{
								H: []byte("\r\nb"),
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
					},
				},
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "bob",
								FileName: "long.txt",
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
						true: {
							D: repo.Disposition{
								H: []byte("\r\nb"),
							},
							B: repo.BeginningData{Part: 2},
							E: repo.Probably,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
							true: {
								D: repo.Disposition{
									H: []byte("\r\nb"),
								},
								B: repo.BeginningData{Part: 2},
								E: repo.Probably,
							},
						},
					},
				},
			},
		},

		{
			name: "B() == repo.True & E() == repo.False, askd is not met",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			sc: repo.StoreChange{
				A: repo.Buffer,
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					},
				},
			},
		},

		{
			name: "sc.KA == Remove",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
			sc: repo.StoreChange{
				A: repo.Remove,
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
		},

		{
			name: "B() == repo.True & E() == repo.True, part matched, disposition is filled",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				From: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.True,
						},
					},
				},
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.True,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
		},

		{
			name: "B() == repo.True & E() == repo.True, part matched, disposition is not filled",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: {
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; file"),
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("name=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			sc: repo.StoreChange{
				A: repo.Change,
				From: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: repo.Disposition{
								H: []byte("Content-Disposition: form-data; name=\"alice\"; file"),
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.True,
						},
					},
				},
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{
								Part: 0,
							},
							E: repo.True,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: {
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								B: repo.BeginningData{
									Part: 0,
								},
								E: repo.True,
							},
						},
					},
				},
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			v.store.Act(v.d, v.sc)

			s.Equal(v.wantStore, v.store)
		})
	}
}
