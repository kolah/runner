package server

import (
	"net"
	"log"
	"syscall"
)

type SocketServerHandlerFunc = func(c net.Conn)

type SocketServer struct {
	handler     SocketServerHandlerFunc
	socketPath  string
	socket      net.Listener
}

func NewSocketServer(path string, h SocketServerHandlerFunc) *SocketServer {
	return &SocketServer{
		handler:     h,
		socketPath:  path,
	}
}

func (s *SocketServer) Start() {
	syscall.Unlink(s.socketPath)
	socket, err := net.Listen("unix", s.socketPath)

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
