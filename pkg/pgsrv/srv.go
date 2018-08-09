package pgsrv

import (
	"net"

	"github.com/buglloc/simplelog"
)

type server struct {
	listener net.Listener
}

func New() Server {
	return &server{}
}

func (s *server) Listen(laddr string) (err error) {
	s.listener, err = net.Listen("tcp", laddr)
	if err != nil {
		return err
	}

	log.Info("server started", "addr", laddr)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.serve(conn)
	}
	return
}

func (s *server) serve(conn net.Conn) error {
	defer conn.Close()

	logger := log.Child("client_addr", conn.RemoteAddr().String())
	sess := &session{
		Server: s,
		Conn:   conn,
		log:    logger,
	}

	err := sess.Serve()
	if err != nil {
		logger.Info("failed to process session", "err", err.Error())
	} else {
		logger.Info("session successfully processed, time to stop listening")
		s.listener.Close()
	}

	return err
}
