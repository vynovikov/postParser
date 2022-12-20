package store

import (
	"bytes"
	"errors"
	"fmt"
	"postParser/internal/repo"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
)

type StoreStruct struct {
	R map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue
	B map[repo.AppStoreKeyGeneral][]repo.DataPiece
	C map[repo.AppStoreKeyGeneral]int
	L sync.RWMutex
}

func NewStore() *StoreStruct {
	return &StoreStruct{
		R: make(map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue),
		B: make(map[repo.AppStoreKeyGeneral][]repo.DataPiece),
		C: make(map[repo.AppStoreKeyGeneral]int),
		L: sync.RWMutex{},
	}
}

type Store interface {
	IsEmpty() bool
	DetailedKey(string) repo.AppStoreKeyDetailed
	Register(repo.DataPiece, repo.Boundary) (repo.AppDistributorUnit, error)
	RegisterBuffer(repo.Boundary) ([]repo.AppDistributorUnit, []error)
	Inc(repo.AppStoreKeyGeneral, int)
	Dec(repo.AppStoreKeyGeneral)
	Counter(repo.AppStoreKeyGeneral) int
}

func (s *StoreStruct) IsEmpty() bool {
	return len(s.R) == 0
}

func (s *StoreStruct) HasGeneral(ts string) bool {
	/*
		askg := repo.NewAppStoreKeyGeneral(ts)
		if _, ok := s.R[askg]; ok {
			return true
		}
	*/
	return false
}

func (s *StoreStruct) DetailedPart(ts string) (int, error) {
	/*parts := make([]int, 0)

	askg := repo.NewAppStoreKeyGeneral(ts)
	d, ok := s.R[askg]
	//logger.L.Infof("in store.DetailedPart d = %c, ok %t", d, ok)
	if ok {
		for i, _ := range d {
			parts = append(parts, i.Part)
		}

		if len(parts) > 0 {
			return parts[0], nil
		}
	}
	*/
	return -1, errors.New("not found")

}

func (s *StoreStruct) DetailedKey(ts string) repo.AppStoreKeyDetailed {
	iMax := repo.AppStoreKeyDetailed{}
	/*
		askg := repo.NewAppStoreKeyGeneral(ts)
		d, ok := s.R[askg]

		if ok {
			for i := range d {
				if i.Part > iMax.Part {
					iMax = i
				}
			}

		}
	*/
	return iMax
}

func (s *StoreStruct) Register(d repo.DataPiece, bou repo.Boundary) (repo.AppDistributorUnit, error) {

	//logger.L.Infof("store.Register invoked with d header: %v, body: %q, while s.R: %v, s.C: %v\n", d.GetHeader(), d.GetBody(0), s.R, s.C)

	var (
		du        repo.AppDistributorUnit
		dispo     repo.Disposition
		err       error
		askdParts []int
	)
	if s.B == nil { // delete after testing
		s.B = make(map[repo.AppStoreKeyGeneral][]repo.DataPiece)
	}

	askg := repo.NewAppStoreKeyGeneral(d)
	//logger.L.Infof("in store.Register TS: %q, askg: %q\n", d.TS(), askg)
	askd := repo.NewAppStoreKeyDetailed(d)
	//logger.L.Infof("in store.Register TS: %q, askd: %q\n", d.TS(), askd)
	b := repo.NewBeginningData(d.TS(), d.Part())

	switch {

	case d.B(): //DataPiece needs beginning

		if m1, ok := s.R[askg]; ok { // General key met

			if m2, ok := m1[askd]; ok { // Detailed key met

				//logger.L.Infof("in store.Register specific map is %v\n", m2)
				//logger.L.Infof("in store.Register in dataPiece TS: %s, Part = %d, E() = %d, body: %q\n", d.TS(), d.Part(), d.E(), d.GetBody(0))

				m3f := m2[false]
				//logger.L.Infof("in store.Register right place\n")
				if m3f.D.FormName == "" { // Disposition is not filled
					header, err := d.H(bou)
					//logger.L.Warnf("in store.Register header: %q, error %v\n", header, err)
					if m3t, ok := m2[true]; ok { // true branch is present
						if strings.Contains(err.Error(), "is ending part") {

							m3f.D.H = append(m3f.D.H, m3t.D.H...)
							m3f.D.H = append(m3f.D.H, header...)
							m3f.D.FormName, m3f.D.FileName = repo.GetFoFi(m3f.D.H)
							m3f.E = d.E()

							d.BodyCut(len(header))

							//logger.L.Infof("in store.Register m3t.D.H: %q, m3t.D.FormName: %q, m3t.D.FileName: %q\n", m3t.D.H, m3t.D.FormName, m3t.D.FileName)

							vvv := map[bool]repo.AppStoreValue{}
							vvv[false] = m3f
							delete(s.R[askg], askd)
							s.R[askg][askd.IncPart()] = vvv

							return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), fmt.Errorf("in store.Register new dataPiece group with TS \"%s\" and BeginPart = %d", askg.TS, m3f.B.BeginPart)
						}

					}
					// no true branch
					if strings.Contains(err.Error(), "is ending part") {

						m3f.D.H = append(m3f.D.H, header...)
						m3f.D.FormName, m3f.D.FileName = repo.GetFoFi(m3f.D.H)
						m3f.E = d.E()

						d.BodyCut(len(header))

						vvv := map[bool]repo.AppStoreValue{}
						vvv[false] = m3f

						//logger.L.Infof("in store.Register s.M: %v\n", s.M)
						delete(s.R[askg], askd)

						s.R[askg][askd.IncPart()] = vvv

						//logger.L.Infof("in store.Register s.M: %v\n", s.M)
						//logger.L.Infof("in store.Register adu header: %v, body: %q\n", adu.H, adu.B.B)
						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), fmt.Errorf("in store.Register new dataPiece group with TS \"%s\" and BeginPart = %d", askg.TS, m3f.B.BeginPart)
					}
				}
				// Disposition is filled
				if m3t, ok := m2[true]; ok { // true branch is present

					if m3f.E == repo.True { // false branch needs next part -- still waiting for dataPiece with d.E == repo.Probably which is current one

						//logger.L.Infof("in store.Register is this case with m3f: %q\n", m3f)

						m3f.E = d.E()

						vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{askd.IncPart(): {false: m3f, true: m3t}}
						delete(s.R[askg], askd)
						s.R[askg] = vv
						//logger.L.Warnf("in store.Register s.R: %v\n", s.R)

						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil

					}

					header, err := d.H(bou)
					//logger.L.Warnf("in store.Register in dataPiece TS: %s, Part = %d, E() = %d -> header: %q, error: %v \n", d.TS(), d.Part(), d.E(), header, err)

					if strings.Contains(err.Error(), "is ending part") { // header in dataPiece's body present

						if repo.IsLastBoundary(m3t.D.H, header, bou) {
							adu := repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: d.TS(), Part: m3t.B.BeginPart}, M: repo.Message{S: "finish"}}}}
							delete(s.R, askg)

							return adu, fmt.Errorf("in store.Register dataPiece group with TS \"%s\" is finished", d.TS())
						}

						if repo.IsBoundary(m3t.D.H, header, bou) && bytes.Contains(header, []byte("Content-Disposition")) {

							m3f.D.H = header[bytes.Index(header, []byte("Content-Disposition")):]
							m3f.D.FormName, m3f.D.FileName = repo.GetFoFi(header)
							m3f.B = m3t.B
							m3f.E = d.E()

							d.BodyCut(len(header))

							vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{askd.IncPart(): {false: m3f}}

							delete(s.R[askg], askd)

							s.R[askg] = vv

							return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.StopLast, PostAction: repo.None}), fmt.Errorf("in store.Register new dataPiece group with TS \"%s\" and BeginPart = %d", askg.TS, m3f.B.BeginPart)
						}

						d.Prepend(m3t.D.H)

						m3f.E = d.E()

						vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{askd.IncPart(): {false: m3f}}

						delete(s.R[askg], askd)

						s.R[askg] = vv

						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil

					}
					if strings.Contains(err.Error(), "no header found") {

						if d.E() == repo.Last { // if last dataPiece

							d.Prepend(m3t.D.H)

							adu := repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.Finish})

							//logger.L.Infof("in store.Register adu: %v\n", adu)

							delete(s.R, askg)

							return adu, fmt.Errorf("in store.Register dataPiece group with TS \"%s\" is finished", d.TS())
						}

						m3f.E = d.E()

						vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{askd.IncPart(): {false: m3f}}
						delete(s.R[askg], askd)
						s.R[askg] = vv

						d.Prepend(m3t.D.H)

						return repo.NewDistributorUnitStream(s.R[askg][askd.IncPart()][false], d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil

					}
					if strings.Contains(err.Error(), "is the last") {

						delete(s.R, askg)

						adu := repo.NewDistributorUnitStreamEmpty(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.Finish})

						return adu, fmt.Errorf("in store.Register dataPiece group with TS \"%s\" is finished", d.TS())
					}

				}
				// no true branch
				vvv := map[bool]repo.AppStoreValue{}
				switch d.E() {

				case repo.Probably: // deteiled group can be closed or finished

					//logger.L.Warnf("in store.Register s.R before: %v\n", s.R)
					//logger.L.Warnf("in store.Register m1: %v, len(m1) = %d\n", m1, len(m1))

					if _, ok := m2[true]; ok { // level 3 map has true branch

						//logger.L.Infof("in store.Register got right place with previously saved header line %q and newly body %q\n", m3t.D.H[0], d.GetBody())
						//increment askd

					}
					m2f, ok1 := m1[askd.F()]
					m2t, ok2 := m1[askd.T()]
					if ok1 && ok2 { // there is another ASKD with same part which has true branch only
						//logger.L.Infof("in store.Register m1: %v, askd.T(): %v\n", m1, askd.T())
						m3t := m2t[true]

						//logger.L.Infof("in store.Register m3t: %v\n", m3t)
						m3f := m2f[false]
						m3f.E = d.E()
						//logger.L.Infof("in store.Register m3f: %v\n", m3f)

						vvv := map[bool]repo.AppStoreValue{}
						vvv[true] = m3t
						vvv[false] = m3f

						delete(s.R[askg], askd.T())
						delete(s.R[askg], askd.F())
						//logger.L.Warnf("in store.Register s.R after deleting: %v\n", s.R)
						s.R[askg][askd.IncPart()] = vvv
						//logger.L.Warnf("in store.Register s.R after adding: %v\n", s.R)
						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil
						/*						for i,v:=range m1{
													if m3t,ok:=v[true];ok&&len(v)==1[

													]
												}
						*/
					}
					// no true branch

					m3f.E = d.E()

					delete(s.R[askg], askd)
					s.R[askg][askd] = map[bool]repo.AppStoreValue{false: m3f}

					return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), fmt.Errorf("in store.Register got double-meaning dataPiece")

				case repo.True: // detailed group shall be continued

					vvv[false] = m3f
					/*
						if m2, ok := s.R[askg][askd.IncPart()]; ok && len(m2) == 1 {
							if m3t, ok := m2[true]; ok {
								vvv[true] = m3t
							}
						}
					*/
					delete(s.R[askg], askd)
					//logger.L.Warnf("in store.Register after deleting s.R = %v\n", s.R)

					s.R[askg][askd.IncPart()] = vvv
					//logger.L.Warnf("in store.Register s.R = %v\n", s.R)

					adu := repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.None})
					//logger.L.Warnf("in store.Register adu header: %v, body: %q\n", adu.H, adu.B.B)

					return adu, nil

				case repo.False: // detailed group shall be closed and checked for last boundary

					//logger.L.Infof("in store.Register d.B(), d.E()==repo.False \n")

					if m3t, ok := m2[true]; ok {

						boundaryTrimmed := repo.GenBoundary(bou)[2:]
						lb := append(m3t.D.H, d.GetBody(repo.Min(d.LL(), repo.MaxHeaderLimit))...)

						if repo.LastBoundary(lb, boundaryTrimmed) {
							//logger.L.Infof("in store.Register got last boundary %q\n", lb)
							//s.R = make(map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue)
							//logger.L.Infof("in store.Register deleting askg %q\n", askg)
							delete(s.R, askg)
							//logger.L.Infof("in store.Register after deleting askg s.R: %v\n", s.R)
							return repo.NewDistributorUnitStreamEmpty(m3t, d, repo.Message{PreAction: repo.None, PostAction: repo.Stop}), fmt.Errorf("in store.Register dataPiece group with TS \"%s\" is finished", d.TS())
						}

						delete(s.R[askg], askd)

						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.Stop}), fmt.Errorf("in store.Register dataPiece group with TS %q and BeginPart = %d is finished", m3f.B.TS, m3f.B.BeginPart)
					}
					//logger.L.Infof("in store.Register before deleting s.R %v, len(s.R[askg]) = %d\n", s.R[askg], len(s.R[askg]))

					delete(s.R[askg], askd)

					return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.Stop}), fmt.Errorf("in store.Register dataPiece group with TS %q and BeginPart \"%d\" is finished", m3f.B.TS, m3f.B.BeginPart)

				case repo.Last:
					//logger.L.Infof("in store.Register handling last dataPiece while counter = %d\n", s.Counter(askg))
					if s.Counter(askg) == 1 { // dataPiece hamdled last, s.R may be cleared
						delete(s.R, askg)
						//logger.L.Infof("in store.Register after deleting askg s.R: %v\n", s.R)
						adu := repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.None, PostAction: repo.Finish})

						//logger.L.Infof("in store.Register adu: %q\n", adu)
						return adu, fmt.Errorf("in store.Register dataPiece group with TS \"%s\" and BeginPart \"%d\" is finished", m3f.B.TS, m3f.B.BeginPart)

					}
					// dataPiece in real is not last
					s.BufferAdd(d)

					return repo.AppDistributorUnit{}, fmt.Errorf("in store.Register dataPiece with TS \"%s\" and Part \"%d\" added to buffer", d.TS(), d.Part())
				}
			}
			// askd not met

			s.BufferAdd(d)
			/*
				logger.L.Infof("in store.Register dataPiece with body %q added to buffer and later became\n", d.GetBody(0))

				for _, v := range s.B {
					for j, w := range v {
						logger.L.Infof("in store.Register in buffer j = %d, header: %v, body: %q\n", j, w.GetHeader(), w.GetBody(0))
					}
				}
			*/
			for i := range s.R[askg] {
				//logger.L.Warnf("in store.Register ranging over s.R => askd = %v\n", i)

				askdParts = append(askdParts, i.SK.Part)

			}
			sort.Ints(askdParts)
			//logger.L.Warnf("in store.Register askdParts: %d\n", askdParts)
			if len(askdParts) == 1 {
				return du, fmt.Errorf("in store.Register dataPiece's Part for given TS \"%s\" should be \"%d\" but got \"%d\"", d.TS(), askdParts[0], d.Part())
			}
			if len(askdParts) > 1 {
				return du, fmt.Errorf("in store.Register dataPiece's Part for given TS \"%s\" should be one of \"%d\" but got \"%d\"", d.TS(), askdParts, d.Part())
			}

		}
		// askg not met

		s.BufferAdd(d)
		/*
			logger.L.Infof("in store.Register dataPiece with body %q added to buffer and later became\n", d.GetBody(0))

			for _, v := range s.B {
				for j, w := range v {
					logger.L.Infof("in store.Register j = %d, header: %v, body: %q\n", j, w.GetHeader(), w.GetBody(0))
				}
			}
		*/
		return du, fmt.Errorf("in store.Register dataPiece's TS %q is unknown", askg.TS)

	default: // DataPiece doesn't need beginning
		//logger.L.Infof("in store.Register d header %v, body %q\n", d.GetHeader(), d.GetBody(0))

		if !d.IsSub() { // DataPiece is not a subPiece

			if d.E() == repo.False {

				return repo.NewAppDistributorUnitUnary(d, bou, repo.Message{}), nil
			}
			if d.E() == repo.Last {
				//logger.L.Infof("in store.Register d.E()==repo.Last, s.C = %v\n", s.C)

				if s.C[askg] != 1 {
					s.BufferAdd(d)
					//logger.L.Errorf("in store.Register dataPiece with body %q d.E()==repo.Last, s.C = %v, adding to buffer %v\n", d.GetBody(0), s.C, s.B)
					return repo.AppDistributorUnit{}, fmt.Errorf("in store.Register dataPiece with TS \"%s\" and part \"%d\" added to buffer", d.TS(), d.Part())
				}

				delete(s.R, askg)
				adu := repo.NewAppDistributorUnitUnary(d, bou, repo.Message{PostAction: repo.Finish})
				//logger.L.Infof("in store.Register adu %v\n", adu)
				return adu, nil
			}

			dispo, err = repo.NewDispositionFilled(d, bou)

			asv := repo.AppStoreValue{B: b, D: dispo, E: d.E()}

			//logger.L.Infof("in store.Register dataPiece made store value with B: %v, and Dispo H: %q, Dispo Fo: %q, Dispo Fi: %q\n", asv.B, asv.D.H, asv.D.FormName, asv.D.FileName)
			//logger.L.Infof("in store.Register askd: %v\n", askd)

			v := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{}
			vv := map[bool]repo.AppStoreValue{}
			/*
				if d.E() == repo.True { //detailed part increment
					askd.Part++
				}
			*/
			if m1, ok := s.R[askg]; ok {

				if m2, ok := m1[askd]; ok {

					if m3t, ok := m2[true]; ok { //has true branch => incrementing askd
						vv[false] = asv
						vv[true] = m3t
						v[askd.IncPart()] = vv
						delete(s.R[askg], askd)
						s.R[askg] = v
						//logger.L.Infof("in store.Register after adding second dataPiece with header %v, askg s.R: %v\n", d.GetHeader(), s.R)
						return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil
					}
					//no true branch => incrementing or not depends on d.E()
					switch d.E() {
					case repo.True:
						vv[false] = asv
						//v[askd.IncPart()] = vv

						//delete(s.R[askg], askd)
						s.R[askg][askd.IncPart()] = vv

						//logger.L.Infof("in store.Register after adding dataPiece header %v, body %q, s.R: %v\n", d.GetHeader(), d.GetBody(0), s.R)

						return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil

					case repo.False: // unary
						delete(s.R[askg], askd)
						return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.Stop}), fmt.Errorf("in store.Register dataPiece group with TS \"%s\" and BeginPart = %d is finished", askg.TS, asv.B.BeginPart)

					case repo.Probably:
						vv[false] = asv
						v[askd] = vv
						//logger.L.Infof("in store.Register after adding dataPiece header %v, body %q, asv: %v\n", d.GetHeader(), d.GetBody(0), asv)
						delete(s.R[askg], askd)
						s.R[askg] = v
						//logger.L.Infof("in store.Register after adding dataPiece header %v, body %q, S.R: %v\n", d.GetHeader(), d.GetBody(0), s.R)
						return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), fmt.Errorf("in store.Register got double-meaning dataPiece")
					}
				}

				switch d.E() {

				case repo.True:

					vv[false] = asv

					s.R[askg][askd.IncPart()] = vv

					//logger.L.Infof("in store.Register askd = %q, s.R: %v\n", askd, s.R)

					if err != nil {

						if strings.Contains(err.Error(), "is not full") {
							errCore := fmt.Sprint(err.Error()[len("in repo.NewDispositionFilled header \""):strings.Index(err.Error(), "\" is not full")])
							return du, fmt.Errorf("in store.Register header from dataPiece's body \"%s\" is not full", errCore)
						}
					}
					//logger.L.Warnf("in store.Register just before returning s.R = %v, err %v\n", s.R, err)
					return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil

				case repo.Probably:

					vv[false] = asv

					if m3t, ok := m1[askd.T()][true]; ok {
						delete(s.R[askg], askd.T())
						vv[true] = m3t
						s.R[askg][askd.F().IncPart()] = vv
						//logger.L.Warnf("in store.Register s.R became %v\n", s.R)
						return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil
					}
					//test this case

				}

				s.R[askg][askd] = vv

				if err != nil {

					if strings.Contains(err.Error(), "is not full") {
						errCore := fmt.Sprint(err.Error()[len("in repo.NewDispositionFilled header \""):strings.Index(err.Error(), "\" is not full")])
						return du, fmt.Errorf("in store.Register header from dataPiece's body \"%s\" is not full", errCore)
					}
				}

				return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), fmt.Errorf("in store.Register got double-meaning dataPiece")

			}
			// ASKG not met
			vv[false] = asv
			//logger.L.Infof("in store.Register vv: %v\n", vv)

			switch d.E() {
			case repo.True:

				v[askd.IncPart()] = vv
				s.R[askg] = v
				//logger.L.Infof("in store.Register after adding dataPiece with body %q, s.R: %v\n", d.GetBody(0), s.R)

			case repo.Probably:
				v[askd] = vv
				s.R[askg] = v
			}
			if err != nil {
				if strings.Contains(err.Error(), "is not full") {
					errCore := fmt.Sprint(err.Error()[len("in repo.NewDispositionFilled header \""):strings.Index(err.Error(), "\" is not full")])
					return du, fmt.Errorf("in store.Register header from dataPiece's body \"%s\" is not full", errCore)
				}
			}

			return repo.NewDistributorUnitStream(asv, d, repo.Message{PreAction: repo.None, PostAction: repo.None}), nil

		}
		// Datapiece is subPiece => creating level 3 true branch
		vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{}
		vvv := map[bool]repo.AppStoreValue{}
		dispo, _ = repo.NewDispositionFilled(d, bou)
		m3t := repo.AppStoreValue{B: b, D: dispo, E: d.E()}

		if m1, ok := s.R[askg]; ok {
			if m2f, ok := m1[askd.F()]; ok {
				//logger.L.Infof("in store.Register s.R: %v has false branch askd = %v\n", s.R, askd.F())

				if m2f[false].E == repo.Probably { //adding true branch for askd
					//logger.L.Infof("in store.Register m2f[false]: %v\n", m2f[false])
					vvv[false] = m2f[false]
					vvv[true] = m3t
					vv[askd.IncPart().F()] = vvv
					delete(s.R[askg], askd.F())
					s.R[askg][askd.IncPart().F()] = vvv
					//logger.L.Warnf("in store.Register s.R: %v\n", s.R)
					return du, fmt.Errorf("in store.Register got double-meaning dataPiece")
				}
				// m2f[false].E == repo.True => adding separate ASKD
				vvv[true] = m3t
				s.R[askg][askd.T()] = vvv
				//logger.L.Infof("in store.Register s.R: %v\n", s.R)
				return du, fmt.Errorf("in store.Register got double-meaning dataPiece")
			}
			// no ASKD met => creating ASKD
			vvv[true] = m3t
			s.R[askg][askd.T()] = vvv
			return du, fmt.Errorf("in store.Register got double-meaning dataPiece")
			//logger.L.Infof("in store.Register s.R: %v, askd = %v\n", s.R, askd)
		}

		vvv[true] = m3t
		vv[askd] = vvv

		s.R[askg] = vv
		//logger.L.Infof("in store.Register after adding dataPiece TS: %s, Part = %d, E() = %d, body %q, s.R became %v\n", d.TS(), d.Part(), d.E(), d.GetBody(0), s.R)

		return du, fmt.Errorf("in store.Register got double-meaning dataPiece")
	}
}

func (s *StoreStruct) BufferAdd(d repo.DataPiece) {

	askg := repo.NewAppStoreKeyGeneral(d)
	if s.B == nil {
		s.B = map[repo.AppStoreKeyGeneral][]repo.DataPiece{}
	}

	switch lenb := len(s.B[askg]); {
	case lenb == 0:
		s.B[askg] = make([]repo.DataPiece, 0)
		s.B[askg] = append(s.B[askg], d)

	case lenb == 1:

		e := s.B[askg][0]

		if Equal(d, e) { // avoiding dublicates
			return
		}

		if d.Part() >= e.Part() {
			//logger.L.Infof("in store.BufferAdd in case 1 trying to append dataPiece with body %q to buffer %v\n", d.GetBody(0), s.B[askg])
			s.B[askg] = append(s.B[askg], d)
		}
		if d.Part() < e.Part() {
			s.Prepend(s.B[askg], d)
		}

	default:

		for _, v := range s.B[askg] { // avoiding dublicates
			if Equal(d, v) {
				return
			}
		}

		f := s.B[askg][0]
		l := s.B[askg][len(s.B[askg])-1]

		if f.Part()+1 != l.Part() { //if first and last elements are not neighbours

			if d.Part() > l.Part() { //new element's Part is more than last element's
				s.B[askg] = append(s.B[askg], d)
			}
			if d.Part() < f.Part() { //new element's Part is less than first element's
				s.Prepend(s.B[askg], d)
			}
			if d.Part() == l.Part()-1 {

				lastIndex := len(s.B[askg]) - 1
				s.B[askg] = append(s.B[askg], d)

				Swap(s.B[askg], lastIndex, lastIndex+1)

			}
			if d.Part() == f.Part()+1 {

				s.Prepend(s.B[askg], d)

				Swap(s.B[askg], 0, 1)

			}
			if d.Part() < l.Part()-1 && //new element's Part is less than last element's
				d.Part() > f.Part()+1 { //new element's Part is more than first element's

				s.B[askg] = append(s.B[askg], d)
				b := s.B[askg]
				sort.SliceStable(s.B[askg], func(i int, j int) bool { return b[i].Part() < b[j].Part() })
			}

		}
		if f.Part()+1 == l.Part() { //if first and last element are neighbours
			if d.Part() > l.Part() { //new element's Part is more than last element's
				s.B[askg] = append(s.B[askg], d)
			}
			if d.Part() < f.Part() { //new element's Part is less than first element's
				s.Prepend(s.B[askg], d)
			}
		}

	}
	/*
		logger.L.Warnln("in store.BufferAdd buffer after all:")
		for i, v := range s.B[askg] {
			logger.L.Infof("in store.BufferAdd i = %d, header: %q, body: %q\n", i, v.GetHeader(), v.GetBody(0))
		}
	*/
}

// Inserting new element to the beginning of slice, shifting the rest
func (s *StoreStruct) Prepend(tsb []repo.DataPiece, d repo.DataPiece) {

	askg := repo.NewAppStoreKeyGeneral(d)

	if b, ok := s.B[askg]; ok {
		s.B[askg] = append([]repo.DataPiece{}, d)
		s.B[askg] = append(s.B[askg], b...)
		//logger.L.Infof("in store.Prepend len became %d\n", len(s.B[d.TS()]))
		//logger.L.Infof("in store.Prepend s.B[d.TS()][1] = %v\n", s.B[d.TS()][1])
		return
	}
	s.B[askg] = append([]repo.DataPiece{}, d) // something wrong case
}

// Swapping two slice elements
func Swap(s []repo.DataPiece, i, j int) {
	e := s[i]
	s[i] = s[j]
	s[j] = e
}

func (s *StoreStruct) CleanBuffer(ids repo.AppStoreBufferIDs) {
	//logger.L.Infof("store.CleanBuffer invoked with ids: %v\n", ids)
	marked := 0
	if len(ids.I) == 0 {
		return
	}

	askg := ids.ASKG
	// setting registerd dataPieces as empty
	for i := range ids.I {

		//logger.L.Infof("in store.CleanBuffer i = %d, element: %v, buffer element: %v\n", i, ids.I[i], s.B[askg])

		if s.B[askg][ids.I[i]].E() == repo.Last && s.Counter(askg) == 0 {

			delete(s.B, askg)
			return
		}
		s.B[askg][ids.I[i]] = &repo.AppPieceUnit{}
		marked++
	}
	if marked == len(s.B[askg]) {
		delete(s.B, askg)
		return
	}
	/*
		logger.L.Infof("store.CleanBuffer s.B after nulling:\n")
		for i := range s.B {
			for j, w := range s.B[i] {
				logger.L.Infof("store.CleanBuffer id = %d, value: %v\n", j, w)
			}
		}
	*/
	// sorting B putting empty ones to buffer beginning

	sort.SliceStable(s.B[askg], func(a, b int) bool {
		return s.B[askg][a].TS() < s.B[askg][b].TS()
	})
	/*
		logger.L.Infof("store.CleanBuffer s.B after sorting:\n")
		for i := range s.B {
			for j, w := range s.B[i] {
				logger.L.Infof("store.CleanBuffer j = %d, w:%v\n", j, w)
			}
		}
	*/
	//cutting off the buffer beginning
	for j := range s.B[askg] {
		//logger.L.Infof("store.CleanBuffer j = %d, s.B[ids.TS][j].TS(): %v\n", j, s.B[ids.TS][j].TS())
		if s.B[askg][j].TS() != "" {
			s.B[askg] = s.B[askg][j:]
			break
		}
		if j == len(s.B[askg])-1 {
			s.B[askg] = []repo.DataPiece{}
		}
	}
	/*
		logger.L.Infof("store.CleanBuffer s.B after deleting:\n")
		for i := range s.B {
			for j, w := range s.B[i] {
				logger.L.Infof("store.CleanBuffer j = %d, w:%v\n", j, w)
			}
		}
	*/
}

// Register buffer elements, considering they are sorted by Part
func (s *StoreStruct) RegisterBuffer(bou repo.Boundary) ([]repo.AppDistributorUnit, []error) {
	//logger.L.Infof("store.RegisterBuffer invoked with R: %v, B: %v and counter = %v\n", s.R, s.B, s.C)
	ids, adus, errs, repeat := repo.NewAppStoreBufferIDs(), make([]repo.AppDistributorUnit, 0), make([]error, 0), true

	if len(s.B) == 0 {
		//logger.L.Errorln("in store.RegisterBuffer buffer is empty, leaving")
		errs = append(errs, fmt.Errorf("in store.RegisterBuffer buffer has no elements"))
		return adus, errs
	}

	for i, v := range s.B {

		for repeat {

			repeat = false

			for j, w := range v {

				askg, askd := repo.NewAppStoreKeyGeneral(w), repo.NewAppStoreKeyDetailed(w)
				//logger.L.Infof("in store.RegisterBuffer j = %d w = %v, counter %d, ids %d\n", j, w, s.Counter(askg), ids.GetIDs(askg))

				if isIn(j, ids.GetIDs(askg)) { // index'v been registered already skipping

					if len(ids.GetIDs(askg)) == len(s.B[askg]) {

						break
					}

					continue
				}
				if !w.B() && w.E() == repo.Last && s.Counter(askg) == 1 {
					//logger.L.Infof("in store.RegisterBuffer dataPiece with header %v, body %q is the last, should dec counter & clean buffer\n", w.GetHeader(), w.GetBody(0))

					adu, err := s.Register(w, bou)
					//logger.L.Infof("in store.RegisterBuffer last adu header: %v, body: %q error: %v\n", adu.H, adu.B.B, err)

					if err != nil {
						errs = append(errs, err)
					}

					if !cmp.Equal(adu, repo.AppDistributorUnit{}) ||
						(cmp.Equal(adu, repo.AppDistributorUnit{}) &&
							strings.Contains(err.Error(), "double-meaning")) {
						adus = append(adus, adu)
						s.Dec(askg)
						//logger.L.Infof("in store.RegisterBuffer counter decremented and became %d\n", s.Counter(askg))
						repeat = true
						ids.Add(repo.NewAppStoreBufferID(i, j))
						continue
					}
				}
				// not last elements
				if m1, ok := s.R[askg]; ok {
					if _, ok := m1[askd]; ok {

						//logger.L.Infof("in store.RegisterBuffer checking dataPiece with header %v, body %q, counter = %v\n", w.GetHeader(), w.GetBody(0), s.C)

						if w.E() == repo.Last && s.Counter(i) > 1 {
							continue
						}
						if w.E() == repo.Last {
							//logger.L.Infof("in store.RegisterBuffer Last element dataPiece with header %v, body %q and counter = 1 true counter = %v\n", w.GetHeader(), w.GetBody(0), s.Counter(i))
							adu, err := s.Register(w, bou)
							//logger.L.Infof("in store.RegisterBuffer adu header: %v, body: %q error: %v\n", adu.H, adu.B.B, err)

							if err != nil {
								errs = append(errs, err)
							}

							if !cmp.Equal(adu, repo.AppDistributorUnit{}) {
								adus = append(adus, adu)
								s.Dec(askg)
								//logger.L.Infof("in store.RegisterBuffer counter decremented and became %d\n", s.Counter(askg))
								repeat = true
								ids.Add(repo.NewAppStoreBufferID(i, j))
								continue
							}
						}

						//logger.L.Infof("in store.RegisterBuffer dataPiece with header %v, body %q matched to R\n", w.GetHeader(), w.GetBody(0))
						adu, err := s.Register(w, bou)
						//logger.L.Infof("in store.RegisterBuffer adu header: %v, body: %q error: %v\n", adu.H, adu.B.B, err)

						if err != nil {
							errs = append(errs, err)
						}

						if !cmp.Equal(adu, repo.AppDistributorUnit{}) {
							adus = append(adus, adu)
							s.Dec(askg)
							//logger.L.Infof("in store.RegisterBuffer counter decremented and became %d\n", s.Counter(askg))
							repeat = true
							ids.Add(repo.NewAppStoreBufferID(i, j))
							continue
						}
					}
				}
				repeat = false
			}
		}
	}
	s.CleanBuffer(*ids)

	return adus, errs
}

func (s *StoreStruct) Inc(askg repo.AppStoreKeyGeneral, n int) {
	s.L.Lock()
	defer s.L.Unlock()
	s.C[askg] += n
}

func (s *StoreStruct) Dec(askg repo.AppStoreKeyGeneral) {
	s.L.Lock()
	defer s.L.Unlock()
	s.C[askg]--
}
func (s *StoreStruct) Counter(askg repo.AppStoreKeyGeneral) int {
	s.L.RLock()
	defer s.L.RUnlock()
	return s.C[askg]
}

func Equal(a, b repo.DataPiece) bool {
	if a.TS() == b.TS() &&
		a.Part() == b.Part() &&
		a.B() == b.B() &&
		a.E() == b.E() &&
		bytes.Contains(a.GetBody(0), b.GetBody(0)) {
		return true
	}
	return false
}

func isIn(i int, d []int) bool {
	for _, v := range d {
		if i == v {
			return true
		}
	}
	return false
}
