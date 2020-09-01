package pool

import (
	"errors"
	"fmt"
	"net"
)

// A Constructor handles the logic to create new homogenous
// network connection.
type Constructor func() (net.Conn, error)

// Enumeration of errors
var (
	ErrBadNilConstructor = errors.New("constructor passed cannot be a nil function")
	ErrPoolConnClosed    = errors.New("connection pool (queue) is already closed/has nil connsQueue. rejected from the queue")
	ErrPoolClosingFail   = errors.New("fail in attempt to close the connection pool (queue)")
	ErrNilConnEnqueued   = errors.New("putting nil connection in the queue. rejected from the queue")
	ErrCreateConnFail    = errors.New("error in creating new connection in the constructor func")
)

// A ErrConnQueueClose happens during QueuePool.Close() if
// any of the connection in the queue failed to be closed.
// ErrConnQueueClose holds the errors list resulting in each
// of the failed conn.Close() operation.
type ErrConnQueueClose []error

// Len counts how many errors in the list
func (errConn *ErrConnQueueClose) Len() int { return len(*errConn) }

// Collect error and put it in the list
func (errConn *ErrConnQueueClose) Collect(err error) {
	if err == nil {
		return
	}
	*errConn = append(*errConn, err)
}

// Error messages for debugging
func (errConn *ErrConnQueueClose) Error() string {
	msg := "errors occured during QueuePool.Close() :\n"
	for _, err := range *errConn {
		if err != nil {
			msg += fmt.Sprintf("\t- error: %s\n", err.Error())
		}
	}
	return msg
}
