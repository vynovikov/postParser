package repo

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type repoSuite struct {
	suite.Suite
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(repoSuite))
}

func (s *repoSuite) TestLineStartPosLimit() {
	line := "12345\r\nabcdef"
	p := LineStartPosLimit(line, len(line)-1, len(line))
	s.Equal("abcdef", line[p:])

}

func (s *repoSuite) TestLineEndPosLimit() {
	line := "12345\r\nabcdef"
	p := LineEndPosLimit(line, 0, len(line))
	s.Equal("12345", line[:p])
}

func (s *repoSuite) TestLastPrintPosLimit() {
	line := "12345\r\n\n\r"
	p := LastPrintPosLimit(line, len(line)-1, len(line))
	s.Equal("5", string(line[p]))
}

func (s *repoSuite) TestGetPrevLineLimit() {
	line := "12345\r\nabcdef"
	l := GetPrevLineLimit(line, len(line)-1, len(line))
	s.Equal("abcdef", l)
}

func (s *repoSuite) TestGetLinesFw() {
	v := Vocabulaty{
		Boundary: Boundary{
			Prefix: "12",
			Root:   "3456",
		},
		CType: "bbbb",
	}
	prevLines := []string{"1234"}
	str := "56" + Sep + "Content-Disposition: form-data; name=\"alice\"" + Sep + "bbbb" + Sep + "zzzzzzzzzzzzzzzzzzzzzzz" + Sep + "xxxxxxxxxxxxxxxxxxxxxx"
	L := GetLinesFw(str, prevLines, len("Content-Disposition: form-data; name=\"alice\""), v)
	s.Equal(NewLines([]string{"56", "Content-Disposition: form-data; name=\"alice\"", "bbbb"}, false), L)
}

func (s *repoSuite) TestJoinLines() {
	tc := []struct {
		name string
		pl   []string
		cl   Lines
		want []string
	}{
		{
			name: "separator separated",
			pl:   []string{},
			cl: Lines{
				CurLines: []string{"11111", "22222", "33333"},
				IsWhole:  true,
			},
			want: []string{"11111", "22222", "33333"},
		},
		{
			name: "1 line separated",
			pl:   []string{"11"},
			cl: Lines{
				CurLines: []string{"111", "22222", "33333"},
				IsWhole:  false,
			},
			want: []string{"11111", "22222", "33333"},
		},
		{
			name: "2 line separated",
			pl:   []string{"aaaaa", "bbbbb"},
			cl: Lines{
				CurLines: []string{"11111", "22222"},
				IsWhole:  false,
			},
			want: []string{"bbbbb", "aaaaa11111", "22222"},
		},
		{
			name: "3 line separated",
			pl:   []string{"aaaaa", "bbbbb", "ccccc"},
			cl: Lines{
				CurLines: []string{"11111"},
				IsWhole:  false,
			},
			want: []string{"ccccc", "bbbbb", "aaaaa11111"},
		},

		{
			name: "between lines",
			pl:   []string{"aaaaa", "bbbbb"},
			cl: Lines{
				CurLines: []string{"11111"},
				IsWhole:  true,
			},
			want: []string{"bbbbb", "aaaaa", "11111"},
		},
	}
	for _, v := range tc {
		s.Run(v.name, func() {
			lines := JoinLines(v.pl, v.cl)
			s.Equal(v.want, lines)
		})
	}
}
