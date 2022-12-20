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
			name: "!B, E()=repo.True, askg not met  => creting ASKG and ADU, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, Part is matched, E() == repo.False => creating ADU, deleting ASKD, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "!B => appending ASKD, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							}},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, D.H not full => fulfilling D_1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte"),
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("nt-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, D.H not full => fulfilling D_2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n"),
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, D.H not full => fulfilling D_3",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, part is match, E()==repo.True => incrementing detailed key part",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, part is not match => adding datapiece to buffer",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
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
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, TS is not match => add to buffer",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "www", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "www"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
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
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"3\" is finished"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E == repo.Last => deleting TS, issuing adu with finish message",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},
			wantErr: errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"3\" is finished"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E == repo.Probably => detailed part unchanged",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E == repo.Probably, AppSub present => detailed part incrementing",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "!B, header not full, 1 line => Dispositiion filling",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\" is not full"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B, while appSub present, 2 different ASKDs, 1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}}}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, while appSub present, 2 different ASKDs, 1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
					}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E()==repo.Probably, while appSub present, join new and old ASKD, 1",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
					}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							},
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "B, E()==repo.Probably, 2 different askd with 1 branch in each => joining askd",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Disposition: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
					}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("bzbzbz")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							},
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("bzbzbz")}},
		},

		{
			name: "!B, header not full, 2 lines => Dispositiion filling, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte"),
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte\" is not full"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B, header full, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Tyoe: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Tyoe: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "AppSub received => askd = d.Part, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{Part: 1, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("bP")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("bP"),
								},
								E: repo.Probably,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register got double-meaning dataPiece"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B, E()=repo.Probably => askd = d.Part, 2",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.Probably,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\" is not full"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B, E()=repo.Probably, askg mathced => creating askd with SK.Part == d.Part",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n\r\nazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register got double-meaning dataPiece"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "!B, E()=repo.Probably,AppSub present => askd = d.Part++",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H: []byte("\r\n-----"),
								},
								E: repo.Probably,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("-------123456\r\nContent-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},

			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------123456")},

			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("-------123456\r\nContent-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},

							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H: []byte("\r\n-----"),
								},
								E: repo.Probably,
							},
						},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "!B, askg met, header not full 1 line => New askd, D.H inconplete",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\""),
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\" is not full"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B, part is not match, full header",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 8}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 7},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 7}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "!B, part is not match, <1 header lines",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; file")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 8}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 7},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; file"),
									FormName: "",
									FileName: "",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},

			wantErr: errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"bob\"; file\" is not full"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B, part is not match, exactly 1 header line",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 8}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 7},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r"),
									FormName: "",
									FileName: "",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},

			wantErr: errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\" is not full"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B, part is not match, <2 header lines",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Ty")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 8}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 7},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Ty"),
									FormName: "",
									FileName: "",
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},

			wantErr: errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Ty\" is not full"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "!B. detailed key already present (by previously accepted appSub, which also made level 3 map true branch). Adding level 3 map false branch",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							},
						},
					}}},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: false, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							},
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},

			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
		},

		{
			name: "Registering appSub. Creating ASKG and ASKD",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{Part: 1, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("--------")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register got double-meaning dataPiece"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "Registering appSub after previous appSub. Creating ASKD",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("\r\n--------"),
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{Part: 2, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H: []byte("\r\n--------"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H: []byte("\r\n"),
								},
								E: repo.Probably,
							}},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register got double-meaning dataPiece"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "Registering appSub. Creating true detailed key",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}}}},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{Part: 5, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("--------")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: true}: {
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register got double-meaning dataPiece"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "Registering appSub white false btanch has E ==repo.Probably. Creating alternative specific key, detailed part incrementing",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							}}}},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{Part: 5, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("--------")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register got double-meaning dataPiece"),
			wantADU: repo.AppDistributorUnit{},
		},

		{
			name: "Registering dataPiece next to appSub. Header lines in the beginning of body",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 7}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("--------"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("---------0123456789\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazazazazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------------0123456789")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 8}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
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
			wantErr: errors.New("in store.Register new dataPiece group with TS \"qqq\" and BeginPart = 5"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 7}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.StopLast, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azazazazazaza")}},
		},

		{
			name: "Registering dataPiece next to appSub. No header lines in the beginning of body",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("\r\n--------"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("--------azazaza\r\nbzbzbzbzb")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------------0123456789")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 7}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 6}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}}}, B: repo.AppDistributorBody{B: []byte("\r\n----------------azazaza\r\nbzbzbzbzb")}},
		},

		{
			name: "Registering dataPiece next to appSub. Body is ending part of last boundary",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("\r\n--------"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("---------0123456789--\r\n")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------------0123456789")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: errors.New("in store.Register dataPiece group with TS \"qqq\" is finished"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 6}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}},
		},

		{
			name: "Registering dataPiece next to appSub. Body contains ending of false branch header => fulfilling false brach disposition",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.Probably,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H: []byte("\r"),
								},
								E: repo.Probably,
							},
						}}},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("---------------0123456789")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 7}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
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
			wantErr: errors.New("in store.Register new dataPiece group with TS \"qqq\" and BeginPart = 5"),
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 6}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "!B, E() == repo.False => store remains, unary adu not last",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte"),
								},
								E: repo.True,
							}}}},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: false, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte"),
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice"}, M: repo.Message{}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
		},

		{
			name: "!B, E() == repo.False => store remains, unary adu last",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte"),
								},
								E: repo.True,
							}}}},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazaza")},
			},
			bou: repo.Boundary{},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nConte"),
								},
								E: repo.True,
							}}}},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazaza")}}},
				},
			},
			wantErr: nil,
			wantADU: repo.AppDistributorUnit{},
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
		bou       repo.Boundary
		wantStore Store
		wantADUs  []repo.AppDistributorUnit
		wantErr   []error
	}{

		{
			name: "Releasing not last, leaving last element",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "counter > 1, doing nothing",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantADUs: []repo.AppDistributorUnit{},
			wantErr:  []error{},
		},

		{
			name: "counter = 2, registering from buffer",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							},
						},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"0\" is finished"),
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"2\" is finished"),
			},
		},

		{
			name: "No elements",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 5},
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
			name: "len(B)=1, B",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					}},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 5,
				},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 4,
				},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "len(B)=1, !B",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					}},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 2,
				},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 1,
				},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "len(b)=2, C > 1, doing nothing",

			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzbz")}},
					}},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 2,
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzbz")}},
					}},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 2,
				},
			},
			wantADUs: []repo.AppDistributorUnit{},
			wantErr:  []error{},
		},

		{
			name: "len(b)=2, C == 2, parts matched, releasing buffer",

			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							},
						},
					}},

				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("bzbzbz")}},
					}},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 2,
				},
			},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{
					{TS: "qqq"}: 0,
				},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("bzbzbz")}},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"0\" is finished"),
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"2\" is finished"),
			},
		},
		{
			name: "len(B)>0, all E == repo.True",
			store: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 5},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 6}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.True,
							},
						}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 5}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("czczc")}},
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
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 5},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantStore: &StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
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
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
			},
			wantADUs: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("bzbzb")}},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {

			gotADUs, gotErr := v.store.RegisterBuffer(v.bou)

			s.Equal(v.wantStore, v.store)
			s.Equal(v.wantADUs, gotADUs)
			s.Equal(v.wantErr, gotErr)
		})
	}
}
