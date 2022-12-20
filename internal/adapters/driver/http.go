package driver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"postParser/internal/adapters/application"
	"postParser/internal/logger"
	"postParser/internal/repo"
)

type Receiver struct {
	a application.Application
}

func NewReceiver(a application.Application) *Receiver {
	return &Receiver{a: a}
}

func (r *Receiver) Run() {
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
		go r.Handle(conn)

	}

}

func (r *Receiver) Handle(conn net.Conn) {

	r.Respond(conn)

	bufReader := bufio.NewReader(conn)
	header, err := bufReader.Peek(512)
	if err != nil {
		logger.L.Info(err)
	}
	ts := repo.NewTS()

	h := repo.NewReceiverHeader(ts, header)

	s := repo.NewReceiverSignal("NEW")

	ru := repo.NewReceiverUnit(h, repo.ReceverBody{}, s)

	//sending to start recording
	r.a.AddToFeeder(ru)

	for {
		b := repo.NewReceiverBody(1024)

		n, err := bufReader.Read(b.B)
		if n < len(b.B) {
			b.B = b.B[:n]
		}
		if err != nil {
			if err == io.EOF {

				//logger.L.Info("reading EOF sending EOF")
				//logger.L.Infof("ReceiverUnit gonna be %v\n", ru)

				s := repo.NewReceiverSignal("EOF")

				ru.SetSignal(s)

				r.a.AddToFeeder(ru)

				break
			}
			logger.L.Errorf("in driver.Handle error: %v\n", err)
			break
		}

		//logger.L.Infof("Receiver.Handle made header: %v\n", h)
		logger.L.Infof("Receiver.Handle has body:\n%q\n", string(b.B))

		r.a.AddToFeeder(repo.NewReceiverUnit(h, b, repo.ReceiverSignal{}))

		repo.IncPart(&h)
	}

}

func (r *Receiver) Respond(conn net.Conn) {

	logger.L.Info("Responding to conn")

	body := `<!DOCTYPE html><html lang ="en"><head><meta
	charet="UTF-8"><title></title></head><body><strong>Hello World</strong></body></html>`

	fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	fmt.Fprint(conn, "\r\n")
	fmt.Fprint(conn, body)
}
