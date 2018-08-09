package pgsrv

// ResultTag can be implemented by driver.Result to provide the tag name to be
// used to notify the postgres client of the completed command. If left
// unimplemented, the default behavior follows the spec described in the link
// below. For all un-documented cases, "UPDATE N" will be used, where N is the
// number of affected rows.
// See CommandComplete: https://www.postgresql.org/docs/10/static/protocol-message-formats.html
type ResultTag interface {
	Tag() (string, error)
}

// Session represents a connected client session.
type Session interface {
}

// Server is an interface for objects capable for handling the postgres protocol
// by serving client connections. Each connection is assigned a Session that's
// maintained in-memory until the connection is closed.
type Server interface {
	Listen(laddr string) error // blocks. Run in go-routine.
}
