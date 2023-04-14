package repo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type modelsSuite struct {
	suite.Suite
}

func TestModelsSuite(t *testing.T) {
	suite.Run(t, new(modelsSuite))
}

func (s *modelsSuite) TestH() {
	tt := []struct {
		name        string
		d           DataPiece
		bou         Boundary
		wantedH     []byte
		wantedError error
	}{
		{
			name: "<1 header line",
			d: &AppPieceUnit{
				APB: AppPieceBody{B: []byte("Content-Dispo")},
			},
			wantedH:     []byte("Content-Dispo"),
			wantedError: errors.New("in repo.GetHeaderLines header \"Content-Dispo\" is not full"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotH, err := v.d.H(v.bou)
			if v.wantedError != nil {
				s.Equal(v.wantedError, err)
			}
			s.Equal(v.wantedH, gotH)
		})
	}
}

func (s *modelsSuite) TestBoduCut() {
	tt := []struct {
		name    string
		d       DataPiece
		cut     int
		wantedD DataPiece
	}{
		{
			name:    "cut < bodyLen",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: False, E: True}, APB: AppPieceBody{B: []byte("0123456789")}},
			cut:     3,
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: False, E: True}, APB: AppPieceBody{B: []byte("3456789")}},
		},

		{
			name:    "cut == bodyLen",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: False, E: True}, APB: AppPieceBody{B: []byte("0123456789")}},
			cut:     10,
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: False, E: True}, APB: AppPieceBody{B: []byte("")}},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			v.d.BodyCut(v.cut)
			s.Equal(v.wantedD, v.d)
		})
	}

}

func (s *modelsSuite) TestAppStoreBufferIDsAdd() {
	tt := []struct {
		name       string
		asbi       AppStoreBufferIDs
		adding     AppStoreBufferID
		wantedASBI AppStoreBufferIDs
	}{

		{
			name:       "no TS",
			asbi:       AppStoreBufferIDs{},
			adding:     AppStoreBufferID{ASKG: AppStoreKeyGeneral{TS: "qqq"}, I: 1},
			wantedASBI: AppStoreBufferIDs{ASKG: AppStoreKeyGeneral{TS: "qqq"}, I: []int{1}},
		},

		{
			name:       "TS matched",
			asbi:       AppStoreBufferIDs{ASKG: AppStoreKeyGeneral{TS: "qqq"}, I: []int{0}},
			adding:     AppStoreBufferID{ASKG: AppStoreKeyGeneral{TS: "qqq"}, I: 1},
			wantedASBI: AppStoreBufferIDs{ASKG: AppStoreKeyGeneral{TS: "qqq"}, I: []int{0, 1}},
		},

		{
			name:       "TS unmatched",
			asbi:       AppStoreBufferIDs{ASKG: AppStoreKeyGeneral{TS: "qqq"}, I: []int{0}},
			adding:     AppStoreBufferID{ASKG: AppStoreKeyGeneral{TS: "www"}, I: 1},
			wantedASBI: AppStoreBufferIDs{ASKG: AppStoreKeyGeneral{TS: "qqq"}, I: []int{0}},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			v.asbi.Add(v.adding)
			s.Equal(v.wantedASBI, v.asbi)
		})
	}
}

func (s *modelsSuite) TestNewStoreChange() {
	tt := []struct {
		name    string
		p       Presense
		bou     Boundary
		d       DataPiece
		wantSC  StoreChange
		wantErr error
	}{
		{
			name: "Error TS is unknown",
			p:    Presense{},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 0, B: True, E: True}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Buffer,
			},
			wantErr: errors.New("in repo.NewStoreChange TS \"qqq\" is unknown"),
		},

		{
			name: "Error Part is unexpected",
			p:    Presense{ASKG: true},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 1, B: True, E: True}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Buffer,
			},
			wantErr: errors.New("in repo.NewStoreChange for given TS \"qqq\", Part \"1\" is unexpected"),
		},

		{
			name: "B() == False, header full, d.Part() == 0",
			p:    Presense{},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 0, B: False, E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "B() == False, header full, d.Part() != 0",
			p:    Presense{},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 1, B: False, E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 1},
							E: True,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "B() == False, header not full",
			p:    Presense{},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 0, B: False, E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: te")},
			},
			wantSC: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: te"),
								FormName: "",
								FileName: "",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			wantErr: errors.New("in repo.GetHeaderLines header \"Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: te\" is not full"),
		},

		{
			name: "B() == False, E() == Probably, no OP",
			p:    Presense{},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 0, B: False, E: Probably}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 0}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "B() == False, E() == Probably, OP met",
			p: Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: False, E: Probably}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "AppSub, ASKG",
			p: Presense{
				ASKG: true,
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppSub{
				ASH: AppSubHeader{Part: 3, TS: "qqq"}, ASB: AppSubBody{B: []byte("\r\n")},
			},
			wantSC: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: true}: {
						true: {
							D: Disposition{
								H:        []byte("\r\n"),
								FormName: "",
								FileName: "",
							},
							B: BeginningData{Part: 3},
							E: Probably,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "AppSub, no ASKG",
			p:    Presense{},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppSub{
				ASH: AppSubHeader{Part: 1, TS: "qqq"}, ASB: AppSubBody{B: []byte("\r\n")},
			},
			wantSC: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: true}: {
						true: {
							D: Disposition{
								H:        []byte("\r\n"),
								FormName: "",
								FileName: "",
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "AppSub, ASKD, no opposite detailed branch met",
			p: Presense{
				ASKG: true,
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppSub{
				ASH: AppSubHeader{Part: 1, TS: "qqq"}, ASB: AppSubBody{B: []byte("\r\n")},
			},
			wantSC: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: true}: {
						true: {
							D: Disposition{
								H:        []byte("\r\n"),
								FormName: "",
								FileName: "",
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "AppSub, ASKD, opposite detailed branch met",
			p: Presense{
				ASKG: true,
				ASKD: false,
				OB:   true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n "),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppSub{
				ASH: AppSubHeader{Part: 1, TS: "qqq"}, ASB: AppSubBody{B: []byte("\r\n")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n "),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n "),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
						true: {
							D: Disposition{
								H:        []byte("\r\n"),
								FormName: "",
								FileName: "",
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "B() == True, no askd met",
			p:    Presense{},
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 1, B: True, E: False}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Buffer,
			},
			wantErr: errors.New("in repo.NewStoreChange TS \"qqq\" is unknown"),
		},

		{
			name: "B() == True, header not full",
			p: Presense{
				ASKG: true,
				ASKD: true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H: []byte("Content-Disposition: form-data; n"),
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 1, B: True, E: True}, APB: AppPieceBody{B: []byte("ame=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H: []byte("Content-Disposition: form-data; n"),
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
		},

		{
			name: "B() == True, E() == False",
			p: Presense{
				ASKG: true,
				ASKD: true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: False}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: False,
						},
					},
				},
			},
		},

		{
			name: "B() == True, E() == Probably no opposite branch met",
			p: Presense{
				ASKG: true,
				ASKD: true,
				OB:   false,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: Probably}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
			},
		},

		{
			name: "B() == True, E() == Probably opposite branch met",
			p: Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: Probably}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
		},

		{
			name: "After forked askd, B() == True, old body continued",
			p: Presense{
				ASKG: true,
				ASKD: true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: True}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
		},

		{
			name: "After forked askd, B() == True, new header started",
			p: Presense{
				ASKG: true,
				ASKD: true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: True}, APB: AppPieceBody{B: []byte("bPrefixbRoot\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "bob",
								FileName: "long.txt",
							},
							B: BeginningData{Part: 1},
							E: True,
						},
					},
				},
			},
		},

		{
			name: "Forked askd, true ASKD present, B() == Probably, old header continued",
			p: Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\nb"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: Probably}, APB: AppPieceBody{B: []byte("azaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\nb"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\nb"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
		},

		{
			name: "Forked askd, true ASKD present, B() == Probably, new header started",
			p: Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\nb"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: Probably}, APB: AppPieceBody{B: []byte("bPrefixbRoot\r\nContent-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
					{SK: StreamKey{TS: "qqq", Part: 2}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\nb"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"bob\"; filename=\"long.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "bob",
								FileName: "long.txt",
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\nb"),
							},
							B: BeginningData{Part: 2},
							E: Probably,
						},
					},
				},
			},
		},

		{
			name: "Forked askd, OB present, appSub is part of header",
			p: Presense{
				ASKG: true,
				ASKD: true,
				OB:   true,
				GR: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r"),
							},
							B: BeginningData{Part: 1},
						},
						true: {
							D: Disposition{
								H: []byte("\r"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: Probably}, APB: AppPieceBody{B: []byte("\nContent-Type: text/plain\r\n\r\nazaza")},
			},
			wantSC: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 2}, S: false}: {
						false: {
							D: Disposition{
								H: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r"),
							},
							B: BeginningData{Part: 1},
						},
						true: {
							D: Disposition{
								H: []byte("\r"),
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 3}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 1},
							E: Probably,
						},
					},
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotSC, gotErr := NewStoreChange(v.d, v.p, v.bou)
			if v.wantErr != nil {
				s.Equal(v.wantErr, gotErr)
			}
			s.Equal(v.wantSC, gotSC)
		})
	}
}

func (s *modelsSuite) TestNewAppStoreValue() {
	tt := []struct {
		name    string
		d       DataPiece
		bou     Boundary
		wantASV AppStoreValue
		wantErr error
	}{
		{
			name: "from dataPiece, header is not full",
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 1, B: False, E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; fil")},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantASV: AppStoreValue{
				D: Disposition{
					H: []byte("Content-Disposition: form-data; name=\"alice\"; fil"),
				},
				B: BeginningData{Part: 1},
				E: True,
			},
			wantErr: errors.New("in repo.GetHeaderLines header \"Content-Disposition: form-data; name=\"alice\"; fil\" is not full"),
		},

		{
			name: "from dataPiece, header is full",
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 1, B: False, E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n\r\nazaza")},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantASV: AppStoreValue{
				D: Disposition{
					H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n\r\n"),
					FormName: "alice",
					FileName: "short.txt",
				},
				B: BeginningData{Part: 1},
				E: True,
			},
			wantErr: nil,
		},

		{
			name: "from appSub",
			d: &AppSub{
				ASH: AppSubHeader{TS: "qqq", Part: 1}, ASB: AppSubBody{B: []byte("\r\n")},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantASV: AppStoreValue{
				D: Disposition{
					H: []byte("\r\n"),
				},
				B: BeginningData{Part: 1},
				E: Probably,
			},
			wantErr: nil,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotASV, gotErr := NewAppStoreValue(v.d, v.bou)
			if v.wantErr != nil {
				s.Equal(v.wantErr, gotErr)
			}
			s.Equal(v.wantASV, gotASV)
		})
	}
}

func (s *modelsSuite) TestCompleteAppStoreValue() {
	tt := []struct {
		name    string
		asv     AppStoreValue
		d       DataPiece
		bou     Boundary
		wantASV AppStoreValue
		wantErr error
	}{
		{
			name: "Made from appPieceUnit with incomplete header",
			asv: AppStoreValue{
				D: Disposition{
					H: []byte("Content-Disposition: form-data; name=\"alice\"; fil"),
				},
				B: BeginningData{Part: 0},
				E: True,
			},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: True}, APB: AppPieceBody{B: []byte("ename=\"short.txt\"\r\n\r\nazaza")},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantASV: AppStoreValue{
				D: Disposition{
					H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n\r\n"),
					FormName: "alice",
					FileName: "short.txt",
				},
				B: BeginningData{Part: 0},
				E: True,
			},
			wantErr: nil,
		},

		{
			name: "Made from appSub, dataPiece has header lines",
			asv: AppStoreValue{
				D: Disposition{
					H: []byte("\r\n"),
				},
				B: BeginningData{Part: 0},
				E: True,
			},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: True}, APB: AppPieceBody{B: []byte("bPrefixbRoot\r\nContent-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n\r\nazaza")},
			},
			bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantASV: AppStoreValue{
				D: Disposition{
					H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\n\r\n"),
					FormName: "alice",
					FileName: "short.txt",
				},
				B: BeginningData{Part: 0},
				E: True,
			},
			wantErr: nil,
		},

		{
			name: "Made from appSub, dataPiece has no header lines",
			asv: AppStoreValue{
				D: Disposition{
					H: []byte("\r\n"),
				},
				B: BeginningData{Part: 0},
				E: True,
			},
			d: &AppPieceUnit{
				APH: AppPieceHeader{TS: "qqq", Part: 2, B: True, E: True}, APB: AppPieceBody{B: []byte("azaza")},
			},
			bou:     Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			wantASV: AppStoreValue{},
			wantErr: errors.New("in repo.GetHeaderLines no header found"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotASV, gotErr := CompleteAppStoreValue(v.asv, v.d, v.bou)
			if v.wantErr != nil {
				s.Equal(v.wantErr, gotErr)
			}
			s.Equal(v.wantASV, gotASV)
		})
	}
}
func (s *modelsSuite) TestIsPartChanged() {
	tt := []struct {
		name string
		sc   StoreChange
		want bool
	}{
		{
			name: "sc.A ==Buffer",
			sc: StoreChange{
				A: Buffer,
			},
			want: false,
		},

		{
			name: "sc.A ==Change && len(sc.From)==0, E() ==True",
			sc: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			want: true,
		},

		{
			name: "sc.A ==Change && len(sc.From)==0, E() ==Probably",
			sc: StoreChange{
				A: Change,
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 0}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
			},
			want: false,
		},

		{
			name: "sc.A ==Change && len(sc.From)==1, len(sc.To)==1",
			sc: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 0}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
			},
			want: true,
		},

		{
			name: "sc.A ==Change && len(sc.From)==1, len(sc.To)==1, APU",
			sc: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 0}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {D: Disposition{
							H: []byte("\r\n"),
						},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
			},
			want: true,
		},

		{
			name: "sc.A ==Change && len(sc.From)==1, len(sc.To)==1, AppSub",
			sc: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 0}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
			},
			want: true,
		},

		{
			name: "sc.A ==Change && len(sc.From)==2, len(sc.To)==1",
			sc: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 0}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
					{SK: StreamKey{TS: "qqq", Part: 0}, S: true}: {
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
						true: {
							D: Disposition{
								H: []byte("\r\n"),
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
			},
			want: true,
		},

		{
			name: "sc.A ==Change && len(sc.From)==1, len(sc.To)==1, APU part remains",
			sc: StoreChange{
				A: Change,
				From: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: true}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: True,
						},
					},
				},
				To: map[AppStoreKeyDetailed]map[bool]AppStoreValue{
					{SK: StreamKey{TS: "qqq", Part: 1}, S: false}: {
						false: {
							D: Disposition{
								H:        []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\n"),
								FormName: "alice",
								FileName: "short.txt",
							},
							B: BeginningData{Part: 0},
							E: Probably,
						},
					},
				},
			},
			want: false,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {

			s.Equal(v.want, IsPartChanged(v.sc))

		})
	}
}

func (s *modelsSuite) TestIsOk() {
	tt := []struct {
		name   string
		want   []GRequest
		got    []GRequest
		wantOK bool
	}{

		{
			name: "len(y) == 0",
			want: []GRequest{{
				FieldName: "alice",
				ByteChunk: []byte("azaza"),
				RType:     U,
			}},
			got:    []GRequest{},
			wantOK: false,
		},

		{
			name: "len(x) != len(y)",
			want: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					RType:     U,
				},
				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzbzbzb"),
					RType:     U,
				},
			},
			got: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					RType:     U,
				},
			},
			wantOK: false,
		},

		{
			name: "unary all equal",
			want: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					RType:     U,
				},
				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzbzbzb"),
					RType:     U,
				},
			},
			got: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					IsFirst:   true,
					RType:     U,
				},

				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzbzbzb"),
					IsLast:    true,
					RType:     U,
				},
			},
			wantOK: true,
		},

		{
			name: "unary not equal_1",
			want: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					RType:     U,
				},
				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzbzb"),
					RType:     U,
				},
			},
			got: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					IsFirst:   true,
					RType:     U,
				},

				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzbzbzb"),
					IsLast:    true,
					RType:     U,
				},
			},
			wantOK: false,
		},

		{
			name: "unary not equal_2",
			want: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					RType:     U,
				},
				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzbzb"),
					RType:     U,
				},
			},
			got: []GRequest{
				{
					FieldName: "alice",
					ByteChunk: []byte("azaza"),
					IsFirst:   true,
					RType:     U,
				},

				{
					FieldName: "claire",
					ByteChunk: []byte("bzbzbzb"),
					IsLast:    true,
					RType:     U,
				},
			},
			wantOK: false,
		},

		{
			name: "stream all equal, ordered",
			want: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					RType:     S,
				},
			},
			got: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					IsFirst:   true,
					RType:     S,
				},

				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					IsLast:    true,
					RType:     S,
				},
			},
			wantOK: true,
		},
		{
			name: "stream all equal, unordered in the valid way 1",
			want: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					RType:     S,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
				},
			},
			got: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					IsFirst:   true,
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
					IsLast:    true,
				},
			},
			wantOK: true,
		},
		{
			name: "stream all equal, unordered in the valid way 2",
			want: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					RType:     S,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
				},
			},
			got: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					IsFirst:   true,
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
				},

				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					IsLast:    true,
					RType:     S,
				},
			},
			wantOK: true,
		},
		{
			name: "stream all equal, unordered in the unvalid way",
			want: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					RType:     S,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("333333333333"),
					RType:     S,
				},
			},
			got: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					IsFirst:   true,
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					IsLast:    true,
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("333333333333"),
					RType:     S,
					IsLast:    true,
				},
			},
			wantOK: false,
		},

		{
			name: "unaries and unorderd streams",
			want: []GRequest{
				{
					FieldName: "claire",
					ByteChunk: []byte("czczcz"),
					RType:     U,
				},
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					RType:     S,
				},
				{
					FieldName: "david",
					ByteChunk: []byte("dzdzdz"),
					RType:     U,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
				},
			},
			got: []GRequest{
				{
					FieldName: "claire",
					ByteChunk: []byte("czczcz"),
					RType:     U,
					IsFirst:   true,
				},
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("azazazaza"),
					RType:     S,
				},
				{
					FileInfo:  true,
					FieldName: "bob",
					FileName:  "long.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("11111111111"),
					RType:     S,
				},
				{
					FieldName: "david",
					ByteChunk: []byte("dzdzdz"),
					RType:     U,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("bzbzbzbzbz"),
					IsLast:    true,
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "bob",
					ByteChunk: []byte("222222222222"),
					RType:     S,
					IsLast:    true,
				},
			},
			wantOK: true,
		},
		{
			name: "unary + stream + unary",
			want: []GRequest{
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("1111111111"),
					RType:     S,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("2222222222"),
					RType:     S,
				},
				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzb"),
					RType:     U,
				},
				{
					FieldName: "claire",
					ByteChunk: []byte("czczc"),
					RType:     U,
				},
			},
			got: []GRequest{
				{
					FieldName: "bob",
					ByteChunk: []byte("bzbzb"),
					RType:     U,
					IsFirst:   true,
				},
				{
					FileInfo:  true,
					FieldName: "alice",
					FileName:  "short.txt",
					RType:     S,
				},

				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("1111111111"),
					RType:     S,
				},
				{
					FieldName: "claire",
					ByteChunk: []byte("czczc"),
					RType:     U,
				},
				{
					FileData:  true,
					FieldName: "alice",
					ByteChunk: []byte("2222222222"),
					IsLast:    true,
					RType:     S,
				},
			},
			wantOK: true,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {

			s.Equal(v.wantOK, IsOk(v.name, v.want, v.got))

		})
	}
}
