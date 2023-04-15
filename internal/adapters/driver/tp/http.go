package tp

import (
	"io"
	"net"
	"os"
	"postParser/internal/adapters/application"
	"postParser/internal/logger"
	"postParser/internal/repo"
	"strings"
	"sync"
)

type TpServer struct {
	l net.Listener
}

type TpReceiver interface {
	Run()
	HandleRequest(net.Conn, string, *sync.WaitGroup)
	Stop(*sync.WaitGroup)
}

type tpReceiverStruct struct {
	A   application.Application
	srv *TpServer
	wg  sync.WaitGroup
}

func NewTpReceiver(a application.Application) *tpReceiverStruct {

	li, err := net.Listen("tcp", ":3000")
	if err != nil {
		logger.L.Error(err)
	}
	logger.L.Info("listening localhost:3000")

	s := &TpServer{l: li}

	return &tpReceiverStruct{
		A:   a,
		srv: s,
	}
}

func (r *tpReceiverStruct) Run() {
	for {
		conn, err := r.srv.l.Accept()
		//logger.L.Infof("in tp.Run conn: %v, err: %v\n", conn, err)
		if err != nil && conn == nil && r.A.Stopping() {

			//logger.L.Errorln("in tp.Run closing chanIn")
			r.wg.Wait()
			r.A.ChainInClose()

			return

		}
		r.wg.Add(1)
		ts := repo.NewTS()
		//logger.L.Infof("in tp.Run current time %v, ts: %q\n", time.Now().Format("02.01.2006 15_16_17"), ts)
		go r.HandleRequest(conn, ts, &r.wg)

	}

}

func (r *tpReceiverStruct) HandleRequest(conn net.Conn, ts string, wg *sync.WaitGroup) {
	p := 0

	bou, header, errFirst := repo.AnalyzeHeader(conn)

	for {
		h := repo.NewReceiverHeader(ts, p, bou)
		b, errSecond := repo.AnalyzeBits(conn, 1024, p, header)

		u := repo.NewReceiverUnit(h, b)
		if p == 4 || p == 5 {
			logger.L.Infof("in tp.HandleRequest unit header %v, body: %q, errFirst %v, errSecond %v\n", u.H, u.B.B, errFirst, errSecond)
		}
		if errFirst != nil {

			if errFirst == io.EOF || errFirst == io.ErrUnexpectedEOF || os.IsTimeout(errFirst) {
				u.H.Unblock = true
				//logger.L.Errorf("in tp.HandleRequest errFirst case u header: %v, error: %v\n", u.H, errFirst)
				r.A.AddToFeeder(u)
				break
			}
		}
		if errSecond != nil {
			if errSecond == io.EOF || errSecond == io.ErrUnexpectedEOF || os.IsTimeout(errSecond) {
				u.H.Unblock = true
				//logger.L.Errorf("in tp.HandleRequest errSecond case u header: %v, error: %v\n", u.H, errSecond)
				r.A.AddToFeeder(u)
				break
			}
			if strings.Contains(errSecond.Error(), "empty") {
				break
			}
		}

		r.A.AddToFeeder(u)

		p++
	}

	repo.Respond(conn)
	wg.Done()
	if r.A.Stopping() {
		//logger.L.Errorln("in tp.HandleRequest closing chainIn")
		r.A.ChainInClose()
	}
}

func (r *tpReceiverStruct) Stop(wg *sync.WaitGroup) {

	//logger.L.Errorln("in tp.Stop closing tp listener")

	r.srv.l.Close()

	//logger.L.Errorln("in tp.Stop waiting")

	r.wg.Wait()

	//logger.L.Errorln("in tp.Stop all stopped")

	//r.A.ChainInClose()

	wg.Done()
}
