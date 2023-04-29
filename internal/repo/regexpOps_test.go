package repo

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type regexpSuite struct {
	suite.Suite
}

func TestRegexpSuite(t *testing.T) {
	suite.Run(t, new(regexpSuite))
}

func (s *regexpSuite) TestIsCDRight() {
	tt := []struct {
		name   string
		b      []byte
		wanted bool
	}{
		{
			name:   "happy_1 len(b)<len(CD)",
			b:      []byte("Content-Disposition: fo"),
			wanted: true,
		},

		{
			name:   "unhappy len(b)<len(CD)",
			b:      []byte("Cotnent-Disposition: na"),
			wanted: false,
		},

		{
			name:   "happy len(b)>len(CD) count(\")==0",
			b:      []byte("Content-Disposition: form-data; name=\"dasdas"),
			wanted: true,
		},

		{
			name:   "happy_2 len(b)>len(CD) count(\")==0",
			b:      []byte("Content-Disposition: form-data; name=\"aZаЯ!]"),
			wanted: true,
		},

		{
			name:   "unhappy len(b)>len(CD) count(\")==0",
			b:      []byte("Content-Disposition: name=\"dasdas "),
			wanted: false,
		},

		{
			name:   "happy_1 len(b)>len(CD) count(\")==1",
			b:      []byte("Content-Disposition: form-data; name=\"aZаЯ!]\""),
			wanted: true,
		},

		{
			name:   "happy_2 len(b)>len(CD) count(\")==1",
			b:      []byte("Content-Disposition: form-data; name=\"aZаЯ!]\"; filename="),
			wanted: true,
		},

		{
			name:   "unhappy_1 len(b)>len(CD) count(\")==1",
			b:      []byte("Content-Disposition: name=\"azaza \""),
			wanted: false,
		},

		{
			name:   "unhappy_2 len(b)>len(CD) count(\")==1",
			b:      []byte("Content-Disposition: name=\"azaza\"a"),
			wanted: false,
		},

		{
			name:   "unhappy_3 len(b)>len(CD) count(\")==1",
			b:      []byte("Content-Disposition: name=\"azaza\r\"; fielname="),
			wanted: false,
		},

		{
			name:   "happy_1 len(b)>len(CD) count(\")==2",
			b:      []byte("Content-Disposition: form-data; name=\"azaza\"; filename=\"sdf"),
			wanted: true,
		},

		{
			name:   "happy_2 len(b)>len(CD) count(\")==2",
			b:      []byte("Content-Disposition: form-data; name=\"azaza\"; filename=\"aZаЯ!_"),
			wanted: true,
		},

		{
			name:   "unhappy_1 len(b)>len(CD) count(\")==2",
			b:      []byte("Content-Disposition: name=\"azaza\"; filename=\"sd\rf"),
			wanted: false,
		},

		{
			name:   "unhappy_2 len(b)>len(CD) count(\")==2",
			b:      []byte("Content-Disposition: name=\"azaza \"; filename=\"sdf"),
			wanted: false,
		},

		{
			name:   "unhappy_3 len(b)>len(CD) count(\")==2",
			b:      []byte("Content-Disposition: name=\"azaza\"; fliename=\"aZаЯ!_"),
			wanted: false,
		},

		{
			name:   "happy_1 len(b)>len(CD) count(\")==3",
			b:      []byte("Content-Disposition: form-data; name=\"azaza\"; filename=\"sdf.xyz\""),
			wanted: true,
		},

		{
			name:   "happy_2 len(b)>len(CD) count(\")==3",
			b:      []byte("Content-Disposition: form-data; name=\"aZаЯ!_\"; filename=\"aZаЯ!_.aZаЯ!_\""),
			wanted: true,
		},

		{
			name:   "unhappy_1 len(b)>len(CD) count(\")==3",
			b:      []byte("Content-Disposition: name=\"azaza\"; fliename=\"sdf.xyz\""),
			wanted: false,
		},

		{
			name:   "unhappy_2 len(b)>len(CD) count(\")==3",
			b:      []byte("Content-Disposition: name=\"azaza \"; filename=\"sdf.xyz\""),
			wanted: false,
		},

		{
			name:   "unhappy_3 len(b)>len(CD) count(\")==3",
			b:      []byte("Content-Disposition: name=\"azaza\"; filename=\"sdf.xyz \""),
			wanted: false,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			s.Equal(v.wanted, IsCDRight(v.b))
		})
	}
}

func (s *regexpSuite) TestIsCDLeft() {
	tt := []struct {
		name   string
		b      []byte
		wanted bool
	}{
		{
			name:   "happy_1 count(\")==1",
			b:      []byte("\""),
			wanted: true,
		},

		{
			name:   "happy_2 count(\")==1",
			b:      []byte("qasw1рпр_\""),
			wanted: true,
		},

		{
			name:   "unhappy_1 count(\")==1",
			b:      []byte(":"),
			wanted: false,
		},

		{
			name:   "unhappy_2 count(\")==1",
			b:      []byte("qasw1рпр \""),
			wanted: false,
		},

		{
			name:   "happy_1 count(\")==2",
			b:      []byte("\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "happy_2 count(\")==2",
			b:      []byte("name=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "happy_3 count(\")==2",
			b:      []byte("Content-Disposition: form-data; name=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "happy_4 count(\")==2",
			b:      []byte("; filename=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "unhappy_1 count(\")==2",
			b:      []byte("name+\"aZ1аЯ_\""),
			wanted: false,
		},

		{
			name:   "unhappy_2 count(\")==2",
			b:      []byte("filename=\"aZ1аЯ \""),
			wanted: false,
		},

		{
			name:   "unhappy_3 count(\")==2",
			b:      []byte("fielname=\"aZ1аЯ\""),
			wanted: false,
		},

		{
			name:   "happy_1 count(\")==3",
			b:      []byte("\"; filename=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "happy_2 count(\")==3",
			b:      []byte("aZ1аЯ_\"; filename=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "unhappy_1 count(\")==3",
			b:      []byte("aZ1аЯ_\" filename=\"aZ1аЯ_\""),
			wanted: false,
		},

		{
			name:   "unhappy_2 count(\")==3",
			b:      []byte("aZ1аЯ \"; filename=\"aZ1аЯ_\""),
			wanted: false,
		},

		{
			name:   "happy_1 count(\")==4",
			b:      []byte("\"aZ1аЯ_\"; filename=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "happy_2 count(\")==4",
			b:      []byte("ition: form-data; name=\"aZ1аЯ_\"; filename=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "happy_3 count(\")==4",
			b:      []byte("Content-Disposition: form-data; name=\"aZ1аЯ_\"; filename=\"aZ1аЯ_\""),
			wanted: true,
		},

		{
			name:   "unhappy_1 count(\")==4",
			b:      []byte("\"aZ1аЯ \"; filename=\"aZ1аЯ_\""),
			wanted: false,
		},

		{
			name:   "unhappy_2 count(\")==4",
			b:      []byte("Cnotent-Disposition:  name=\"aZ1аЯ\"; filename=\"aZ1аЯ_\""),
			wanted: false,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			s.Equal(v.wanted, IsCDLeft(v.b))
		})
	}
}

func (s *regexpSuite) TestIsCTRight() {
	tt := []struct {
		name   string
		b      []byte
		wanted bool
	}{
		{
			name:   "happy len(b)==0",
			b:      []byte(""),
			wanted: true,
		},
		{
			name:   "happy len(b)<=len(CT)",
			b:      []byte("Content-Type:"),
			wanted: true,
		},

		{
			name:   "unhappy len(b)<=len(CT)",
			b:      []byte("Content-Type."),
			wanted: false,
		},

		{
			name:   "happy_1 len(b)>len(CT)",
			b:      []byte("Content-Type: tkjht"),
			wanted: true,
		},

		{
			name:   "happy_2 len(b)>len(CT)",
			b:      []byte("Content-Type: tkj/ht"),
			wanted: true,
		},

		{
			name:   "happy_3 len(b)>len(CT)",
			b:      []byte("Content-Type: tkj/"),
			wanted: true,
		},

		{
			name:   "unhappy_1 len(b)>len(CT)",
			b:      []byte("Content-Type: tkj /ht"),
			wanted: false,
		},

		{
			name:   "unhappy_2 len(b)>len(CT)",
			b:      []byte("Content-Type: tk\rj"),
			wanted: false,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			s.Equal(v.wanted, IsCTRight(v.b))
		})
	}
}

func (s *regexpSuite) TestIsCTLeft() {
	tt := []struct {
		name   string
		b      []byte
		wanted bool
	}{
		{
			name:   "happy_1 len(b)<=len(from end to whitespace index)",
			b:      []byte("aZ1_"),
			wanted: true,
		},

		{
			name:   "happy_2 len(b)<=len(from end to whitespace index)",
			b:      []byte("/aZ1_"),
			wanted: true,
		},

		{
			name:   "unhappy_1 len(b)<=len(from end to whitespace index)",
			b:      []byte("aZ1 _"),
			wanted: false,
		},

		{
			name:   "happy_1 len(b)>len(from end to whitespace index)",
			b:      []byte("Content-Type: tkjht"),
			wanted: true,
		},

		{
			name:   "happy_2 len(b)>len(from end to whitespace index)",
			b:      []byte("Content-Type: tkj/ht"),
			wanted: true,
		},

		{
			name:   "unhappy_1 len(b)>len(from end to whitespace index)",
			b:      []byte("Content-Type: tkj /ht"),
			wanted: false,
		},

		{
			name:   "unhappy_2 len(b)>len(from end to whitespace index)",
			b:      []byte("Content-Type: tk\rj"),
			wanted: false,
		},

		{
			name:   "unhappy_2 len(b)>len(from end to whitespace index)",
			b:      []byte("Content-Type: tk:j"),
			wanted: false,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			s.Equal(v.wanted, IsCTLeft(v.b))
		})
	}
}

func (s *regexpSuite) TestIsCTFull() {
	tt := []struct {
		name   string
		b      []byte
		wanted bool
	}{
		{
			name:   "happy",
			b:      []byte("Content-Type: text/plain\r\n"),
			wanted: true,
		},

		{
			name:   "unhappy 1",
			b:      []byte("Content-Type: text/plain"),
			wanted: false,
		},

		{
			name:   "unhappy 2",
			b:      []byte("aghahagyy"),
			wanted: false,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			s.Equal(v.wanted, IsCTFull(v.b))
		})
	}
}

func (s *regexpSuite) TestIsBoundary() {
	tt := []struct {
		name string
		p    []byte
		n    []byte
		bou  Boundary
	}{
		{
			name: "",
			p:    []byte("bPre"),
			n:    []byte("fixbRoot\r\nContent-Disposition"),
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
		},
	}
	for _, v := range tt {
		s.True(IsBoundary(v.p, v.n, v.bou))
	}
}

func (s *regexpSuite) TestIsLastBoundary() {
	tt := []struct {
		name string
		p    []byte
		n    []byte
		bou  Boundary
	}{
		{
			name: "",
			p:    []byte("bPre"),
			n:    []byte("fixbRootbSuffix\r\n"),
			bou:  Boundary{Prefix: []byte("bPrefix"), Root: []byte("bRoot")},
		},
	}
	for _, v := range tt {
		s.True(IsLastBoundary(v.p, v.n, v.bou))
	}
}
