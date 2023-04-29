// HTTPS receiver.
//
// x509 pair should be in tls forlder inside root directory
package tps

import (
	"crypto/tls"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/vynovikov/postParser/internal/adapters/application"
	"github.com/vynovikov/postParser/internal/logger"
	"github.com/vynovikov/postParser/internal/repo"
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

	cer, err := tls.LoadX509KeyPair("tls/cert.pem", "tls/key.pem")
	if err != nil {
		logger.L.Errorf("in tps.NewTpsReceiver tls.LoadX509KeyPair returned err: %v\n", err)
		return nil
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	li, err := tls.Listen("tcp", ":443", config)
	if err != nil {
		logger.L.Errorf("in driver.Run error: %v\n", err)
	}
	logger.L.Infoln("listening localhost:443")

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

// Tested in https_test.go
func (r *tpsReceiverStruct) HandleRequest(conn net.Conn, ts string, wg *sync.WaitGroup) {

	p := 0

	bou, header, errFirst := repo.AnalyzeHeader(conn)

	for {
		h := repo.NewReceiverHeader(ts, p, bou)
		b, errSecond := repo.AnalyzeBits(conn, 1024, p, header)

		u := repo.NewReceiverUnit(h, b)

		if errFirst != nil {
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

	r.srv.l.Close()

	r.wg.Wait()

	wg.Done()
}
