package server

import (
	"net"
	"log"
	"fmt"
)

type SocketServerHandlerFunc = func(c net.Conn)

type SocketServer struct {
	handler SocketServerHandlerFunc
	port    int
	socket  net.Listener
}

func NewSocketServer(port int, h SocketServerHandlerFunc) *SocketServer {
	return &SocketServer{
		handler:    h,
		port: port,
	}
}

func (s *SocketServer) Start() {
	socket, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))

	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	s.socket = socket

	go func() {
		for {
			fd, err := socket.Accept()
			if err != nil {
				log.Println("Accept error: ", err)
			}

			go s.handler(fd)
		}
	}()
}

func (s *SocketServer) Stop() {
	log.Println("Stopping socket server")
	s.socket.Close()
}
