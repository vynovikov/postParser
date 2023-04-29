// HTTP Reciever.
package tp

import (
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/vynovikov/postParser/internal/adapters/application"
	"github.com/vynovikov/postParser/internal/logger"
	"github.com/vynovikov/postParser/internal/repo"
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
		if err != nil && conn == nil && r.A.Stopping() {

			r.wg.Wait()
			r.A.ChainInClose()

			return

		}

		r.wg.Add(1)
		ts := repo.NewTS()

		go r.HandleRequest(conn, ts, &r.wg)

	}

}

// Tested in http_test.go
func (r *tpReceiverStruct) HandleRequest(conn net.Conn, ts string, wg *sync.WaitGroup) {
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
func (r *tpReceiverStruct) Stop(wg *sync.WaitGroup) {

	r.srv.l.Close()

	r.wg.Wait()

	wg.Done()
}
