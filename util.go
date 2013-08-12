package canstop

import (
	"net"
	"time"
)

type ConnWrapper struct {
	*net.TCPConn
}

func ReadWithTimeout(c net.Conn, buf []byte, maxWait time.Duration) (read int, err error, timeout bool) {
	c.SetDeadline(time.Now().Add(maxWait))
	if read, err = c.Read(buf); err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return 0, nil, true
		}
	}
	return read, err, false
}

func AcceptTCPWithTimeout(l *net.TCPListener, maxWait time.Duration) (c *net.TCPConn, err error, timeout bool) {
	l.SetDeadline(time.Now().Add(5 * time.Second))
	c, err = l.AcceptTCP()
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return nil, nil, true
		}
	}
	return c, err, false

}
