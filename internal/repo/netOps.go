package repo

import (
	"fmt"
	"io"
	"net"
	"os"
	"postParser/internal/logger"
	"time"
)

func AnalyzeHeader(conn net.Conn) (Boundary, []byte, error) {
	header := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 15)) // tls handshake requires at least 9 ms timeout
	n0, err := io.ReadFull(conn, header)
	if err != nil &&
		(err != io.EOF && err != io.ErrUnexpectedEOF && !os.IsTimeout(err)) {
		return Boundary{}, header, err
	}
	if n0 < len(header) {
		header = header[:n0]
	}
	logger.L.Infof("in repo.AnalyzeHeader header: %q\n", header)
	bou := FindBoundary(header)
	return bou, header, err
}

func AnalyzeBits(conn net.Conn, i, p int, h []byte) (ReceiverBody, error) {
	rb, ending := NewReceiverBody(i), make([]byte, 0)
	if p == 0 {

		lenh := len(h)
		if lenh < 512 {
			ending = make([]byte, 1024-lenh)
		} else {
			ending = make([]byte, 512)
		}

		rb.B = h

		if lenh < 512 {
			return rb, io.EOF
		}
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1))
		n, err := io.ReadFull(conn, ending)

		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF && !os.IsTimeout(err) {
				return rb, err
			}
			// EOF
			if n > 0 && n <= len(ending) {

				ending = ending[:n]
				rb.B = append(rb.B, ending...)

				return rb, err
			}

		}
		if n > 0 && n < len(ending) {
			ending = ending[:n]
		}
		rb.B = append(rb.B, ending...)

		return rb, nil
	}
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1))
	n, err := io.ReadFull(conn, rb.B)
	if err != nil {

		if err != io.EOF && err != io.ErrUnexpectedEOF && !os.IsTimeout(err) {
			return rb, err
		}
		// EOF
		if n == 0 {

			return NewReceiverBody(0), fmt.Errorf("in repo.AnalyzeBits request part %d is empty", p)
		}
		if n > 0 && n <= len(rb.B) {

			rb.B = rb.B[:n]

			return rb, err
		}

	}

	return rb, nil

}

func Respond(conn net.Conn) {

	body := "200 OK"

	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: text/html\r\n\r\n%s", len(body), body)
}
