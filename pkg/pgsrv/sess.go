package pgsrv

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"

	"github.com/buglloc/simplelog"
)

// Session represents a single client-connection, and handles all of the
// communications with that client.
//
// see: https://www.postgresql.org/docs/9.2/static/protocol.html
// for postgres protocol and startup handshake process
type session struct {
	Server      *server
	Conn        net.Conn
	initialized bool
	log         log.Logger
}

// Handle a connection session
func (s *session) Serve() error {
	// read the initial connection startup message
	startupMsg, err := s.Read()
	if err != nil {
		return err
	}

	if startupMsg.IsCancel() {
		return errors.New("client disconnected")
	}

	if startupMsg.IsTLSRequest() {
		// currently we don't support TLS.
		err := s.Write(tlsResponseMsg(false))
		if err != nil {
			return err
		}

		// re-read the full startup message
		startupMsg, err = s.Read()
		if err != nil {
			return err
		}
	}
	s.initialized = true
	startupArgs, _ := startupMsg.StartupArgs()

	err = s.Write(authReqMsg())
	if err != nil {
		return err
	}

	authMsg, err := s.Read()
	if err == nil && len(startupArgs) > 0 {
		password, pErr := authMsg.AuthPassword()
		if pErr == nil {
			s.LogAuth(startupArgs, password)
		}
	}

	err = s.Write(authOKMsg())
	if err != nil {
		return err
	}

	err = s.Write(keyDataMsg(rand.Int31(), rand.Int31()))
	if err != nil {
		return err
	}

	if _, ok := startupArgs["application_name"]; !ok {
		return errors.New("client doesn't send application_name")
	}

	// send unknown app name
	err = s.Write(errMsg(AppnameUnknown()))
	if err != nil {
		return err
	}
	return nil
}

// Read reads and returns a single message from the connection.
//
// The Postgres protocol supports two types of messages: (1) untyped messages
// are only mostly present during the initial startup process and starts with
// the length of the message, followed by the content. (2) typed messages are
// similar to the untyped messages except that they're prefixed with a
// single-byte type character used to distinguish between the different message
// types (query, prepare, etc.), and are the normal messages used for most
// client-server communications.
//
// This method abstracts away this differentiation, returning the next available
// message whether it's typed or not.
func (s *session) Read() (msg, error) {
	typeChar := make([]byte, 1)
	if s.initialized {

		// we've already started up, so all future messages are MUST start with
		// a single-byte type identifier.
		_, err := s.Conn.Read(typeChar)
		if err != nil {
			return nil, err
		}
	}

	// read the actual body of the message
	msg, err := s.readBody()
	if err != nil {
		return nil, err
	}

	if typeChar[0] != 0 {

		// we have a typed-message, prepend it to the message body by first
		// creating a new message that's 1-byte longer than the body in order to
		// make room in memory for the type byte
		body := msg
		msg = make([]byte, len(body)+1)

		// fixing the type byte at the beginning (position 0) of the new message
		msg[0] = typeChar[0]

		// finally append the body to the new message, starting from position 1
		copy(msg[1:], body)
	}

	return newMsg(msg), nil
}

// ReadMsgBody reads the body of the next message in the connection. The body is
// comprised of an Int32 body-length (N), inclusive of the length itself
// followed by N-bytes of the actual body.
func (s *session) readBody() ([]byte, error) {

	// messages starts with an Int32 Length of message contents in bytes,
	// including self.
	lenBytes := make([]byte, 4)
	_, err := io.ReadFull(s.Conn, lenBytes)
	if err != nil {
		return nil, err
	}

	// convert the 4-bytes to int
	length := int(binary.BigEndian.Uint32(lenBytes))

	// read the remaining bytes in the message
	msg := make([]byte, length)
	_, err = io.ReadFull(s.Conn, msg[4:]) // keep 4 bytes for the length
	if err != nil {
		return nil, err
	}

	// append the message content to the length bytes in order to rebuild the
	// original message in its entirety
	copy(msg[:4], lenBytes)
	return msg, nil
}

func (s *session) Write(m msg) error {
	_, err := s.Conn.Write(m)
	return err
}

func (s session) LogAuth(startupArgs map[string]interface{}, password string) {
	var user, database string
	if u, ok := startupArgs["user"]; ok {
		user = u.(string)
	}
	if d, ok := startupArgs["database"]; ok {
		database = d.(string)
	}

	if user == "" && password == "" {
		return
	}

	s.log.Info("credentials from client", "user", user, "password", password, "database", database)
}
