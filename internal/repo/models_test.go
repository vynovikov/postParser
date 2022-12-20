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

func (s *modelsSuite) TestL() {
	tt := []struct {
		name    string
		apu     AppPieceUnit
		bou     Boundary
		wantedL [][]byte
	}{
		{
			name: "happy",
			apu:  AppPieceUnit{APH: AppPieceHeader{Part: 2, TS: "qqq", B: true}, APB: AppPieceBody{B: []byte("rm-data; name=\"claire\"; filename=\"short.txt\"\r\n\r\nContent-Type: text/plain\r\n")}},
			bou:  Boundary{},
			wantedL: [][]byte{
				[]byte("rm-data; name=\"claire\"; filename=\"short.txt\""),
				[]byte("Content-Type: text/plain"),
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			lines, _, _ := v.apu.L(v.bou)
			s.Equal(v.wantedL, lines)
		})
	}
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
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: false, E: True}, APB: AppPieceBody{B: []byte("0123456789")}},
			cut:     3,
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: false, E: True}, APB: AppPieceBody{B: []byte("3456789")}},
		},

		{
			name:    "cut == bodyLen",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: false, E: True}, APB: AppPieceBody{B: []byte("0123456789")}},
			cut:     10,
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 3, TS: "qqq", B: false, E: True}, APB: AppPieceBody{B: []byte("")}},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			v.d.BodyCut(v.cut)
			s.Equal(v.wantedD, v.d)
		})
	}

}

func (s *modelsSuite) TestNewDispositionFilled() {
	tt := []struct {
		name        string
		d           DataPiece
		bou         Boundary
		wantedDispo Disposition
		wantedD     DataPiece
		wantedError error
	}{
		{
			name:    "1 line header",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"\r\n")}},
			bou:     Boundary{},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte{}}},
			wantedDispo: Disposition{
				H:        []byte("Content-Disposition: form-data; name=\"claire\"\r\n"),
				FormName: "claire",
			},
			wantedError: nil,
		},

		{
			name:    "2 line header",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n")}},
			bou:     Boundary{},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte{}}},
			wantedDispo: Disposition{
				H:        []byte("Content-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
				FormName: "claire",
				FileName: "short.txt",
			},
			wantedError: nil,
		},

		{
			name:    "full header && some data",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\ndsahjfgdsajfkgdsgafajhsd\r\n23te8237ter732te732e")}},
			bou:     Boundary{},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("dsahjfgdsajfkgdsgafajhsd\r\n23te8237ter732te732e")}},
			wantedDispo: Disposition{
				H:        []byte("Content-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
				FormName: "claire",
				FileName: "short.txt",
			},
			wantedError: nil,
		},

		{
			name:    "partial header 1 line right",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"")}},
			bou:     Boundary{},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte{}}},
			wantedDispo: Disposition{
				H: []byte("Content-Disposition: form-data; name=\"claire\""),
			},
			wantedError: errors.New("in repo.NewDispositionFilled header \"Content-Disposition: form-data; name=\"claire\"\" is not full"),
		},

		{
			name:    "partial header 1 line left",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("ain\r\nsdahfhjsdgfhsdjfsdf")}},
			bou:     Boundary{},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("sdahfhjsdgfhsdjfsdf")}},
			wantedDispo: Disposition{
				H: []byte("ain\r\n"),
			},
			wantedError: errors.New("in repo.NewDispositionFilled header \"ain\r\n\" is not full"),
		},

		{
			name:    "partial header 2 lines right",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text")}},
			bou:     Boundary{},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte{}}},
			wantedDispo: Disposition{
				H: []byte("Content-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text"),
			},

			wantedError: errors.New("in repo.NewDispositionFilled header \"Content-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text\" is not full"),
		},

		{
			name:    "partial header 2 lines left, body line",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("m-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\nazazabzbzbzb")}},
			bou:     Boundary{},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("azazabzbzbzb")}},
			wantedDispo: Disposition{
				H: []byte("m-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
			},

			wantedError: errors.New("in repo.NewDispositionFilled header \"m-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\" is not full"),
		},

		{
			name:    "partial header 3 lines left, body line",
			d:       &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("oot\r\nContent-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\nazazabzbzbzb")}},
			bou:     Boundary{Prefix: []byte("bPrefiz"), Root: []byte("bRoot")},
			wantedD: &AppPieceUnit{APH: AppPieceHeader{Part: 0, TS: "qqq", E: True}, APB: AppPieceBody{B: []byte("azazabzbzbzb")}},
			wantedDispo: Disposition{
				H: []byte("oot\r\nContent-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n"),
			},

			wantedError: errors.New("in repo.NewDispositionFilled header \"oot\r\nContent-Disposition: form-data; name=\"claire\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\" is not full"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotDispo, err := NewDispositionFilled(v.d, v.bou)
			if err != nil || v.wantedError != nil {
				s.Equal(v.wantedError, err)
			}
			s.Equal(v.wantedD, v.d)
			s.Equal(v.wantedDispo, gotDispo)
		})
	}
}

/*
	func (s *modelsSuite) TestNewAppDistributorUnitUnary() {
		tt := []struct {
			name      string
			d         DataPiece
			wantedADU AppDistributorUnit
		}{
			{
				name: "FormName only",
				d: &AppPieceUnit{
					APH: AppPieceHeader{TS: "qqq", Part: 2, B: false, E: False}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"\r\n\r\nazazaza")},
				},
				wantedADU: AppDistributorUnit{H: AppDistributorHeader{T: Unary, S: StreamData{}, U: UnaryData{TS: "qqq", F: FiFo{FormName: "alice"}}}, B: AppDistributorBody{B: []byte("azazaza")}},
			},

			{
				name: "FormName and fileName",
				d: &AppPieceUnit{
					APH: AppPieceHeader{TS: "qqq", Part: 2, B: false, E: False}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"short.txt\"\r\nContent-Type: text/plain\r\n\r\nazazaza")},
				},
				wantedADU: AppDistributorUnit{H: AppDistributorHeader{T: Unary, S: StreamData{}, U: UnaryData{TS: "qqq", F: FiFo{FormName: "alice", FileName: "short.txt"}}}, B: AppDistributorBody{B: []byte("azazaza")}},
			},
		}
		for _, v := range tt {
			s.Run(v.name, func() {
				s.Equal(v.wantedADU, NewAppDistributorUnitUnary(v.d))
			})
		}
	}

	func (s *modelsSuite) TestNewDispositionFilled() {
		tt := []struct {
			name      string
			d         DataPiece
			bou       Boundary
			wantedD   Disposition
			wantedErr error
		}{
			{
				name: "<1 line of header",
				d: &AppPieceUnit{
					APH: AppPieceHeader{}, APB: AppPieceBody{B: []byte("Content-Dis")},
				},
				bou:       Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
				wantedD:   Disposition{H: [][]byte{[]byte("Content-Dis")}},
				wantedErr: errors.New("in repo.NewDispositionFilled header \"Content-Dis\" is not full"),
			},

			{
				name: "exactly 1 line of header",
				d: &AppPieceUnit{
					APH: AppPieceHeader{}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"long.txt\"\r")},
				},
				bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
				wantedD: Disposition{H: [][]byte{
					[]byte("Content-Disposition: form-data; name=\"alice\"; filename=\"long.txt\""),
				}},
				wantedErr: errors.New("in repo.NewDispositionFilled header \"Content-Disposition: form-data; name=\"alice\"; filename=\"long.txt\"\" is not full"),
			},

			{
				name: "<2 lines of header",
				d: &AppPieceUnit{
					APH: AppPieceHeader{}, APB: AppPieceBody{B: []byte("Content-Disposition: form-data; name=\"alice\"; filename=\"long.txt\"\r\nConte")},
				},
				bou: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
				wantedD: Disposition{H: [][]byte{
					[]byte("Content-Disposition: form-data; name=\"alice\"; filename=\"long.txt\""),
					[]byte("Conte"),
				}},
				wantedErr: errors.New("in repo.NewDispositionFilled header \"Content-Disposition: form-data; name=\"alice\"; filename=\"long.txt\"\r\nConte\" is not full"),
			},
		}
		for _, v := range tt {
			s.Run(v.name, func() {
				gotD, gotErr := NewDispositionFilled(v.d, v.bou)
				if v.wantedErr != nil {
					s.Equal(v.wantedErr, gotErr)
				}
				s.Equal(v.wantedD, gotD)
			})
		}
	}
*/
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
