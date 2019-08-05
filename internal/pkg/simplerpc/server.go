package simplerpc

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type ServerHandlerFunc = func(c net.Conn, args []string)

type Server struct {
	handlers map[string]ServerHandlerFunc
	port     int
	socket   net.Listener
}

func NewServer(port int) *Server {
	return &Server{
		handlers: make(map[string]ServerHandlerFunc),
		port:     port,
	}
}

func (s *Server) Start() {
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

			go s.handleConnection(fd)
		}
	}()
}

func (s *Server) AddHandler(command string, handler ServerHandlerFunc) {
	s.handlers[command] = handler
}

func (s *Server) handleConnection(c net.Conn) {
	b := bufio.NewReader(c)

	for {
		line, err := b.ReadBytes('\n')
		if err != nil { // EOF, or worse
			break
		}
		// split command into parts, remove last character (new line)
		parts := strings.Split(string(line[0:len(line)-1]), " ")

		if len(parts) == 0 {
			//noinspection ALL
			fmt.Fprintln(c, "ERR", "Command required", parts[0])
			//noinspection ALL
			c.Close()
		}
		command := parts[0]

		if !s.hasHandler(command) {
			//noinspection ALL
			fmt.Fprintln(c, "ERR", "Unknown command", parts[0])
			//noinspection ALL
			c.Close()
			continue
		}

		handler := s.handler(command)
		handler(c, parts[1:])
		//noinspection ALL
		c.Close()
	}
}

func (s *Server) hasHandler(command string) bool {
	_, ok := s.handlers[command]

	return ok
}

func (s *Server) handler(command string) ServerHandlerFunc {
	if s.hasHandler(command) {
		return s.handlers[command]
	}

	return nil
}

func (s *Server) Stop() error {
	log.Println("Stopping socket server")

	return s.socket.Close()
}
