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

func (self *Service) Run(l *canstop.Lifecycle) (e error) {
	for !l.IsInterrupted() {
		self.listener.SetDeadline(time.Now().Add(5 * time.Second))
		conn, err := self.listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}
		log.Println(conn)

		session := &Session{conn}
		// calling Run causes the session.Run method to be executed in a goroutine
		// with the Lifecycle instance as an argument (providing access to the Interrupt channel).
		// This also guarantees that any transaction already taking place on the connection
		// will have a chance to complete
		l.ManageSession(session.Run)
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
		self.conn.SetDeadline(time.Now().Add(5 * time.Second))
		buf := make([]byte, 4096)
		if _, err := self.conn.Read(buf); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
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
	go l.ManageService(svc.Run, "echo listener")

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	l.StopAndWait()
}
