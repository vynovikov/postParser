package repo

import (
	"errors"
	"postParser/internal/logger"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
)

type byteOpsSuite struct {
	suite.Suite
}

func TestByteOps(t *testing.T) {
	suite.Run(t, new(byteOpsSuite))
}

func (s *byteOpsSuite) TestLineEndPosLimitB() {
	bs := []byte("012345" + Sep)

	p := LineEndPosLimitB(bs, 1, 10)
	s.Equal(6, p)

}

func (s *byteOpsSuite) TestReverse() {
	bs := []byte("012345")

	bbs := Reverse(bs)

	s.Equal([]byte("543210"), bbs)
}

func (s *byteOpsSuite) TestPrevLineLimit() {
	bs := []byte("11111" + Sep + "22222" + Sep + "3333")

	bsPrev := LineLeftLimit(bs, 10, len(bs))

	s.Equal([]byte("2222"), bsPrev)
}

func (s *byteOpsSuite) TestFindBoundaryB() {
	bs := []byte("1111" + Sep + "2222" + Sep + "3333" + Sep + BoundaryField + "bRoot" + Sep + "4444" + Sep + "bPrefix" + "bRoot")

	boundary := FindBoundary(bs)

	s.True(cmp.Equal(boundary, Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}))
}

func (s *byteOpsSuite) TestGenBoundary() {
	boundaryVoc := Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}

	boundaryCalced := GenBoundary(boundaryVoc)

	s.Equal([]byte("bPrefix"+"bRoot"), boundaryCalced)
}

func (s *byteOpsSuite) TestLineStartPosLimitB() {
	bs := []byte("a12345" + Sep + "b12345" + Sep)
	p := LineStartPosLimitB(bs, 10, len(bs))

	s.Equal(8, p)
	s.Equal(string(bs[p]), "b")
}

func (s *byteOpsSuite) TestPartlyBoundaryLen() {
	bs := []byte("a12345" + Sep + "b12345" + Sep + "bPrefix" + "bRo")
	b := []byte("bPrefixbRoot")

	s.Equal(10, PartlyBoundaryLenLeft(bs, b))
}

func (s *byteOpsSuite) TestSlicer() {

	tt := []struct {
		name     string
		bs       []byte
		boundary []byte
		wantedB  AppPieceUnit
		wantedM  []AppPieceUnit
		wantedE  AppPieceUnit
	}{
		{
			name:     "no boundary at all",
			bs:       []byte("a12345" + Sep + "b12345"),
			boundary: []byte("bPrefixbRoot"),

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: true}, APB: AppPieceBody{B: []byte("a12345" + Sep + "b12345")}},
			wantedM: nil,
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: true, E: true}, APB: AppPieceBody{B: []byte("a12345" + Sep + "b12345")}},
		},
		{
			name:     "1 partly boundary",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoo"),
			boundary: []byte("bPrefixbRoot"),

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: false}, APB: AppPieceBody{B: []byte("a12345")}},
			wantedM: nil,
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoo")}},
		},
		{
			name:     "1 full boundary",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoot" + Sep + "b12345"),
			boundary: []byte("bPrefixbRoot"),

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: false}, APB: AppPieceBody{B: []byte("a12345")}},
			wantedM: nil,
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "b12345")}},
		},
		{
			name:     "3 full boundaries",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoot" + Sep + "b12345" + Sep + "bPrefix" + "bRoot" + Sep + "c12345"),
			boundary: []byte("bPrefixbRoot"),

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: false}, APB: AppPieceBody{B: []byte("a12345")}},
			wantedM: []AppPieceUnit{
				{APH: AppPieceHeader{B: false, E: false}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "b12345")}},
			},
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "c12345")}},
		},
		{
			name:     "3 full + partly boundaries",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoot" + Sep + "b12345" + Sep + "bPrefix" + "bRoot" + Sep + "c12345" + Sep + "bPrefix" + "bRo"),
			boundary: []byte("bPrefixbRoot"),

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: false}, APB: AppPieceBody{B: []byte("a12345")}},
			wantedM: []AppPieceUnit{
				{APH: AppPieceHeader{B: false, E: false}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "b12345")}},
				{APH: AppPieceHeader{B: false, E: false}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "c12345")}},
			},
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRo")}},
		},
	}
	for _, tc := range tt {
		s.Run(tc.name, func() {
			b, m, e := Slicer(tc.bs, tc.boundary)

			s.True(cmp.Equal(tc.wantedB, b))
			s.True(cmp.Equal(tc.wantedM, m))
			s.True(cmp.Equal(tc.wantedE, e))
		})
	}
}

func (s *byteOpsSuite) TestSliceParser() {
	piece1 := AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRo")}}
	/*
		piece2 := []AppPieceUnit{
			{APH: AppPieceHeader{B: false, E: false}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "b12345")}},
			{APH: AppPieceHeader{B: false, E: false}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "c12345")}},
		}
	*/
	boundary := Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}
	//piece2 = append(piece2, piece1)

	SliceParser(piece1, boundary)

	s.True(true)
}

func (s *byteOpsSuite) TestLineRightLimit() {

	tt := []struct {
		name      string
		bs        []byte
		fromIndex int
		want      []byte
	}{
		{
			name:      "without Sep",
			bs:        []byte("11111" + Sep + "222222" + Sep + "33333333333" + Sep + "444444444444"),
			fromIndex: 0,
			want:      []byte("11111"),
		},
		{
			name:      "with Sep",
			bs:        []byte("\r\n11111" + Sep + "222222" + Sep + "33333333333" + Sep + "444444444444"),
			fromIndex: 0,
			want:      []byte("11111"),
		},
	}
	for _, tc := range tt {
		s.Run(tc.name, func() {
			l := LineRightLimit(tc.bs, tc.fromIndex, len(tc.bs))
			s.Equal(tc.want, l)
		})
	}

}

func (s *byteOpsSuite) TestSingleLineRightTrimmed() {
	tt := []struct {
		name      string
		bs        []byte
		limit     int
		wantValue []byte
		wantError error
	}{
		{
			name:      "err zero length",
			bs:        []byte(""),
			limit:     3,
			wantValue: nil,
			wantError: errors.New("passed byte slice with zero length"),
		},
		{
			name:      "err no actual characters",
			bs:        []byte("\r\n\r\n\r\n\r\n"),
			limit:     3,
			wantValue: nil,
			wantError: errors.New("no actual characters before limit"),
		},
		{
			name:      "happy without Sep",
			bs:        []byte("11111" + Sep + "222222" + Sep + "33333333333" + Sep + "444444444444"),
			limit:     10,
			wantValue: []byte("11111"),
			wantError: nil,
		},
		{
			name:      "happy with Sep",
			bs:        []byte("\r\n11111" + Sep + "222222" + Sep + "33333333333" + Sep + "444444444444"),
			limit:     12,
			wantValue: []byte("11111"),
			wantError: nil,
		},
		{
			name:      "happy last boundary part",
			bs:        []byte("-"),
			limit:     12,
			wantValue: []byte("-"),
			wantError: nil,
		},
	}
	for _, tc := range tt {
		s.Run(tc.name, func() {
			l, err := SingleLineRightTrimmed(tc.bs, tc.limit)

			s.Equal(tc.wantValue, l)

			if err != nil {
				s.Equal(tc.wantError, err)
			} else {
				s.NoError(err)
			}

		})
	}
}

func (s *byteOpsSuite) TestSingleLineRight() {
	bs := []byte("11111" + Sep + "222222" + Sep + "33333333333" + Sep + "444444444444")
	l, err := SingleLineRightUnchanged(bs, 28)

	s.NoError(err)
	s.Equal([]byte("11111"), l)
}

func (s *byteOpsSuite) TestIsPartlyBoundaryRight() {
	bs := []byte("Root")
	boundary := Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}

	s.True(IsPartlyBoundaryRight(bs, boundary))
}

func (s *byteOpsSuite) TestIsPrintable() {
	bs := []byte("Root \"")

	s.True(IsPrintable(bs))
}

func (s *byteOpsSuite) TestNoDigits() {
	tt := []struct {
		name string
		bs   []byte
	}{
		{
			name: "absent",
			bs:   []byte("afahf _=\""),
		},
		{
			name: "present",
			bs:   []byte("sadjfd345_\""),
		},
	}
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.name == "absent" {
				s.True(NoDigits(tc.bs))
			}
			if tc.name == "present" {
				s.False(NoDigits(tc.bs))
			}
		})
	}
}

func (s *byteOpsSuite) TestAllPrintalbe() {
	tt := []struct {
		name string
		bs   []byte
		want bool
	}{
		{
			name: "all printable",
			bs:   []byte("afahf _="),
			want: true,
		},
		{
			name: "have at least 1 NonPrintable",
			bs:   []byte("sadjfd345_\r777"),
			want: false,
		},
	}
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.want {
				s.True(AllPrintalbe(tc.bs))
			}
			if !tc.want {
				s.False(AllPrintalbe(tc.bs))
			}

		})
	}
}

func (s *byteOpsSuite) TestGetLinesRight() {
	tt := []struct {
		name string
		bs   []byte
		voc  Vocabulaty
		want [][]byte
	}{
		{
			name: "happy first line separated",
			bs:   []byte("oot" + Sep + "Content-Disposition: form-data; name=\"david\"; filename=\"digits.txt\"" + Sep + "Content-Type: text/plain" + Sep + "111111111"),
			voc:  Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CType: "Content-Type: "},
			want: [][]byte{
				[]byte("oot"),
				[]byte("Content-Disposition: form-data; name=\"david\"; filename=\"digits.txt\""),
				[]byte("Content-Type: text/plain"),
			},
		},
		{
			name: "happy second line separated",
			bs:   []byte("ent-Disposition: form-data; name=\"david\"; filename=\"digits.txt\"" + Sep + "Content-Type: text/plain" + Sep + "111111111"),
			voc:  Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CType: "Content-Type: "},
			want: [][]byte{
				[]byte("ent-Disposition: form-data; name=\"david\"; filename=\"digits.txt\""),
				[]byte("Content-Type: text/plain"),
			},
		},
		{
			name: "happy third line separated",
			bs:   []byte("ype: text/plain" + Sep + "aaaaaaaaaaaaa" + Sep + "2222222222222"),
			voc:  Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CType: "Content-Type: "},
			want: [][]byte{
				[]byte("ype: text/plain"),
			},
		},
		{
			name: "happy separator before 1 line",
			bs:   []byte("\r\n" + "bPrefix" + "bRoot" + Sep + "Content-Disposition: form-data; name=\"david\"; filename=\"digits.txt\"" + Sep + "Content-Type: text/plain" + Sep + "111111111"),
			voc:  Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CType: "Content-Type: "},
			want: [][]byte{
				[]byte("bPrefix" + "bRoot"),
				[]byte("Content-Disposition: form-data; name=\"david\"; filename=\"digits.txt\""),
				[]byte("Content-Type: text/plain"),
			},
		},
		{
			name: "happy separator after 1 line",
			bs:   []byte("\n" + "Content-Disposition: form-data; name=\"david\"; filename=\"digits.txt\"" + Sep + "Content-Type: text/plain" + Sep + "111111111"),
			voc:  Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CType: "Content-Type: "},
			want: [][]byte{
				[]byte("Content-Disposition: form-data; name=\"david\"; filename=\"digits.txt\""),
				[]byte("Content-Type: text/plain"),
			},
		},
		{
			name: "happy separator after 2 line",
			bs:   []byte("\n" + "Content-Type: text/plain" + Sep + "111111111"),
			voc:  Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CType: "Content-Type: "},
			want: [][]byte{
				[]byte("Content-Type: text/plain"),
			},
		},
	}
	for _, tc := range tt {
		s.Run(tc.name, func() {
			got, err := GetLinesRight(tc.bs, MaxHeaderLimit, tc.voc)
			logger.L.Infof("in repo.TestGetLinesRight lines: %q\n", got)
			if err != nil {
				logger.L.Infof("in repo.TestGetLinesRight err: %v\n", err)
			}

			s.Equal(tc.want, got)
		})
	}
}

func (s *byteOpsSuite) TestCurrentLineFirstPrintIndexLeft() {
	tt := []struct {
		name      string
		bs        []byte
		wantValue int
		wantError error
	}{
		{
			name:      "happy",
			bs:        []byte("12345\r\n"),
			wantValue: 4,
			wantError: nil,
		},
		{
			name:      "unhappy no printable",
			bs:        []byte("\n\r\n\r\n\r\n\r\r\n"),
			wantValue: -1,
			wantError: errors.New("no actual characters before limit"),
		},
		{
			name:      "unhappy zero lenght",
			bs:        []byte(""),
			wantValue: -1,
			wantError: errors.New("passed byte slice with zero length"),
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			i, err := CurrentLineFirstPrintIndexLeft(v.bs, len(v.bs)-2)
			if err != nil {
				s.Equal(v.wantError, err)

			} else {
				s.NoError(err)
			}
			s.Equal(v.wantValue, i)
		})
	}

}

func (s *byteOpsSuite) TestGetCurrentLineLeft() {
	tt := []struct {
		name      string
		bs        []byte
		fromIndex int
		limit     int
		wantValue []byte
		wantError error
	}{
		{
			name:      "happy",
			bs:        []byte("\r\n12345\r\n"),
			fromIndex: 6,
			limit:     9,
			wantValue: []byte("12345"),
			wantError: nil,
		},
		{
			name:      "unhappy limit exceeded",
			bs:        []byte("\r\n0123456789\r\n"),
			fromIndex: 11,
			limit:     8,
			wantValue: nil,
			wantError: errors.New("line limit exceeded. No separator met"),
		},
		{
			name:      "unhappy zero length",
			bs:        []byte(""),
			fromIndex: 11,
			limit:     8,
			wantValue: []byte{},
			wantError: errors.New("passed byte slice with zero length"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := GetCurrentLineLeft(v.bs, v.fromIndex, v.limit)
			if err != nil {
				s.Equal(v.wantError, err)
			} else {
				s.NoError(err)
			}
			s.Equal(v.wantValue, got)
		})
	}
}
func (s *byteOpsSuite) TestSingleLineLeftTrimmed() {
	tt := []struct {
		name      string
		bs        []byte
		limit     int
		wantValue []byte
		wantError error
	}{
		{
			name:      "happy",
			bs:        []byte("\r\n12345\r\n"),
			limit:     9,
			wantValue: []byte("12345"),
			wantError: nil,
		},
		{
			name:      "unhappy limit exceeded",
			bs:        []byte("\r\n0123456789\r\n"),
			limit:     8,
			wantValue: nil,
			wantError: errors.New("line limit exceeded. No separator met"),
		},
		{
			name:      "unhappy zero length",
			bs:        []byte{},
			limit:     8,
			wantValue: nil,
			wantError: errors.New("passed byte slice with zero length"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := SingleLineLeftTrimmed(v.bs, v.limit)
			if err != nil {
				s.Equal(v.wantError, err)
			} else {
				s.NoError(err)
			}
			s.Equal(v.wantValue, got)
		})
	}
}
func (s *byteOpsSuite) TestGetLinesLeft() {
	tt := []struct {
		name      string
		bs        []byte
		limit     int
		voc       Vocabulaty
		wantValue [][]byte
		wantError error
	}{
		{
			name:  "happy 1 line separated",
			bs:    []byte("\r\nbPrefi"),
			limit: 9,
			voc:   Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CType: "Content-Type"},
			wantValue: [][]byte{
				[]byte("bPrefi"),
			},
			wantError: nil,
		},
		{
			name:  "happy 2 line separated",
			bs:    []byte("\r\nbPrefixbRoot\r\nContent-Disposit"),
			limit: 40,
			voc:   Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CDisposition: "Content-Disposition", CType: "Content-Type"},
			wantValue: [][]byte{
				[]byte("Content-Disposit"),
				[]byte("bPrefixbRoot"),
			},
			wantError: nil,
		},
		{
			name:  "happy 3 line separated",
			bs:    []byte("\r\nbPrefixbRoot\r\nContent-Disposition :\r\nContent-Type"),
			limit: 60,
			voc:   Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CDisposition: "Content-Disposition", CType: "Content-Type"},
			wantValue: [][]byte{
				[]byte("Content-Type"),
				[]byte("Content-Disposition :"),
				[]byte("bPrefixbRoot"),
			},
			wantError: nil,
		},
		{
			name:  "happy third line separator separated",
			bs:    []byte("\r\nbPrefixbRoot\r\nContent-Disposition :\r\nContent-Type :\r"),
			limit: 60,
			voc:   Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CDisposition: "Content-Disposition", CType: "Content-Type"},
			wantValue: [][]byte{
				[]byte("Content-Type :"),
				[]byte("Content-Disposition :"),
				[]byte("bPrefixbRoot"),
			},
			wantError: nil,
		},
		{
			name:  "happy second line separator separated",
			bs:    []byte("\r\nbPrefixbRoot\r\nContent-Disposition :\r"),
			limit: 60,
			voc:   Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CDisposition: "Content-Disposition", CType: "Content-Type"},
			wantValue: [][]byte{
				[]byte("Content-Disposition :"),
				[]byte("bPrefixbRoot"),
			},
			wantError: nil,
		},
		{
			name:  "happy first line separator separated",
			bs:    []byte("\r\nbPrefixbRoot\r"),
			limit: 60,
			voc:   Vocabulaty{Boundary: Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}, CDisposition: "Content-Disposition", CType: "Content-Type"},
			wantValue: [][]byte{
				[]byte("bPrefixbRoot"),
			},
			wantError: nil,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := GetLinesLeft(v.bs, v.limit, v.voc)
			if err != nil {
				s.Equal(v.wantError, err)
			} else {
				s.NoError(err)
			}
			s.Equal(v.wantValue, got)
		})
	}
}

func (s *byteOpsSuite) TestPartlyBoundaryLeft() {

	tt := []struct {
		name string
		bs   []byte
		bou  Boundary
		want []byte
	}{
		{
			name: "bPrefix separated",
			bs:   []byte("1111111111" + Sep + "bPref"),
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			want: []byte("bPref"),
		},
		{
			name: "bRoot separated",
			bs:   []byte("1111111111" + Sep + "bPrefix" + "bRo"),
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			want: []byte("bPrefix" + "bRo"),
		},
		{
			name: "bSuffix separated",
			bs:   []byte("1111111111" + Sep + "bPrefix" + "bRoot" + "bSuf"),
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
			want: []byte("bPrefix" + "bRoot" + "bSuf"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := PartlyBoundaryLeft(v.bs, v.bou)

			s.NoError(err)
			s.Equal(v.want, got)
		})

	}

}

func (s *byteOpsSuite) TestPartlyBoundaryRight() {
	tt := []struct {
		name  string
		bs    []byte
		limit int
		want  []byte
	}{
		{
			name:  "happy last bPrefix separated",
			bs:    []byte("efix" + "bRoot" + "bSuffix"),
			limit: len("efix" + "bRoot" + "bSuffix"),
			want:  []byte("efix" + "bRoot" + "bSuffix"),
		},
		{
			name:  "happy last bRoot separated ",
			bs:    []byte("oot" + "bSuffix"),
			limit: len("oot" + "bSuffix"),
			want:  []byte("oot" + "bSuffix"),
		},
		{
			name:  "happy not last bRoot separated ",
			bs:    []byte("oot" + Sep + "111111111"),
			limit: 8,
			want:  []byte("oot"),
		},
		{
			name:  "happy last bSuffix separated",
			bs:    []byte("fix"),
			limit: len("fix"),
			want:  []byte("fix"),
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := PartlyBoundaryRight(v.bs, v.limit)
			s.NoError(err)
			s.Equal(v.want, got)
		})
	}
}

func (s *byteOpsSuite) TestGetLinesMiddle() {
	tt := []struct {
		name  string
		bs    []byte
		limit int
		voc   Vocabulaty
		want  [][]byte
	}{
		{
			name: "happy 2 line header",
			bs: []byte("Content-Disposition: form-data; name=\"claire\"; filename=\"digits.txt\"" + Sep + "Content-Type: text/plain" +
				Sep + "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111" +
				Sep + "2222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222"),
			limit: MaxHeaderLimit,
			voc:   Vocabulaty{},
			want: [][]byte{
				[]byte("Content-Disposition: form-data; name=\"claire\"; filename=\"digits.txt\""),
				[]byte("Content-Type: text/plain"),
			},
		},
		{
			name: "happy 1 line header",
			bs: []byte("Content-Disposition: form-data; name=\"alice\"" +
				Sep + "Who the hell is Alice?"),
			limit: MaxHeaderLimit,
			voc:   Vocabulaty{},
			want: [][]byte{
				[]byte("Content-Disposition: form-data; name=\"alice\""),
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := GetLinesMiddle(v.bs, v.limit, v.voc)
			s.NoError(err)
			s.Equal(v.want, got)
		})
	}
}
