package pool

import (
	"net"
	"sync"
)

// Default constants for initializing new QueuePool
const (
	DefaultInitConn = 30
	DefaultMaxConn  = 100
)

// A QueuePool contains a queue of pool, in which we can enqueue
// and dequeue network connection into the pool (FIFO). The QueuePool
// allows network connections to be reusable.
type QueuePool struct {
	mutex       sync.RWMutex
	connsQueue  chan net.Conn
	constructor Constructor
}

// NewQueuePool creates new QueuePool and initialise the
// queue with specified number of initial connection using
// the contructor function that is passed.
func NewQueuePool(initConn, maxConns int, constructor Constructor) (*QueuePool, error) {
	if initConn < 0 {
		initConn = DefaultInitConn
	}
	if maxConns < 1 {
		maxConns = DefaultMaxConn
	}
	if constructor == nil {
		return nil, ErrBadNilConstructor
	}
	
	connsQueue := make(chan net.Conn, maxConns)
	queuePool := &QueuePool{
		mutex:       sync.RWMutex{},
		connsQueue:  connsQueue,
		constructor: constructor,
	}
	
	var errorLists = make(ErrConnQueueClose, 0, initConn)
	for i := 0; i < initConn; i++ {
		if conn, err := constructor(); err != nil {
			errorLists.Collect(err)
		} else {
			queuePool.connsQueue <- conn
		}
	}
	if errorLists.Len() > 0 {
		return queuePool, &errorLists
	}
	return queuePool, nil
}

// Get dequeues a connection from the pool (queue)
func (qp *QueuePool) Get() (net.Conn, error) {
	qp.mutex.RLock()
	defer qp.mutex.RUnlock()

	if qp.connsQueue == nil {
		return nil, ErrPoolConnClosed
	}
	select {
	case conn := <-qp.connsQueue:
		if conn == nil {
			return nil, ErrPoolConnClosed
		}
		return conn, nil
	default:
		conn, err := qp.constructor()
		if err != nil {
			return nil, ErrCreateConnFail
		}
		return conn, nil
	}
}

// Put enqueues a connection from the pool (queue)
func (qp *QueuePool) Put(conn net.Conn) error {
	if conn == nil {
		return ErrNilConnEnqueued
	}

	qp.mutex.RLock()
	defer qp.mutex.RUnlock()

	if qp.connsQueue == nil {
		return conn.Close()
	}
	select {
	case qp.connsQueue <- conn:
		// enqueue connection into the pool
		return nil
	default:
		// close connection when queue is full
		return conn.Close()
	}
}

// Close all network connection in the pool (queue)
func (qp *QueuePool) Close() error {
	qp.mutex.Lock()
	defer qp.mutex.Unlock()

	var errorLists = make(ErrConnQueueClose, 0, len(qp.connsQueue))

	close(qp.connsQueue)
	for conn := range qp.connsQueue {
		if err := conn.Close(); err != nil {
			errorLists.Collect(err)
		}
	}
	qp.connsQueue = nil
	if errorLists.Len() > 0 {
		return &errorLists
	}
	return nil
}

// Len counts how many connection in the pool (queue)
func (qp *QueuePool) Len() int {
	qp.mutex.RLock()
	defer qp.mutex.RUnlock()
	return len(qp.connsQueue)
}
