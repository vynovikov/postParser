package core

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type coreSuite struct {
	suite.Suite
}

func TestCoreSuite(t *testing.T) {
	suite.Run(t, new(coreSuite))
}

/*
func (c *coreSuite) TestParseBegin() {
	ts := repo.RandomString(6)
	afu := repo.AppFeederUnit{
		R: repo.ReceiverUnit{
			H: repo.ReceiverHeader{
				Part: 4,
				TS:   ts,
				Voc:  repo.NewVocabulary(repo.NewBoundary(repo.RandomString(2), repo.RandomString(6), repo.RandomString(2))),
			},
		},
		H: repo.NewAppFeederHeader(repo.NewSepHeader(false, []string{}), nil, 3),
	}
	adu := repo.AppDistributorUnit{
		H: repo.NewAppDistributorHeader(repo.NewMultipartHeader(4), ts, "bob", ""),
		B: repo.NewDistributorBody([]byte{}),
	}
	newAFUH := repo.NewAppFeederHeaderBP(repo.SepHeader{}, repo.SepBody{}, 4)
	line := "\r"

	tc := []struct {
		name         string
		prepatreAFU  func(*repo.AppFeederUnit)
		prepareADU   func(*repo.AppFeederUnit)
		prepareAFUBP func(*repo.AppFeederHeaderBP, repo.SepBody)
	}{
		{
			name: "boundary separated",

			prepatreAFU: func(afu *repo.AppFeederUnit) {
				afu.H.SepHeader.Set(false, []string{(afu.R.H.Voc.Boundary.Prefix + afu.R.H.Voc.Boundary.Root)[:4]})
				afu.R.H.SetPart(4)
				afu.SetBody([]byte((afu.R.H.Voc.Boundary.Prefix + afu.R.H.Voc.Boundary.Root)[4:] + repo.Sep + "Content-Disposition: form-data; name=\"bob\"" + repo.Sep + "Content-Type: text/plain" + repo.Sep + "000000000000000000" + repo.Sep + "12345"))

			},
			prepareADU: func(afu *repo.AppFeederUnit) {
				adu = repo.AppDistributorUnit{
					H: repo.AppDistributorHeader{
						FormName: "bob",
						TS:       ts,
					},
					B: repo.DistributorBody{
						B: []byte("000000000000000000" + repo.Sep + "12345"),
					},
				}
			},
			prepareAFUBP: func(afuH *repo.AppFeederHeaderBP, sb repo.SepBody) {
				afuH = &repo.AppFeederHeaderBP{
					SepHeader: repo.SepHeader{},
					SepBody:   sb,
					PrevPart:  afu.H.PrevPart,
				}
			},
		},
		{
			name: "second line separated",

			prepatreAFU: func(afu *repo.AppFeederUnit) {
				afu.H.SepHeader.Set(false, []string{"Co", "Content-Disposition: form-data; name=\"bob\"", afu.R.H.Voc.Boundary.Prefix + afu.R.H.Voc.Boundary.Root})
				afu.R.H.SetPart(4)
				afu.SetBody([]byte("ntent-Type: text/plain" + repo.Sep + "000000000000000000" + repo.Sep + "12345"))
			},
			prepareADU: func(afu *repo.AppFeederUnit) {
				adu = repo.AppDistributorUnit{
					H: repo.AppDistributorHeader{
						FormName: "bob",
						TS:       ts,
					},
					B: repo.DistributorBody{
						B: []byte("000000000000000000" + repo.Sep + "12345"),
					},
				}
			},
			prepareAFUBP: func(afuH *repo.AppFeederHeaderBP, sb repo.SepBody) {

				afuH = &repo.AppFeederHeaderBP{
					SepHeader: repo.SepHeader{},
					SepBody:   sb,
					PrevPart:  afu.H.PrevPart,
				}
			},
		},
		{
			name: "first line separated",
			prepatreAFU: func(afu *repo.AppFeederUnit) {
				afu.H.SepHeader.Set(false, []string{"Content-Disposition: form-", afu.R.H.Voc.Boundary.Prefix + afu.R.H.Voc.Boundary.Root})
				afu.H.PrevPart = 4
				afu.SetBody([]byte("data; name=\"bob\"" + repo.Sep + "Content-Type: text/plain" + repo.Sep + "000000000000000000" + repo.Sep + "12345"))
			},
			prepareADU: func(afu *repo.AppFeederUnit) {
				adu = repo.AppDistributorUnit{
					H: repo.AppDistributorHeader{
						FormName: "bob",
						TS:       ts,
					},
					B: repo.DistributorBody{
						B: []byte("000000000000000000" + repo.Sep + "12345"),
					},
				}
			},
			prepareAFUBP: func(afuH *repo.AppFeederHeaderBP, sb repo.SepBody) {

				afuH = &repo.AppFeederHeaderBP{
					SepHeader: repo.SepHeader{},
					SepBody:   sb,
					PrevPart:  afu.H.PrevPart,
				}
			},
		},
		{
			name: "between lines",
			prepatreAFU: func(afu *repo.AppFeederUnit) {
				afu.H.SepHeader.Set(false, []string{afu.R.H.Voc.Boundary.Prefix + afu.R.H.Voc.Boundary.Root})
				afu.H.PrevPart = 4
				afu.SetBody([]byte(string(byte(10)) + "Content-Disposition: form-data; name=\"bob\"" + repo.Sep + "Content-Type: text/plain" + repo.Sep + "000000000000000000" + repo.Sep + "12345"))
			},
			prepareADU: func(afu *repo.AppFeederUnit) {
				adu = repo.AppDistributorUnit{
					H: repo.AppDistributorHeader{
						FormName: "bob",
						TS:       ts,
					},
					B: repo.DistributorBody{
						B: []byte("000000000000000000" + repo.Sep + "12345"),
					},
				}
			},
			prepareAFUBP: func(afuH *repo.AppFeederHeaderBP, sb repo.SepBody) {

				afuH = &repo.AppFeederHeaderBP{
					SepHeader: repo.SepHeader{},
					SepBody:   sb,
					PrevPart:  afu.H.PrevPart,
				}
			},
		},
	}

	for _, v := range tc {
		c.Run(v.name, func() {

			v.prepatreAFU(&afu)
			v.prepareADU(&afu)

			sb := repo.NewSepBodyBP(line)
			v.prepareAFUBP(&newAFUH, sb)

			c1 := NewCore()
			aduC, afuHC, err := c1.ParseBegin(afu)

			logger.L.Infof("in repo.TestParseBegin given adu: %v\n", adu)
			logger.L.Infof("in repo.TestParseBegin calculated adu: %v\n", aduC)
			logger.L.Infof("in repo.TestParseBegin difference: %s\n", cmp.Diff(aduC, adu))

			c.NoError(err)
			c.True(cmp.Equal(adu, aduC))
			c.True(cmp.Equal(newAFUH, afuHC))

		})
	}

}
*/
