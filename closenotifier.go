package closenotifier

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

func makeCloseNotifier(r io.Reader) <-chan bool {
	closeNotifier := make(chan bool, 1)
	go func() {
		buf := make([]byte, 1)
		r.Read(buf)
		closeNotifier <- true
	}()
	return closeNotifier
}

func Hijack(w http.ResponseWriter) (buf *bufio.ReadWriter, closeNotifier <-chan bool, closer func(), err error) {
	var rwc net.Conn
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		err = errors.New("can't hijack")
		return
	}
	rwc, buf, err = hijacker.Hijack()
	if err != nil {
		return
	}
	closer = func() {
		buf.WriteString("0\r\n\r\n")
		buf.Flush()
		buf.Reader.Reset(nil)
		buf.Writer.Reset(nil)
		buf = nil
		rwc.Close()
		rwc = nil
	}
	closeNotifier = makeCloseNotifier(rwc)
	return
}
