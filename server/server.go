package server

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"server/pkg/handler"
	"time"
)

type Server struct {
	listener *net.TCPListener
}

func (s *Server) bind(network string, addr string) error {
	var err error
	lAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return err
	}
	s.listener, err = net.ListenTCP(network, lAddr)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Run(network string, addr string, hand *handler.Handler, ctx context.Context) error {
	if err := s.bind(network, addr); err != nil {
		return err
	}
	log.Printf("The server was started at %s\n", addr)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := s.listener.SetDeadline(time.Now().Add(time.Second)); err != nil {
				return err
			}

			conn, err := s.listener.Accept()
			if err != nil {
				if os.IsTimeout(err) {
					continue
				}
				return err
			}
			log.Printf("accepted %s", conn.RemoteAddr())

			go func(conn net.Conn) {
				hand.Handle(bufio.NewReader(conn), bufio.NewWriter(conn))
				if err := conn.Close(); err != nil {
					log.Printf("%s", err.Error())
				}
			}(conn)
		}
	}
}

func (s *Server) Shutdown(cancel context.CancelFunc) error{
	cancel()
	log.Println("Server Shutting down")
	return s.listener.Close()
}
