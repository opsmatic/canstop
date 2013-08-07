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

func (self *Service) Run(c *canstop.Control) {
	for !c.IsPoisoned() {
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
		// with the Control instance as an argument (providing access to the Poison channel).
		// This also guarantees that any transaction already taking place on the connection
		// will have a chance to complete
		c.Run(session.Run)
	}
}

type Session struct {
	conn *net.TCPConn
}

func (self *Session) Run(c *canstop.Control) {
	defer self.conn.Close()
	for !c.IsPoisoned() {
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

	r := canstop.NewManager(5 * time.Second)

	svc := &Service{listener}
	r.Manage(svc.Run, "echo listener")

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	r.Stop()
}
