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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Close}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Close}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 7}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azazazazazaza")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 6}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("\r\n----------------azazaza\r\nbzbzbzbzb")}},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 6}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
		d         repo.DataPiece
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
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
			},
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
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
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
					{TS: "qqq"}: {Max: 6, Cur: 5}},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzbz")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
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
					{TS: "qqq"}: {Max: 6, Cur: 4}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
					{TS: "qqq"}: {Max: 6, Cur: 5}},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
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
					{TS: "qqq"}: {Max: 6, Cur: 3}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("\r\nczczczcz")}},
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
					{TS: "qqq"}: {Max: 3, Cur: 3}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
			},
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
					{TS: "qqq"}: {Max: 3, Cur: 2}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
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
					{TS: "qqq"}: {Max: 6, Cur: 5}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("dzdzdzdz")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
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
					{TS: "qqq"}: {Max: 6, Cur: 2}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("czczc")}},
			},
			wantErr: []error{},
		},

		{
			name: "len(B)>0, all E() == repo.True, TS == dataPiece's TS, parts are matched, dataPiece with part 5 is last",
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
					{TS: "qqq"}: {Max: 6, Cur: 3}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("dzdzdzdz")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]repo.Counter{},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("czczc")}},
			},
			wantErr: []error{},
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
					{TS: "qqq"}: {Max: 6, Cur: 5}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("dzdzdzdz")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
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
					{TS: "qqq"}: {Max: 6, Cur: 3}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
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
					{TS: "qqq"}: {Max: 6, Cur: 5}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 2, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("dzdzdzdz")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
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
					{TS: "qqq"}: {Max: 6, Cur: 3}},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4, N: false}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {

			gotADUs, gotErr := v.store.RegisterBuffer(v.d, v.bou)

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
	}{
		{
			name: "No ASKG",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d:            &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantPresense: repo.Presense{},
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
		},

		{
			name: "AppSub, no ASKG",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d:            &repo.AppSub{ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")}},
			wantPresense: repo.Presense{},
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
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {

			s.Equal(v.wantPresense, v.store.Presense(v.d))
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
	}{
		{
			name: "Unordered",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 4},
				},
			},
			d:     &repo.AppSub{ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")}},
			wantO: repo.Unordered,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 3, Cur: 3},
				},
			},
		},

		{
			name: "First",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 4},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Conte")},
			},
			wantO: repo.First,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 3},
				},
			},
		},

		{
			name: "First and Last",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 1, Cur: 1},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantO: repo.FirstAndLast,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{},
			},
		},

		{
			name: "Intermediate",
			store: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 3},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("nt-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantO: repo.Intermediate,
			wantStore: &StoreStruct{
				C: map[repo.AppStoreKeyGeneral]repo.Counter{
					{TS: "qqq"}: {Max: 4, Cur: 2},
				},
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			gotO := v.store.Dec(v.d)

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
