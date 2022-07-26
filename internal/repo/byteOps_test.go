package repo

import (
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

func (s *byteOpsSuite) TestReverseB() {
	bs := []byte("012345")

	bbs := ReverseB(bs)

	s.Equal([]byte("543210"), bbs)
}

func (s *byteOpsSuite) TestPrevLineLimit() {
	bs := []byte("11111" + Sep + "22222" + Sep + "3333")

	bsPrev := PrevLineLimit(bs, 10, len(bs))

	s.Equal([]byte("2222"), bsPrev)
}

func (s *byteOpsSuite) TestFindBoundaryB() {
	bs := []byte("1111" + Sep + "2222" + Sep + "3333" + Sep + BoundaryField + "bRoot" + Sep + "4444" + Sep + "bPrefix" + "bRoot")

	boundary := FindBoundaryB(bs)

	s.True(cmp.Equal(boundary, BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}))
}

func (s *byteOpsSuite) TestGenBoundary() {
	boundaryVoc := BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}

	boundaryCalced := genBoundary(boundaryVoc)

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
	b := BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")}

	s.Equal(10, PartlyBoundaryLen(bs, b))
}

func (s *byteOpsSuite) TestSlicer() {

	tt := []struct {
		name     string
		bs       []byte
		boundary BoundaryB
		wantedB  AppPieceUnit
		wantedM  []AppPieceUnit
		wantedE  AppPieceUnit
	}{
		{
			name:     "no boundary at all",
			bs:       []byte("a12345" + Sep + "b12345"),
			boundary: BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: true}, APB: AppPieceBody{B: []byte("a12345" + Sep + "b12345")}},
			wantedM: nil,
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: true, E: true}, APB: AppPieceBody{B: []byte("a12345" + Sep + "b12345")}},
		},
		{
			name:     "1 partly boundary",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoo"),
			boundary: BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: false}, APB: AppPieceBody{B: []byte("a12345")}},
			wantedM: nil,
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoo")}},
		},
		{
			name:     "1 full boundary",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoot" + Sep + "b12345"),
			boundary: BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: false}, APB: AppPieceBody{B: []byte("a12345")}},
			wantedM: nil,
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "b12345")}},
		},
		{
			name:     "3 full boundaries",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoot" + Sep + "b12345" + Sep + "bPrefix" + "bRoot" + Sep + "c12345"),
			boundary: BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},

			wantedB: AppPieceUnit{APH: AppPieceHeader{B: true, E: false}, APB: AppPieceBody{B: []byte("a12345")}},
			wantedM: []AppPieceUnit{
				{APH: AppPieceHeader{B: false, E: false}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "b12345")}},
			},
			wantedE: AppPieceUnit{APH: AppPieceHeader{B: false, E: true}, APB: AppPieceBody{B: []byte("bPrefix" + "bRoot" + Sep + "c12345")}},
		},
		{
			name:     "3 full + partly boundaries",
			bs:       []byte("a12345" + Sep + "bPrefix" + "bRoot" + Sep + "b12345" + Sep + "bPrefix" + "bRoot" + Sep + "c12345" + Sep + "bPrefix" + "bRo"),
			boundary: BoundaryB{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},

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
