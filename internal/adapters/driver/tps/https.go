package tps

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"os"
	"postParser/internal/adapters/application"
	"postParser/internal/logger"
	"postParser/internal/repo"
	"strings"
)

type tpsReceiver struct {
	A application.Application
}

func NewTpsReceiver(a application.Application) *tpsReceiver {
	return &tpsReceiver{A: a}
}

func (r *tpsReceiver) Run() {
	cer, err := tls.LoadX509KeyPair("tls/cert.pem", "tls/key.pem")
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	li, err := tls.Listen("tcp", ":443", config)
	if err != nil {
		logger.L.Errorf("in driver.Run error: %v\n", err)
	}
	logger.L.Info("listening on :443")
	defer li.Close()

	for {
		conn, err := li.Accept()
		if err != nil {
			logger.L.Error(err)
		}
		ts := repo.NewTS()
		go r.HandleRequest(conn, ts)

	}

}

func (r *tpsReceiver) HandleRequest(conn net.Conn, ts string) {
	p := 0

	bou, header, errFirst := repo.AnalyzeHeader(conn)

	for {
		h := repo.NewReceiverHeader(ts, p, bou)
		b, errSecond := repo.AnalyzeBits(conn, 1024, p, header)

		u := repo.NewReceiverUnit(h, b)

		if errFirst != nil {
			//logger.L.Errorf("in http.Handle errFirst: %v, breaking\n", errFirst)
			if errFirst == io.EOF || errFirst == io.ErrUnexpectedEOF || os.IsTimeout(errFirst) {
				r.A.AddToFeeder(u)
				break
			}
		}
		if errSecond != nil {
			if errSecond == io.EOF || errSecond == io.ErrUnexpectedEOF || os.IsTimeout(errSecond) {
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
}
