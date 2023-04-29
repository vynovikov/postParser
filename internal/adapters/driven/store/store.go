// Store adapter.
// Stores information neccessary for dataPiece handling
package store

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/vynovikov/postParser/internal/repo"
)

type StoreStruct struct {
	R map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue
	B map[repo.AppStoreKeyGeneral][]repo.DataPiece
	C map[repo.AppStoreKeyGeneral]repo.Counter
	L sync.RWMutex
}

func NewStore() *StoreStruct {
	return &StoreStruct{
		R: make(map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue),
		B: make(map[repo.AppStoreKeyGeneral][]repo.DataPiece),
		C: make(map[repo.AppStoreKeyGeneral]repo.Counter),
		L: sync.RWMutex{},
	}
}

type Store interface {
	BufferAdd(repo.DataPiece)
	Register(repo.DataPiece, repo.Boundary) (repo.AppDistributorUnit, error)
	RegisterBuffer(repo.AppStoreKeyGeneral, repo.Boundary) ([]repo.AppDistributorUnit, []error)
	Inc(repo.AppStoreKeyGeneral, int)
	Dec(repo.DataPiece) (repo.Order, error)
	Presence(repo.DataPiece) (repo.Presense, error)
	Act(repo.DataPiece, repo.StoreChange)
	Unblock(repo.AppStoreKeyGeneral)
}

// Register updates store.R with data from particular dataPiece.
// Tested in store_test.go
func (s *StoreStruct) Register(d repo.DataPiece, bou repo.Boundary) (repo.AppDistributorUnit, error) {

	var (
		du        repo.AppDistributorUnit
		askdParts []int
	)
	if s.B == nil { // delete after testing
		s.B = make(map[repo.AppStoreKeyGeneral][]repo.DataPiece)
	}

	askg := repo.NewAppStoreKeyGeneralFromDataPiece(d)
	askd := repo.NewAppStoreKeyDetailed(d)

	switch {

	case d.B() == repo.True: //DataPiece needs beginning

		if m1, ok := s.R[askg]; ok { // General key met

			if m2, ok := m1[askd]; ok { // Detailed key met
				m3f := m2[false]

				if m3f.D.FormName == "" { // Disposition is not filled
					header, err := d.H(bou)

					if m3t, ok := m2[true]; ok { // true branch is present
						if strings.Contains(err.Error(), "is ending part") {

							m3f.D.H = append(m3f.D.H, m3t.D.H...)

							m3f.D.H = append(m3f.D.H, header...)
							m3f.D.FormName, m3f.D.FileName = repo.GetFoFi(m3f.D.H)
							m3f.E = d.E()

							d.BodyCut(len(header))

							vvv := map[bool]repo.AppStoreValue{}
							vvv[false] = m3f
							delete(s.R[askg], askd)
							s.R[askg][askd.IncPart()] = vvv

							if m3f.B.Part == 0 {
								return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Start, PostAction: repo.Continue}), nil
							}
							return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}), fmt.Errorf("in store.Register new dataPiece group with TS \"%s\" and Part = %d", askg.TS, m3f.B.Part)
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
						delete(s.R[askg], askd)

						s.R[askg][askd.IncPart()] = vvv
						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}), fmt.Errorf("in store.Register new dataPiece group with TS \"%s\" and Part = %d", askg.TS, m3f.B.Part)
					}
				}
				// Disposition is filled
				if m3t, ok := m2[true]; ok { // true branch is present

					if m3f.E == repo.True { // false branch needs next part -- still waiting for dataPiece with d.E == repo.Probably which is current one

						m3f.E = d.E()

						vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{askd.IncPart(): {false: m3f, true: m3t}}
						delete(s.R[askg], askd)
						s.R[askg] = vv

						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}), nil

					}

					header, err := d.H(bou)

					if err != nil {
						if strings.Contains(err.Error(), "is ending part") { // header in dataPiece's body present

							if repo.IsLastBoundary(m3t.D.H, header, bou) {
								adu := repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.ClientStream, S: repo.StreamData{SK: repo.StreamKey{TS: d.TS(), Part: m3t.B.Part}, M: repo.Message{S: "finish"}}}}
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

								adu := repo.AppDistributorUnit{}

								switch d.E() {
								case repo.True:
									adu = repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.StopLast, PostAction: repo.Continue})
								case repo.False:

									adu = repo.NewAppDistributorUnitUnaryComposed(m3f, d, repo.Message{PreAction: repo.StopLast, PostAction: repo.Close})
								}

								return adu, fmt.Errorf("in store.Register new dataPiece group with TS \"%s\" and Part = %d", askg.TS, m3f.B.Part)
							}

							d.Prepend(m3t.D.H)

							m3f.E = d.E()

							vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{askd.IncPart(): {false: m3f}}

							delete(s.R[askg], askd)

							s.R[askg] = vv

							return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}), nil

						}

						if strings.Contains(err.Error(), "no header found") {

							m3f.E = d.E()

							vv := map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue{askd.IncPart(): {false: m3f}}
							delete(s.R[askg], askd)
							s.R[askg] = vv

							d.Prepend(m3t.D.H)

							return repo.NewDistributorUnitStream(s.R[askg][askd.IncPart()][false], d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}), nil

						}
						if strings.Contains(err.Error(), "is the last") {

							delete(s.R, askg)

							adu := repo.NewDistributorUnitStreamEmpty(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Finish})

							return adu, fmt.Errorf("in store.Register dataPiece group with TS \"%s\" is finished", d.TS())
						}
					}
					if repo.IsLastBoundary(m3t.D.H, header, bou) {
						return repo.AppDistributorUnit{H: repo.AppDistributorHeader{T: repo.Tech, C: repo.CloseData{TS: d.TS()}}}, nil
					}
				}
				// no true branch
				vvv := map[bool]repo.AppStoreValue{}
				switch d.E() {

				case repo.Probably: // deteiled group can be closed or finished

					m2f, ok1 := m1[askd.F()]
					m2t, ok2 := m1[askd.T()]
					if ok1 && ok2 { // there is another ASKD with same part which has true branch only
						m3t := m2t[true]

						m3f := m2f[false]
						m3f.E = d.E()

						vvv := map[bool]repo.AppStoreValue{}
						vvv[true] = m3t
						vvv[false] = m3f

						delete(s.R[askg], askd.T())
						delete(s.R[askg], askd.F())
						s.R[askg][askd.IncPart()] = vvv
						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}), nil
					}
					// no true branch

					m3f.E = d.E()

					delete(s.R[askg], askd)
					s.R[askg][askd] = map[bool]repo.AppStoreValue{false: m3f}

					return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue}), fmt.Errorf("in store.Register got double-meaning dataPiece")

				case repo.True: // detailed group shall be continued

					vvv[false] = m3f
					delete(s.R[askg], askd)

					s.R[askg][askd.IncPart()] = vvv

					adu := repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Continue})

					return adu, nil

				case repo.False:

					if m3t, ok := m2[true]; ok {

						boundaryTrimmed := repo.GenBoundary(bou)[2:]
						lb := append(m3t.D.H, d.GetBody(repo.Min(d.LL(), repo.MaxHeaderLimit))...)

						if repo.LastBoundary(lb, boundaryTrimmed) {

							delete(s.R, askg)

							return repo.NewDistributorUnitStreamEmpty(m3t, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Close}), fmt.Errorf("in store.Register dataPiece group with TS \"%s\" is finished", d.TS())
						}

						delete(s.R[askg], askd)

						return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Close}), fmt.Errorf("in store.Register dataPiece group with TS %q and Part = %d is finished", d.TS(), m3f.B.Part)
					}

					delete(s.R[askg], askd)

					return repo.NewDistributorUnitStream(m3f, d, repo.Message{PreAction: repo.Continue, PostAction: repo.Close}), fmt.Errorf("in store.Register dataPiece group with TS %q and Part \"%d\" is finished", d.TS(), m3f.B.Part)
				}
			}
			// askd not met

			s.BufferAdd(d)

			for i := range s.R[askg] {

				askdParts = append(askdParts, i.SK.Part)

			}
			sort.Ints(askdParts)

			if len(askdParts) == 1 {
				return du, fmt.Errorf("in store.Register dataPiece's Part for given TS \"%s\" should be \"%d\" but got \"%d\"", d.TS(), askdParts[0], d.Part())
			}
			if len(askdParts) > 1 {
				return du, fmt.Errorf("in store.Register dataPiece's Part for given TS \"%s\" should be one of \"%d\" but got \"%d\"", d.TS(), askdParts, d.Part())
			}

		}
		// askg not met

		s.BufferAdd(d)
		return du, fmt.Errorf("in store.Register dataPiece's TS %q is unknown", askg.TS)
	}
	return du, nil
}

// BufferAdd adds dataPiece to store.B.
// Store.B keeps being sorted after each addition.
// Tested in store_test.go
func (s *StoreStruct) BufferAdd(d repo.DataPiece) {

	askg := repo.NewAppStoreKeyGeneralFromDataPiece(d)
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
}

// Prepend is a helper function for BufferAdd.
// Inserts new element to the beginning of slice, shifting the rest.
func (s *StoreStruct) Prepend(tsb []repo.DataPiece, d repo.DataPiece) {

	askg := repo.NewAppStoreKeyGeneralFromDataPiece(d)

	if b, ok := s.B[askg]; ok {
		s.B[askg] = append([]repo.DataPiece{}, d)
		s.B[askg] = append(s.B[askg], b...)
		return
	}
	s.B[askg] = append([]repo.DataPiece{}, d) // something wrong case
}

// Swap ia a helper function for BufferAdd.
// Swaps two slice elements.
func Swap(s []repo.DataPiece, i, j int) {
	e := s[i]
	s[i] = s[j]
	s[j] = e
}

// Tested in store_test.go
func (s *StoreStruct) CleanBuffer(ids repo.AppStoreBufferIDs) {
	s.L.Lock()
	defer s.L.Unlock()

	marked := 0
	if len(ids.I) == 0 {
		return
	}

	askg := ids.ASKG
	// setting registerd dataPieces as empty
	for i := range ids.I {

		s.B[askg][ids.I[i]] = &repo.AppPieceUnit{}
		marked++
	}
	if marked == len(s.B[askg]) {
		delete(s.B, askg)
		return
	}
	// sorting B putting empty ones to buffer beginning

	sort.SliceStable(s.B[askg], func(a, b int) bool {
		return s.B[askg][a].TS() < s.B[askg][b].TS()
	})
	//cutting off the buffer beginning
	for j := range s.B[askg] {

		if s.B[askg][j].TS() != "" {
			s.B[askg] = s.B[askg][j:]
			break
		}
		if j == len(s.B[askg])-1 {
			s.B[askg] = []repo.DataPiece{}
		}
	}
}

// RegisterBuffer ranges overt store.B elements.
// Registers them until first fail
// Tested in store_test.go
func (s *StoreStruct) RegisterBuffer(askg repo.AppStoreKeyGeneral, bou repo.Boundary) ([]repo.AppDistributorUnit, []error) {
	ids, adus, errs, repeat, _ := repo.NewAppStoreBufferIDs(), make([]repo.AppDistributorUnit, 0), make([]error, 0), true, len(s.B[askg]) > 0 && !s.C[askg].Blocked

	if len(s.B) == 0 {
		errs = append(errs, fmt.Errorf("in store.RegisterBuffer buffer has no elements"))
		return adus, errs
	}

	if len(s.B[askg]) == 1 {
		if s.B[askg][0].B() == repo.False && s.B[askg][0].E() == repo.False {
			if s.C[askg].Cur == 1 && !s.C[askg].Blocked {
				switch s.C[askg].Max {
				case 1:
					adus = append(adus, repo.NewAppDistributorUnitUnary(s.B[askg][0], bou, repo.Message{PreAction: repo.Start, PostAction: repo.Finish}))
				default:
					adus = append(adus, repo.NewAppDistributorUnitUnary(s.B[askg][0], bou, repo.Message{PreAction: repo.None, PostAction: repo.Finish}))
				}
				return adus, nil

			}
		}
	}
	if len(s.B[askg]) == 1 && s.C[askg].Cur == 1 {
		if s.C[askg].Blocked {
			errs = append(errs, fmt.Errorf("in store.RegisterBuffer buffer has single element and current counter == 1 and blocked"))
			return adus, errs
		}

	}

	for i, v := range s.B {

		if i.TS != askg.TS {
			continue
		}

		for repeat {
			repeat = false

			for j, w := range v {

				askg, askd := repo.NewAppStoreKeyGeneralFromDataPiece(w), repo.NewAppStoreKeyDetailed(w)

				if isIn(j, ids.GetIDs(askg)) { // index'v been registered already, skipping

					if len(ids.GetIDs(askg)) == len(s.B[askg]) {
						break
					}
					continue
				}

				if m1, ok := s.R[askg]; ok {
					if _, ok := m1[askd]; ok {
						if !s.C[askg].Blocked {
							adu, err := s.Register(w, bou)

							if err != nil {
								errs = append(errs, err)
							}
							if s.C[askg].Cur == 1 {
								if adu.H.T == repo.Unary {
									adu.H.U.M.PostAction = repo.Finish
								}
								if adu.H.T == repo.ClientStream {
									adu.H.S.M.PostAction = repo.Finish
								}
							}
							s.Dec(w)

							adus = append(adus, adu)
							ids.Add(repo.NewAppStoreBufferID(i, j))

							continue
						}
						if s.C[askg].Blocked && s.C[askg].Cur > 1 {
							adu, err := s.Register(w, bou)
							if err != nil {
								errs = append(errs, err)
							}

							s.Dec(w)
							adus = append(adus, adu)
							repeat = true
							ids.Add(repo.NewAppStoreBufferID(i, j))
							continue

						}
						if s.C[askg].Blocked && s.C[askg].Cur == 1 {
							errs = append(errs, fmt.Errorf("in store.RegisterBuffer buffer has single element and current counter == 1 and blocked"))
							repeat = false
							break
						}

					}
				}
				repeat = false
			}
		}
	}
	if len(ids.I) > 0 && len(s.B[askg]) > 0 {
		s.CleanBuffer(*ids)
	}

	return adus, errs
}

// Increments Store.Counter by n
func (s *StoreStruct) Inc(askg repo.AppStoreKeyGeneral, n int) {
	s.L.Lock()
	defer s.L.Unlock()
	c := repo.NewCounter()

	if cm, ok := s.C[askg]; ok {
		cm.Max += n
		cm.Cur += n
		s.C[askg] = cm
		return
	}
	c.Max = n
	c.Cur = n
	c.Blocked = true
	s.C[askg] = c

}

// Decrements Store.Counter by one if possible and returns result and error
// Tested in store_test.go
func (s *StoreStruct) Dec(d repo.DataPiece) (repo.Order, error) {
	askg := repo.NewAppStoreKeyGeneralFromDataPiece(d)
	s.L.Lock()
	defer s.L.Unlock()

	if cm, ok := s.C[askg]; ok {
		started := cm.Started

		if (cm.Blocked && cm.Cur > 1) ||
			(cm.Blocked && cm.Cur == 1 && d.B() == repo.False && d.E() != repo.False) ||

			!cm.Blocked {
			if d.IsSub() {
				cm.Cur--
				cm.Max--
			} else {
				cm.Cur--

				if !cm.Started {
					cm.Started = true
				}
			}

			delete(s.C, askg)
			s.C[askg] = cm
			if cm.Cur == 0 && cm.Max-cm.Cur == 1 {

				if !started && cm.Started {

					if d.B() == repo.False && d.E() == repo.False {
						s.Reset(askg)
						return repo.FirstAndLast, nil
					}

					return repo.First, nil
				}
				if d.B() == repo.True {
					return repo.Intermediate, nil
				}

			}
			if !started && cm.Started { // may be different d.Part
				return repo.First, nil
			}
			if cm.Cur == cm.Max {
				return repo.Unordered, nil
			}
			if cm.Cur == 0 {

				if s.C[askg].Blocked {
					return repo.Intermediate, nil
				}
				// !Blocked
				s.Reset(askg)
				return repo.Last, nil
			}
			return repo.Intermediate, nil
		}
		return repo.Unordered, fmt.Errorf("in store.Dec cannot dec further")
	}
	return repo.Unordered, fmt.Errorf("in store.Dec askg \"%v\" not found", askg)
}

// Reset deletes record in Store maps. Maps are recreated if record is last.
func (s *StoreStruct) Reset(askg repo.AppStoreKeyGeneral) {
	if _, ok := s.R[askg]; ok {
		if len(s.R) < 2 {
			s.R = make(map[repo.AppStoreKeyGeneral]map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue)
		}
		delete(s.R, askg)
	}

	if _, ok := s.B[askg]; ok {
		if len(s.B) < 2 {
			s.B = make(map[repo.AppStoreKeyGeneral][]repo.DataPiece)
		}
		delete(s.B, askg)
	}

	if _, ok := s.C[askg]; ok {
		if len(s.C) < 2 {

			s.C = make(map[repo.AppStoreKeyGeneral]repo.Counter)
		}
		delete(s.C, askg)
	}
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

// Presense determines how dataPiece is documented in store.R
// Tested in store_test.go
func (s *StoreStruct) Presence(d repo.DataPiece) (repo.Presense, error) {
	askg, askd, vv := repo.NewAppStoreKeyGeneralFromDataPiece(d), repo.NewAppStoreKeyDetailed(d), make(map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue)
	if m1, ok := s.R[askg]; ok {
		if m2, ok := m1[askd]; ok && d.B() == repo.True {
			if s.C[askg].Cur == 1 && s.C[askg].Blocked {
				return repo.Presense{}, fmt.Errorf("in store.Presense matched but Cur == 1 && Blocked")
			}
			vv[askd.F()] = m2
			if m2t, ok := m1[askd.T()]; ok && d.E() == repo.Probably {
				vv[askd.T()] = m2t
				return repo.NewPresense(true, true, true, vv), nil
			}
			return repo.NewPresense(true, true, false, vv), nil
		}
		if d.IsSub() {
			if m2f, ok := s.R[askg][askd.F()]; ok && m2f[false].E == repo.Probably {
				vv[askd.F()] = m2f
				return repo.NewPresense(true, true, true, vv), nil
			}
			return repo.NewPresense(true, false, false, nil), nil
		}
		if d.B() == repo.False && d.E() == repo.Probably {
			if m2t, ok := s.R[askg][askd.T()]; ok && m2t[true].E == repo.Probably {
				vv[askd.T()] = m2t
				return repo.NewPresense(true, true, true, vv), nil
			}
			return repo.NewPresense(true, true, false, nil), nil
		}
		return repo.NewPresense(true, false, false, nil), nil
	}
	if d.B() == repo.False && d.E() == repo.Probably {

	}
	return repo.Presense{}, nil
}

// Act updates store.R
// Tested in store_test.go
func (s *StoreStruct) Act(d repo.DataPiece, sc repo.StoreChange) {
	askg, vv := repo.NewAppStoreKeyGeneralFromDataPiece(d), make(map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue)

	switch sc.A {
	case repo.Change:
		if m1, ok := s.R[askg]; ok { // ASKD met

			for i, _ := range sc.From {
				if _, ok := m1[i]; ok {
					delete(s.R[askg], i)
				}

			}
			for i, v := range sc.To {
				s.R[askg][i] = v
				return
			}
		}
		// no ASKG
		if d.B() == repo.False {
			for i, v := range sc.To {
				vv[i] = v
				s.R[askg] = vv
				return
			}
		}
		s.R[askg] = make(map[repo.AppStoreKeyDetailed]map[bool]repo.AppStoreValue)

	case repo.Buffer:
		s.BufferAdd(d)
	case repo.Remove:
		delete(s.R, askg)
	}
}

func (s *StoreStruct) Unblock(askg repo.AppStoreKeyGeneral) {
	s.L.Lock()
	defer s.L.Unlock()

	mr := s.C[askg]
	mr.Blocked = s.L.TryLock()

	delete(s.C, askg)

	s.C[askg] = mr
}
