package tps

import (
	"crypto/tls"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"workspaces/postParser/internal/adapters/application"
	"workspaces/postParser/internal/logger"
	"workspaces/postParser/internal/repo"
)

type TpsServer struct {
	l net.Listener
}

type TpsReceiver interface {
	Run()
	HandleRequest(conn net.Conn, ts string, wg *sync.WaitGroup)
	Stop(*sync.WaitGroup)
}

type tpsReceiverStruct struct {
	A   application.Application
	srv *TpsServer
	wg  sync.WaitGroup
}

func NewTpsReceiver(a application.Application) *tpsReceiverStruct {

	cer, err := tls.LoadX509KeyPair("../../tls/cert.pem", "../../tls/key.pem")
	if err != nil {
		logger.L.Errorf("in tp.NewTpsReceiver tls.LoadX509KeyPair returned err: %v\n", err)
		return nil
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	li, err := tls.Listen("tcp", ":443", config)
	if err != nil {
		logger.L.Errorf("in driver.Run error: %v\n", err)
	}
	logger.L.Info("listening on :443")

	srv := &TpsServer{l: li}

	return &tpsReceiverStruct{
		A:   a,
		srv: srv,
	}
}

func (r *tpsReceiverStruct) Run() {

	for {
		conn, err := r.srv.l.Accept()
		if err != nil {

			if r.A.Stopping() {
				return
			}
			logger.L.Error(err)
		}
		r.wg.Add(1)

		ts := repo.NewTS()
		go r.HandleRequest(conn, ts, &r.wg)

	}

}

func (r *tpsReceiverStruct) HandleRequest(conn net.Conn, ts string, wg *sync.WaitGroup) {

	p := 0

	bou, header, errFirst := repo.AnalyzeHeader(conn)

	for {
		h := repo.NewReceiverHeader(ts, p, bou)
		b, errSecond := repo.AnalyzeBits(conn, 1024, p, header)

		u := repo.NewReceiverUnit(h, b)

		if errFirst != nil {
			//logger.L.Errorf("in http.Handle errFirst: %v, breaking\n", errFirst)
			if errFirst == io.EOF || errFirst == io.ErrUnexpectedEOF || os.IsTimeout(errFirst) {
				u.H.Unblock = true
				r.A.AddToFeeder(u)
				break
			}
		}
		if errSecond != nil {
			if errSecond == io.EOF || errSecond == io.ErrUnexpectedEOF || os.IsTimeout(errSecond) {
				u.H.Unblock = true
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
		r.A.ChainInClose()
	}
}

func (r *tpsReceiverStruct) Stop(wg *sync.WaitGroup) {

	//logger.L.Errorln("in tps.Stop closing tps listener")

	r.srv.l.Close()

	//logger.L.Errorln("in tps.Stop waiting")

	r.wg.Wait()

	//logger.L.Errorln("in tps.Stop all stopped")

	wg.Done()
}
