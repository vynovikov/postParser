package tp

import (
	"io"
	"net"
	"os"
	"postParser/internal/adapters/application"
	"postParser/internal/logger"
	"postParser/internal/repo"
	"strings"
)

type tpReceiver struct {
	A *application.App
}

func NewTpReceiver(a *application.App) *tpReceiver {
	return &tpReceiver{A: a}
}

func (r *tpReceiver) Run() {
	li, err := net.Listen("tcp", ":3000")
	if err != nil {
		logger.L.Error(err)
	}
	logger.L.Info("listening on :3000")
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

func (r *tpReceiver) HandleRequest(conn net.Conn, ts string) {
	p := 0

	bou, header, errFirst := repo.AnalyzeHeader(conn)

	for {
		h := repo.NewReceiverHeader(ts, p, bou)
		b, errSecond := repo.AnalyzeBits(conn, 1024, p, header)

		u := repo.NewReceiverUnit(h, b)

		if errFirst != nil {

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
