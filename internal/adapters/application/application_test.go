package application

import (
	"errors"
	"postParser/internal/adapters/driven/store"
	"postParser/internal/repo"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type applicationSuite struct {
	suite.Suite
}

func TestApplicationSuite(t *testing.T) {
	suite.Run(t, new(applicationSuite))
}

func (s *applicationSuite) TestDecide() {
	tt := []struct {
		name         string
		d            repo.DataPiece
		wantDecision []string
	}{
		{
			name:         "!B & E() == repo.False => unary",
			d:            &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 4, TS: "www", B: false, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantDecision: []string{"unary"},
		},

		{
			name: "B() & E() == repo.False => register",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("")}},
			wantDecision: []string{
				"register", "register buffer",
			},
		},

		{
			name: "!B() & E() == repo.True => register",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("")}},
			wantDecision: []string{
				"register", "register buffer",
			},
		},

		{
			name: "B() & E() == repo.True => register",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("")}},
			wantDecision: []string{
				"register", "register buffer",
			},
		},

		{
			name: "B() & E() == repo.Probably => register",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("")}},
			wantDecision: []string{
				"register", "register buffer",
			},
		},

		{
			name: "!B() & E() == repo.Probably => register",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("")}},
			wantDecision: []string{
				"register", "register buffer",
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got := Decide(v.d)
			//logger.L.Infof("in applicatuon.TestDeside desicion = %q\n", got)
			s.Equal(v.wantDecision, got)
		})
	}
}

func (s *applicationSuite) TestExecute() {
	tt := []struct {
		name      string
		store     store.Store
		d         repo.DataPiece
		bou       repo.Boundary
		wantStote store.Store
		wantADU   []repo.AppDistributorUnit
		wantErr   []error
	}{

		{
			name: "!B & E==repo.Falce, decrementing counter, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
				L: sync.RWMutex{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "!B & E==repo.Last, C > 1, storing in buffer, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
				L: sync.RWMutex{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")}}},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{
				errors.New("in store.Register dataPiece with TS \"qqq\" and part \"3\" added to buffer"),
			},
		},

		{
			name: "!B & E==repo.Last, releasing buffer, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")}}},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"\r\n\r\nbzbzbzb")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "bob"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("bzbzbzb")}},
				{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "!B & E==repo.True, C == 3, releasing from buffer, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"\r\n\r\nczczczc")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("bzbzbzb")}}},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
				L: sync.RWMutex{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}}, B: repo.AppDistributorBody{B: []byte("bzbzbzb")}},
				{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "claire"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("czczczc")}},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"0\" is finished"),
			},
		},

		{
			name: "B & E==repo.False, C == 2, buffer contains last element => releasing from buffer, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
							}}},
				},

				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"\r\n\r\nczczczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("bzbzbzb")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}}, B: repo.AppDistributorBody{B: []byte("bzbzbzb")}},
				{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "claire"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("czczczc")}},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"0\" is finished"),
			},
		},

		{
			name: "!B & E==repo.True, C == 3, buffer contains last element => doing nothing, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"\r\n\r\nczczczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
				L: sync.RWMutex{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			wantStote: &store.StoreStruct{
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
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: false, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"\r\n\r\nczczczc")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "B & E==repo.Last, C > 1 => storing to buffer, 1",
			store: &store.StoreStruct{
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
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
				L: sync.RWMutex{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantStote: &store.StoreStruct{
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
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{
				errors.New("in store.Register dataPiece with TS \"qqq\" and Part \"1\" added to buffer"),
			},
		},

		{
			name: "B & E==repo.False, C == 1 => releasing from buffer, 1",
			store: &store.StoreStruct{
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
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							}},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.False}, APB: repo.AppPieceBody{B: []byte("bzbzbz")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "bob", FileName: "long.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Stop}}}, B: repo.AppDistributorBody{B: []byte("bzbzbz")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"2\" is finished"),
				errors.New("in store.Register dataPiece group with TS \"qqq\" and BeginPart \"0\" is finished"),
			},
		},

		{
			name: "!B & E==repo.True, buffer contains dataPiece with E()==repo.Probably => releasing from buffer, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
			},
			wantStote: &store.StoreStruct{
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
								E: repo.Probably,
							}},
					},
				},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("bzbzbz")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
			},
		},

		{
			name: "!B & E==repo.True, buffer contains dataPiece with E()==repo.Probably => releasing from buffer, 1",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
			},
			wantStote: &store.StoreStruct{
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
								E: repo.Probably,
							}},
					},
				},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("bzbzbz")}},
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
			},
		},

		{
			name: "AppSub, releasing buffer, 1",
			store: &store.StoreStruct{
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
								E: repo.Probably,
							}},
					},
				},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: true, E: repo.Last}, APB: repo.AppPieceBody{B: []byte("------\r\nazaza")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
				L: sync.RWMutex{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n----")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("--------bRoot")},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 0},
				L: sync.RWMutex{},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 2}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}, B: repo.AppDistributorBody{B: []byte("\r\n----------\r\nazaza")}},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
				errors.New("in store.Register dataPiece group with TS \"qqq\" is finished"),
			},
		},

		{
			name: "Store is empty, B => adding to buffer",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
				},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{errors.New("in store.Register dataPiece's TS \"qqq\" is unknown")},
		},

		{
			name: "Store is not empty, B & E, parts are not matched  => add to buffer",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
				},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{errors.New("in store.Register dataPiece's Part for given TS \"qqq\" should be \"2\" but got \"3\"")},
		},

		{
			name: "Store is not empty, detailed key is different, B & E, parts are not matched  => add to buffer",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
							}}},
					{TS: "www"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "www", BeginPart: 3},
							}},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
			},

			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 1},
							}}},
					{TS: "www"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "www", BeginPart: 3},
							}},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "www"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
				},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{errors.New("in store.Register dataPiece's Part for given TS \"www\" should be \"4\" but got \"5\"")},
		},

		{
			name: "Store is not empty, detailed key has 2 values, B & E, parts are not matched  => add to buffer",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
							}}},
					{TS: "www"}: {
						{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "www", BeginPart: 4},
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
							}}},
					{TS: "www"}: {
						{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "www", BeginPart: 4},
							}}},
				},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{errors.New("in store.Register dataPiece's Part for given TS \"qqq\" should be one of \"[3 5]\" but got \"6\"")},
		},

		{
			name: "Store is empty, !B & E, full header present => register with filled D",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{},
			wantStote: &store.StoreStruct{
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
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},

			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 3}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
			},
			wantErr: []error{errors.New("in store.RegisterBuffer buffer has no elements")},
		},

		{
			name: "Store is empty, !B & E, 1 full line and 1 partial line of header present => register with partial filled D, return nil and error",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type")},
			},
			bou: repo.Boundary{},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type"),
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},

			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{
				errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type\" is not full"),
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "Store is empty, !B & E, 1 partial line of header present => register with partial filled D, return nil and error",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"")},
			},
			bou: repo.Boundary{},
			wantStote: &store.StoreStruct{
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
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{
				errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\" is not full"),
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},
		{
			name: "Non-empty store, dataPiece has !B & E, 1 partial line of header present => register with partial filled D, return nil and error",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: false, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"")},
			},
			bou: repo.Boundary{},
			wantStote: &store.StoreStruct{
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
						},
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{
				errors.New("in store.Register header from dataPiece's body \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\" is not full"),
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B & E, dataPiece got remaining part of header lines  => register, fulfill Disposition",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r"),
								},
								E: repo.True,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{},
			wantStote: &store.StoreStruct{
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
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
			},
			wantErr: []error{errors.New("in store.Register new dataPiece group with TS \"qqq\" and BeginPart = 3"), errors.New("in store.RegisterBuffer buffer has no elements")},
		},

		{
			name: "AppSub. Buffer contains dataPiece with Part next to AppSub",
			store: &store.StoreStruct{
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
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("----\r\nazaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 9, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					}},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
			},
			d:   &repo.AppSub{ASH: repo.AppSubHeader{Part: 5, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("\r\n-----")}},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("-----12345")},
			wantStote: &store.StoreStruct{
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
									H: []byte("\r\n-----"),
								},
								E: repo.Probably,
							}},
					}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("----\r\nazaza")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 9, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantADU: []repo.AppDistributorUnit{},
			wantErr: []error{errors.New("in store.Register got double-meaning dataPiece")},
		},

		{
			name: "AppSub, Map contains false branch with E==repo.Probably, buffer contains dataPiece with Part next to AppSub, but with last boundary ending => clear store map and buffer, sending last stop message ",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
								E: repo.Probably,
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
					{TS: "qqq"}: {
						&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("--12345--\r\n")}},
					},
				},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
			},
			d:   &repo.AppSub{ASH: repo.AppSubHeader{Part: 0, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("\r\n-----")}},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("-----12345")},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, U: repo.UnaryData{}, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, M: repo.Message{PreAction: repo.None, PostAction: repo.Finish}}}},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
				errors.New("in store.Register dataPiece group with TS \"qqq\" is finished"),
			},
		},

		{
			name: "!B & E == repo.False, formName only => prepare ADU",
			store: &store.StoreStruct{
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: false, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
			},
			wantStote: &store.StoreStruct{
				//B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice"}}, S: repo.StreamData{}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "!B & E == repo.False, formName and fileName => prepare ADU",
			store: &store.StoreStruct{
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 3},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: false, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			wantStote: &store.StoreStruct{
				//B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.Unary, U: repo.UnaryData{TS: "qqq", F: repo.FiFo{FormName: "alice", FileName: "short.txt"}}, S: repo.StreamData{}}, B: repo.AppDistributorBody{B: []byte("azazaza")}},
			},
			wantErr: []error{},
		},

		{
			name: "Store is not empty, detailed key has 2 values, B & E",
			store: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H: []byte("\r\n----"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
							}}},
					{TS: "www"}: {
						{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "www", BeginPart: 4},
							}}}},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 2},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: true, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantStote: &store.StoreStruct{
				R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{TS: "qqq"}: {
						{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 0},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "bob",
									FileName: "long.txt",
								},
								E: repo.True,
							},
							true: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 2},
								D: repo.Disposition{
									H: []byte("\r\n----"),
								},
								E: repo.Probably,
							}},
						{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "qqq", BeginPart: 3},
								D: repo.Disposition{
									H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
									FormName: "alice",
									FileName: "short.txt",
								},
							}}},
					{TS: "www"}: {
						{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
							false: repo.AppStoreValue{
								B: repo.BeginningData{TS: "www", BeginPart: 4},
							}}},
				},
				B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				C: map[repo.AppStoreKeyGeneral]int{{TS: "qqq"}: 1},
			},
			wantADU: []repo.AppDistributorUnit{
				{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 4}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.None, PostAction: repo.None}}}, B: repo.AppDistributorBody{B: []byte("azaza")}},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := Execute(v.store, v.d, v.bou)
			if err != nil {
				//logger.L.Errorf("in application.TestExecute got error: %q\n", err)
				s.Equal(v.wantErr, err)
			}

			//logger.L.Warnf("in application.TestExecuteI got store: %q\n", v.store)
			//logger.L.Warnf("in application.TestExecuteI daraPiece became: %v\n", v.d)

			s.Equal(v.wantStote, v.store)
			s.Equal(v.wantADU, got)
		})
	}
}
