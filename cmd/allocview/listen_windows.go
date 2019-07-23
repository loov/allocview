package main

import (
	"fmt"
	"log"
	"net"

	npipe "gopkg.in/natefinch/npipe.v2"
)

type Handler func(conn net.Conn)

type Server struct {
	listener *npipe.PipeListener
	handler  Handler
}

func NewServer(name string, handler Handler) (*Server, error) {
	listener, err := npipe.Listen(`\\.\pipe\` + name)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: listener,
		handler:  handler,
	}, nil
}

func (server *Server) ListenAndServe() error {
	// TODO: pass in contexts
	for {
		conn, err := server.listener.Accept()
		fmt.Println("got connection", conn)
		if err != nil {
			log.Println(err)
			continue
		}

		go server.handler(conn)
	}
}

func (server *Server) String() string {
	return server.listener.Addr().String()
}

func (server *Server) Close() error {
	// TODO: wait for handlers to finish
	return server.listener.Close()
}
