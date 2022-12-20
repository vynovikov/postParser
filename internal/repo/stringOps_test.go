package repo

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type stringOpsSuite struct {
	suite.Suite
}

func TestStringOpsSuite(t *testing.T) {
	suite.Run(t, new(stringOpsSuite))
}

func (s *stringOpsSuite) TestGetLastIndex() {
	tt := []struct {
		name string
		s    string
		occ  string
		want int
	}{
		{
			name: "happy",
			s:    "012aaa1234aaa123aaa4",
			occ:  "aaa",
			want: 16,
		},
	}
	for _, v := range tt {
		s.Equal(v.want, GetLastIndex(v.s, v.occ))
	}
}
