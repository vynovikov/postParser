package application

import (
	"errors"
	"sync"
	"testing"

	"github.com/vynovikov/postParser/internal/adapters/driven/store"
	"github.com/vynovikov/postParser/internal/repo"

	"github.com/stretchr/testify/suite"
)

var (
	a AppService
)

type applicationSuite struct {
	suite.Suite
}

func TestApplicationSuite(t *testing.T) {
	suite.Run(t, new(applicationSuite))
}

func (s *applicationSuite) SetupTest() {
	a = NewAppService(make(chan struct{}))
	//a.MountLogger(NewDistributorSpyLogger())

}

func (s *applicationSuite) TestHandle() {
	go func() {
		for {
			<-a.C.ChanOut
		}
	}()
	tt := []struct {
		name    string
		a       *App
		d       repo.DataPiece
		bou     repo.Boundary
		wg      sync.WaitGroup
		wantA   *App
		wantErr []error
	}{
		{
			name: "B() == repo.False, E() == repo.False, counter.Cur == counter.Max == 0, ADU preAction == repo.Start, postAction == repo.Finish",
			a: &App{
				A: a,
				S: &store.StoreStruct{
					C: map[repo.AppStoreKeyGeneral]repo.Counter{
						{TS: "qqq"}: {Started: false, Blocked: false, Max: 1, Cur: 1},
					},
				},
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.False, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
			},
			wantA: &App{
				A: a,
				S: &store.StoreStruct{
					C: map[repo.AppStoreKeyGeneral]repo.Counter{},
				},
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.Unary,
								U: repo.UnaryData{
									UK: repo.UnaryKey{
										TS:   "qqq",
										Part: 3,
									},
									F: repo.FiFo{FormName: "alice"},
									M: repo.Message{PreAction: repo.Start, PostAction: repo.Finish},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},

			wantErr: []error{},
		},
		{
			name: "B() == repo.False, E() == repo.False, counter.Cur == counter.Max decrementing counter, ADU preAction = repo.Start",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.False, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.Unary,
								U: repo.UnaryData{
									UK: repo.UnaryKey{
										TS:   "qqq",
										Part: 3,
									},
									F: repo.FiFo{FormName: "alice"},
									M: repo.Message{PreAction: repo.Start, PostAction: repo.None},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "B() == repo.False, E() == repo.False, decrementing counter, ADU preAction = repo.None",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.False, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.Unary,
								U: repo.UnaryData{
									UK: repo.UnaryKey{
										TS:   "qqq",
										Part: 3,
									},
									F: repo.FiFo{FormName: "alice"},
									M: repo.Message{PreAction: repo.None, PostAction: repo.None},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "B() == repo.False, E() == repo.True => stream-type ADU, decrementing counter, ADU preAction = repo.Start",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 0,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Start,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B() == repo.False, E() == repo.True => stream-type ADU, decrementing counter, ADU preAction = repo.Open",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 6, Cur: 4, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 1},
									E: repo.True,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 6, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Open,
										PostAction: repo.Continue,
									},
									B: repo.BeginningData{
										Part: 1,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B() == repo.False, E() == repo.True, buffer has dataPiece with matched part => 2 stream-type ADUs, decrementing counter",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzbzb")}},
						},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},

				A: a,
				L: &DistributorSpyLogger{
					calls: 2,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 0,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Start,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("bzbzbzb"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "B() == repo.False, E() == repo.Probably, no OB => askd has same part, stream-type ADU, dont check buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: repo.False, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
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
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 0,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Start,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "B() == repo.False, E() == repo.Probably, ASKG but no OB => askd has same part, stream-type ADU, decrementing counter Cur",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"lonf.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "bob",
										FileName: "long.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.False, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"lonf.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "bob",
										FileName: "long.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
							},
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 2},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 2,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Open,
										PostAction: repo.Continue,
									},
									B: repo.BeginningData{Part: 2},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "B() == repo.False, E() == repo.Probably, ASKG && OB => askd part incremented, stream-type ADU, decrementing counter Cur",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: true}: {
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: repo.False, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
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
									E: repo.Probably,
								},
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 0,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Start,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B() == repo.True, E() == repo.Probably, no ASKD => adding to buffer",
			a: &App{
				S: &store.StoreStruct{
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

					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")}},
						},
					},

					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			wantErr: []error{
				errors.New("in repo.NewStoreChange for given TS \"qqq\", Part \"2\" is unexpected"),
			},
		},

		{
			name: "B() == repo.True, E() == repo.Probably, ASKD but no OB => part remains",
			a: &App{
				S: &store.StoreStruct{
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

					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
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
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "B() == repo.True, E() == repo.Probably, ASKD && OB => part increments",
			a: &App{
				S: &store.StoreStruct{
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
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 1},
									E: repo.Probably,
								},
							},
						},
					},

					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 1},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},
		{
			name: "B() == repo.True, E() == repo.True, no header => old false branch continued",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 1},
									E: repo.Probably,
								},
							},
						},
					},

					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azazazazazazazazazazazazazazazazazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},

					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 2,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("\r\nazazazazazazazazazazazazazazazazazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in repo.GetHeaderLines no header found"),
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B() == repo.True, E() == repo.True, no header => new false branch started",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 1},
									E: repo.Probably,
								},
							},
						},
					},

					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bPrefixbRoot\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbzbzb")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
								false: repo.AppStoreValue{
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "bob",
										FileName: "long.txt",
									},
									B: repo.BeginningData{Part: 1},
									E: repo.True,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},

					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 2,
									},
									F: repo.FiFo{
										FormName: "bob",
										FileName: "long.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
									B: repo.BeginningData{Part: 1},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("bzbzbzbzb"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},
		{
			name: "AppSub, no ASKG  => Adding ASKG and ASKD.T(), decrementing counter Max and Cur",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 0}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: true}: {
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 0},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			wantErr: []error{},
		},
		{
			name: "AppSub, ASKG, no OB => Adding ASKD.T(), decrementing counter Max and Cur",
			a: &App{
				S: &store.StoreStruct{
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
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
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 1},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			wantErr: []error{},
		},

		{
			name: "AppSub, OB => Combining opposite branches, decrementing counter Max and Cur",
			a: &App{
				S: &store.StoreStruct{
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 4, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
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
								true: repo.AppStoreValue{
									D: repo.Disposition{
										H: []byte("\r\n"),
									},
									B: repo.BeginningData{Part: 1},
									E: repo.Probably,
								},
							},
						},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 3, Started: false, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B() == repo.True, E() == repo.True, Part matched => stream-type ADU, decrementing counter",
			a: &App{
				S: &store.StoreStruct{
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},

			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azazaza")},
			},
			bou: repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B() == repo.False & E() == repo.True, C == 2, releasing from buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("bzbzbzb")}}},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: false}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 2,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 0,
									},
									F: repo.FiFo{
										FormName: "bob",
										FileName: "long.txt",
									},
									M: repo.Message{
										PreAction:  repo.Open,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "bob",
										FileName: "long.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Finish,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("bzbzbzb"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" and Part \"0\" is finished"),
			},
		},

		{
			name: "B() == repo.True & E() == repo.False, C > 1 => ADU poatAction == repo.Stop",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									E: repo.True,
								}}}},

					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									E: repo.False,
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Close,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},
		{
			name: "B() == repo.False & E() == repo.True, buffer contains dataPiece with E() == repo.Probably => releasing from buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
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
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 2,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 0,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Open,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("bzbzbz"),
							},
						},
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
			},
		},

		{
			name: "B() == repo.False & E() == repo.True, buffer contains dataPiece with E()==repo.Probably => releasing from buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
						},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 0, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbz")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
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
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 2,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 0,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Open,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("bzbzbz"),
							},
						},
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.Register got double-meaning dataPiece"),
			},
		},

		{
			name: "AppSub, releasing buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
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
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 2, TS: "qqq", B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("------\r\nazaza")}},
						},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: false}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n----")},
			},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("--------bRoot")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{},
				},

				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 2,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Finish,
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("\r\n----------\r\nazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "Store is empty, B => adding to buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
					},
				},
			},
			wantErr: []error{
				errors.New("in repo.NewStoreChange TS \"qqq\" is unknown"),
			},
		},

		{
			name: "Store is not empty, B & E, parts are not matched  => add to buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 1},
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 1},
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
					},
				},
			},
			wantErr: []error{
				errors.New("in repo.NewStoreChange for given TS \"qqq\", Part \"3\" is unexpected"),
			},
		},

		{
			name: "Store is not empty, detailed key is different, B & E, parts are not matched  => add to buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 1},
								}}},
						{TS: "www"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
								}},
						}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
				},
			},
			d: &repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 1},
								}}},
						{TS: "www"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
								}},
						}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "www"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 5, TS: "www", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
					},
				},
			},
			wantErr: []error{
				errors.New("in repo.NewStoreChange for given TS \"www\", Part \"5\" is unexpected"),
			},
		},

		{
			name: "Detailed key has 2 values, B() == repo.True & E() == repo.True, parts are not matched  => add to buffer",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
								},
								true: repo.AppStoreValue{
									B: repo.BeginningData{Part: 2},
								}},
							{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
								}}},
						{TS: "www"}: {
							{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 4},
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
								},
								true: repo.AppStoreValue{
									B: repo.BeginningData{Part: 2},
								}},
							{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
								}}},
						{TS: "www"}: {
							{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 4},
								}}},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}}},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
			},
			wantErr: []error{
				errors.New("in repo.NewStoreChange for given TS \"qqq\", Part \"6\" is unexpected"),
			},
		},

		{
			name: "B() == repo.False & E() == repo.True, full header present => register with filled D",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{},
			wantA: &App{
				S: &store.StoreStruct{
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 3,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Open,
										PostAction: repo.Continue,
									},
									B: repo.BeginningData{Part: 3},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{errors.New("in store.RegisterBuffer buffer has no elements")},
		},

		{
			name: "B() == repo.False & E() == repo.True, 1 full line and 1 partial line of header present => register with partial filled D, return nil and error",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type")},
			},
			bou: repo.Boundary{},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
									D: repo.Disposition{
										H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type"),
									},
									E: repo.True,
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
			},
			wantErr: []error{
				errors.New("in repo.GetHeaderLines header \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type\" is not full"),
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "Empty store, B() == repo.False & E() == repo.True, 1 partial line of header present => register with partial filled D, return nil and error",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"")},
			},
			bou: repo.Boundary{},
			wantA: &App{
				S: &store.StoreStruct{
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
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
			},
			wantErr: []error{
				errors.New("in repo.GetHeaderLines header \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\" is not full"),
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "Store is not empty, B() == repo.False & E() == repo.True, 1 partial line of header present => register with partial filled D, return nil and error",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
									D: repo.Disposition{
										H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
									},
									E: repo.True,
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 3, TS: "qqq", B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"")},
			},
			bou: repo.Boundary{},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
									D: repo.Disposition{
										H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\""),
									},
									E: repo.True,
								},
							},
							{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
									D: repo.Disposition{
										H: []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\""),
									},
									E: repo.True,
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
			},
			wantErr: []error{
				errors.New("in repo.GetHeaderLines header \"Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\" is not full"),
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "B() == repo.True & E() == repo.True, dataPiece got remaining part of header lines  => register, fulfill Disposition",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
									D: repo.Disposition{
										H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r"),
									},
									E: repo.True,
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			bou: repo.Boundary{},
			wantA: &App{
				S: &store.StoreStruct{
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
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 4,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
									B: repo.BeginningData{Part: 3},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},

		{
			name: "AppSub. Buffer contains dataPiece with Part next to AppSub",
			a: &App{
				S: &store.StoreStruct{
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
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("----\r\nazaza")}},
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 9, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
						}},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
			},
			d:   &repo.AppSub{ASH: repo.AppSubHeader{Part: 5, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("\r\n-----")}},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("-----12345")},
			wantA: &App{
				S: &store.StoreStruct{
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
								}},
							{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: true}: {
								true: repo.AppStoreValue{
									B: repo.BeginningData{Part: 5},
									D: repo.Disposition{
										H: []byte("\r\n-----"),
									},
									E: repo.Probably,
								}},
						}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 6, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("----\r\nazaza")}},
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 9, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("czczc")}},
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 7, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("bzbzb")}},
						},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 2, Started: true, Blocked: true}},
				},
			},
			wantErr: []error{},
		},

		{
			name: "AppSub, Map contains false branch with E==repo.Probably, buffer contains dataPiece with Part next to AppSub, but with last boundary ending => clear store map and buffer, sending last stop message ",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 0}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									E: repo.Probably,
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{
						{TS: "qqq"}: {
							&repo.AppPieceUnit{APH: repo.AppPieceHeader{Part: 1, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("--12345--\r\n")}},
						},
					},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d:   &repo.AppSub{ASH: repo.AppSubHeader{Part: 0, TS: "qqq"}, ASB: repo.AppSubBody{B: []byte("\r\n-----")}},
			bou: repo.Boundary{Prefix: []byte("--"), Root: []byte("-----12345")},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 3, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 1,
									},
									F: repo.FiFo{
										FormName: "",
										FileName: "",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Finish,
									},
								},
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.Register dataPiece group with TS \"qqq\" is finished"),
			},
		},

		{
			name: "B() == repo.False & E() == repo.False, formName only => prepare ADU",
			a: &App{
				S: &store.StoreStruct{
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.False, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.Unary,
								U: repo.UnaryData{
									UK: repo.UnaryKey{
										TS:   "qqq",
										Part: 4,
									},
									F: repo.FiFo{
										FormName: "alice",
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "B() == repo.False & E() == repo.False, formName and fileName => prepare ADU",
			a: &App{
				S: &store.StoreStruct{
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 3, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.False, E: repo.False}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.Unary,
								U: repo.UnaryData{
									UK: repo.UnaryKey{
										TS:   "qqq",
										Part: 4,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azazaza"),
							},
						},
					},
				},
			},
			wantErr: []error{},
		},

		{
			name: "Store is not empty, detailed key has 2 values, B() == repo.True, E() == repo.True",
			a: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "bob",
										FileName: "long.txt",
									},
									E: repo.True,
								},
								true: repo.AppStoreValue{
									B: repo.BeginningData{Part: 2},
									D: repo.Disposition{
										H: []byte("\r\n----"),
									},
									E: repo.Probably,
								}},
							{SK: repo.StreamKey{TS: "qqq", Part: 4}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									E: repo.True,
								}}},
						{TS: "www"}: {
							{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 4},
								}}}},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 2, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{},
			},
			d: &repo.AppPieceUnit{
				APH: repo.AppPieceHeader{Part: 4, TS: "qqq", B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")},
			},
			wantA: &App{
				S: &store.StoreStruct{
					R: map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
						{TS: "qqq"}: {
							{SK: repo.StreamKey{TS: "qqq", Part: 3}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 0},
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "bob",
										FileName: "long.txt",
									},
									E: repo.True,
								},
								true: repo.AppStoreValue{
									B: repo.BeginningData{Part: 2},
									D: repo.Disposition{
										H: []byte("\r\n----"),
									},
									E: repo.Probably,
								}},
							{SK: repo.StreamKey{TS: "qqq", Part: 5}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 3},
									D: repo.Disposition{
										H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
										FormName: "alice",
										FileName: "short.txt",
									},
									E: repo.True,
								}}},
						{TS: "www"}: {
							{SK: repo.StreamKey{TS: "www", Part: 5}, S: false}: {
								false: repo.AppStoreValue{
									B: repo.BeginningData{Part: 4},
								}}},
					},
					B: map[repo.AppStoreKeyGeneral][]repo.DataPiece{},
					C: map[repo.AppStoreKeyGeneral]repo.Counter{{TS: "qqq"}: {Max: 4, Cur: 1, Started: true, Blocked: true}},
				},
				A: a,
				L: &DistributorSpyLogger{
					calls: 1,
					params: []repo.AppUnit{
						repo.AppDistributorUnit{
							H: repo.AppDistributorHeader{
								T: repo.ClientStream,
								S: repo.StreamData{
									SK: repo.StreamKey{
										TS:   "qqq",
										Part: 4,
									},
									F: repo.FiFo{
										FormName: "alice",
										FileName: "short.txt",
									},
									M: repo.Message{
										PreAction:  repo.Continue,
										PostAction: repo.Continue,
									},
									B: repo.BeginningData{Part: 3},
								},
							},
							B: repo.AppDistributorBody{
								B: []byte("azaza"),
							},
						},
					},
				},
			},
			wantErr: []error{
				errors.New("in store.RegisterBuffer buffer has no elements"),
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			v.wg.Add(1)
			go v.a.Handle(v.d, v.bou, &v.wg, 0)
			//logger.L.Infoln("in application.TestHandle waiting...")
			v.wg.Wait()

			s.Equal(v.wantA, v.a)
		})
	}
}
func (s *applicationSuite) TestCalcBody() {
	tt := []struct {
		name string

		d          repo.DataPiece
		bou        repo.Boundary
		wantADUB   repo.AppDistributorBody
		wantHeader []byte
		wantError  error
	}{
		{
			name:     "no header present",
			d:        &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			bou:      repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantADUB: repo.AppDistributorBody{B: []byte("azaza")},
		},

		{
			name:       "header present",
			d:          &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("bPrefixbRoot\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nbzbzbzbzb")}},
			bou:        repo.Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantADUB:   repo.AppDistributorBody{B: []byte("bzbzbzbzb")},
			wantHeader: []byte("bPrefixbRoot\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotADUB, gotHeader, gotErr := CalcBody(v.d, v.bou)
			if v.wantError != nil {
				s.Equal(v.wantError, gotErr)
			}
			s.Equal(v.wantADUB, gotADUB)
			s.Equal(v.wantHeader, gotHeader)

		})
	}
}

func (s *applicationSuite) TestCalcHeader() {
	tt := []struct {
		name     string
		d        repo.DataPiece
		sc       repo.StoreChange
		o        repo.Order
		wantADUH repo.AppDistributorHeader
	}{
		{
			name: "AppSub",
			d: &repo.AppSub{
				ASH: repo.AppSubHeader{TS: "qqq", Part: 1}, ASB: repo.AppSubBody{B: []byte("\r\n")},
			},
			wantADUH: repo.AppDistributorHeader{},
		},

		{
			name: "o == repo.First B() == repo.False, E() == repo.True",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 0, B: repo.False, E: repo.True}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")}},
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
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
				},
			},
			o:        repo.First,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 0}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Start, PostAction: repo.Continue}}},
		},

		{
			name: "o == repo.First B() == repo.False, E() == repo.Probably",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.False, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")}},
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
							B: repo.BeginningData{Part: 0},
							E: repo.Probably,
						},
					},
				},
			},
			o:        repo.First,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Start, PostAction: repo.Continue}}},
		},

		{
			name: "o == repo.Intermediate, B() == repo.True, E() == repo.True",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.True}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
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
			o:        repo.Intermediate,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}},
		},

		{
			name: "o == repo.Intermediate, B() == repo.True, E() == repo.False",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.False,
						},
					},
				},
			},
			o:        repo.Intermediate,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Close}}},
		},

		{
			name: "o == repo.Intermediate, B() == repo.True, E() == repo.Probably",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
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
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
				},
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
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
			o:        repo.Intermediate,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}},
		},

		{
			name: "o == repo.Intermediate, B() == repo.True, E() == repo.Probably, len(sc.From) == 2",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.Probably}, APB: repo.AppPieceBody{B: []byte("azaza")}},
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
							B: repo.BeginningData{Part: 0},
							E: repo.True,
						},
					},
					{SK: repo.StreamKey{TS: "qqq", Part: 1}, S: true}: {
						true: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 1},
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
							B: repo.BeginningData{Part: 0},
							E: repo.Probably,
						},
						true: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 1},
							E: repo.Probably,
						},
					},
				},
			},
			o:        repo.Intermediate,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}}},
		},

		{
			name: "o == repo.Intermediate, B() == repo.True, E() == repo.False",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.False,
						},
					},
				},
			},
			o:        repo.Intermediate,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Close}}},
		},

		{
			name: "o == repo.Last, B() == repo.True, E() == repo.False",
			d:    &repo.AppPieceUnit{APH: repo.AppPieceHeader{TS: "qqq", Part: 1, B: repo.True, E: repo.False}, APB: repo.AppPieceBody{B: []byte("azaza")}},
			sc: repo.StoreChange{
				A: repo.Change,
				To: map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{
					{SK: repo.StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: repo.Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: repo.BeginningData{Part: 0},
							E: repo.False,
						},
					},
				},
			},
			o:        repo.Last,
			wantADUH: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: "qqq", Part: 1}, F: repo.FiFo{FormName: "alice", FileName: "short.txt"}, M: repo.Message{PreAction: repo.Continue, PostAction: repo.Finish}}},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {

			s.Equal(v.wantADUH, CalcHeader(v.d, v.sc, v.o))

		})
	}
}

type DistributorSpyLogger struct {
	calls  int
	params []repo.AppUnit
}

func (d *DistributorSpyLogger) LogStuff(au repo.AppUnit) {
	d.calls++
	d.params = append(d.params, au)
}

func NewDistributorSpyLogger() *DistributorSpyLogger {
	return &DistributorSpyLogger{}
}
