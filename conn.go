package pool

import (
	"net"
	"sync"
)

// ReusableConn wraps any net.Conn such that when method Close()
// is called on this connection, it will be put back in the queue
// rather than being closed. It makes it reusable for future requests
type ReusableConn struct {
	net.Conn              // embed the underlying conn
	connPool *QueuePool   // conn pool to put it back when Close()
	mutex    sync.RWMutex // mutex to change useable state
	usable   bool         // whether we can reuse this conn
}

// SetUsable ..
func (rc *ReusableConn) SetUsable() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.usable = true
}

// SetUnusable ...
func (rc *ReusableConn) SetUnusable() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.usable = false
}

// Close overrides underlying net.Conn.Close() which
// instead put this connection back to the registered
// connection pool
func (rc *ReusableConn) Close() error {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	if !rc.usable || rc.Conn == nil {
		if rc.Conn != nil {
			return rc.Conn.Close()
		}
		return nil
	}
	return rc.connPool.put(rc.Conn)
}

// NewReusableConn creates new connection that is reusable
// when Close() is called
func NewReusableConn(conn net.Conn, connPool *QueuePool) *ReusableConn {
	return &ReusableConn{
		Conn:     conn,
		connPool: connPool,
	}
}
