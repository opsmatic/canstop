package main

import (
	"github.com/opsmatic/canstop"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/**
 * This is a re-implementation of @rcrowley's example using canstop
 * http://rcrowley.org/articles/golang-graceful-stop.html
 */

type Service struct {
	listener *net.TCPListener
}

func (self *Service) Run(l *canstop.Lifecycle) (err error) {
	for !l.IsInterrupted() {
		if conn, err, timeout := canstop.AcceptTCPWithTimeout(self.listener, 5*time.Second); timeout {
			continue
		} else if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		} else {
			session := &Session{conn}
			// running the session using Lifecycle guarantees that any transaction
			// already taking place on the connection will have a chance to complete
			go l.Session(session.Run)
		}
	}
	log.Printf("Orderly shutdown of listener %s\n", self.listener.Addr())
	return
}

type Session struct {
	conn *net.TCPConn
}

func (self *Session) Run(l *canstop.Lifecycle) (e error) {
	defer self.conn.Close()
	for !l.IsInterrupted() {
		buf := make([]byte, 4096)
		if _, err, timeout := canstop.ReadWithTimeout(self.conn, buf, 5*time.Second); timeout {
			continue
		} else if err != nil {
			log.Printf("Error reading: %s\n", err)
			return
		}
		if _, err := self.conn.Write(buf); err != nil {
			log.Printf("Error writing: %s\n", err)
			return
		}
	}
	log.Printf("Orderly shutdown of connection %s\n", self.conn.RemoteAddr())
	return
}

func main() {
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:48879")
	if nil != err {
		log.Fatalln(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if nil != err {
		log.Fatalln(err)
	}
	log.Println("listening on", listener.Addr())

	l := canstop.NewLifecycle()

	svc := &Service{listener}
	go l.Service(svc.Run, "echo listener")

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	l.StopAndWait()
}
